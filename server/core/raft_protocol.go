// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"fmt"
	"io"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/golang/protobuf/proto"
	"github.com/lni/dragonboat/v3"
	"github.com/lni/dragonboat/v3/config"
	"github.com/lni/dragonboat/v3/raftio"
	"github.com/lni/dragonboat/v3/statemachine"

	"github.com/pole-group/pole/server/cluster"
	"github.com/pole-group/pole/server/consistency"
	"github.com/pole-group/pole/server/sys"
	"github.com/pole-group/pole/utils"
)

type RaftProtocol struct {
	lock      sync.RWMutex
	cfg       consistency.RaftConfig
	servers   map[string]*raftServer
	serverMgn *cluster.ServerClusterManager
}

// 初始化操作
func (protocol *RaftProtocol) Init(cfg consistency.ProtocolConfig, memberMgn *cluster.ServerClusterManager) {

}

// 接收读请求
func (protocol *RaftProtocol) Read(req *consistency.Read) {

}

// 接收写请求
func (protocol *RaftProtocol) Write(req *consistency.Write) {

}

// 添加请求处理者
func (protocol *RaftProtocol) AddProcessor(processors ...consistency.RequestProcessor) error {
	for _, processor := range processors {
		switch t := processor.(type) {
		case consistency.RequestProcessor4CP:
			server, err := createAndStartRaftServer(protocol, t)
			if err != nil {
				return err
			}
			name, _ := processor.Group()
			protocol.servers[name] = server
		default:
			return fmt.Errorf("processors must be type of RequestProcessor4CP")
		}
	}
	return nil
}

// 节点变更操作
func (protocol *RaftProtocol) OnMemberChange(newMembers []*cluster.Member) {

}

// Raft协议关闭
func (protocol *RaftProtocol) Shutdown() error {
	defer protocol.lock.RUnlock()
	protocol.lock.RLock()
	for _, server := range protocol.servers {
		if err := server.Close(); err != nil {
			return err
		}
	}
	return nil
}

type raftServer struct {
	protocol   *RaftProtocol
	raftDir    string
	innerRaft  *dragonboat.NodeHost
	processor  consistency.RequestProcessor4CP
	leaderInfo atomic.Value
}

func createAndStartRaftServer(protocol *RaftProtocol, p consistency.RequestProcessor4CP) (*raftServer, error) {
	server := &raftServer{
		protocol:   protocol,
		processor:  p,
		leaderInfo: atomic.Value{},
	}
	if err := server.init(); err != nil {
		return nil, err
	}
	if err := server.start(); err != nil {
		return nil, err
	}
	return server, nil
}

func (server *raftServer) init() error {
	name, _ := server.processor.Group()
	dataDir := filepath.Join(sys.GetEnvHolder().DataPath, "protocol", "raft", name)

	nhc := config.NodeHostConfig{
		WALDir:         dataDir,
		NodeHostDir:    dataDir,
		RTTMillisecond: 200,
		RaftAddress: fmt.Sprintf("%s:%d", server.protocol.serverMgn.GetSelf().GetIp(),
			server.protocol.serverMgn.GetSelf().GetExtensionPort(cluster.CPPort)),
		RaftEventListener: server,
	}

	nh, err := dragonboat.NewNodeHost(nhc)
	if err != nil {
		return err
	}

	server.innerRaft = nh

	return nil
}

func (server *raftServer) start() error {
	rc := config.Config{
		NodeID:             server.protocol.serverMgn.GetSelf().MemberID,
		ElectionRTT:        5,
		HeartbeatRTT:       1,
		CheckQuorum:        true,
		SnapshotEntries:    10,
		CompactionOverhead: 5,
	}

	name, id := server.processor.Group()
	dataDir := filepath.Join(sys.GetEnvHolder().DataPath, "protocol", "raft", name)

	isFirstStart := utils.Exists(dataDir)

	initialMembers := make(map[uint64]string)

	for _, member := range server.protocol.serverMgn.GetMemberList() {
		initialMembers[member.MemberID] = member.GetAddr()
	}

	rc.ClusterID = id
	if err := server.innerRaft.StartCluster(initialMembers, isFirstStart, createSM(server), rc); err != nil {
		return fmt.Errorf("failed to add cluster, %v\n", err)
	}

	return nil
}

func (server *raftServer) LeaderUpdated(info raftio.LeaderInfo) {
	server.leaderInfo.Store(info)
}

func (server *raftServer) nodeChange(newMember []*cluster.Member) {
}

func (server *raftServer) Close() error {
	server.innerRaft.Stop()
	return nil
}

func createSM(server *raftServer) func(uint64, uint64) statemachine.IStateMachine {
	return func(clusterID uint64, nodeID uint64) statemachine.IStateMachine {
		return &wrapperRaftStateMachine{p: server.processor, closed: 0}
	}
}

type wrapperRaftStateMachine struct {
	clusterID uint64
	nodeID    uint64
	p         consistency.RequestProcessor4CP
	closed    int32
}

var emptyStateMachineResult = statemachine.Result{}

func (wrs *wrapperRaftStateMachine) Update(date []byte) (statemachine.Result, error) {
	if err := wrs.isClosed(); err != nil {
		return emptyStateMachineResult, err
	}

	req := &consistency.Write{}
	if err := proto.Unmarshal(date, req); err != nil {
		return emptyStateMachineResult, err
	}
	resp := wrs.p.OnWrite(req)
	data, err := proto.Marshal(&resp)
	if err != nil {
		return emptyStateMachineResult, err
	}
	return statemachine.Result{Data: data}, nil
}

func (wrs *wrapperRaftStateMachine) Lookup(query interface{}) (interface{}, error) {
	if err := wrs.isClosed(); err != nil {
		return nil, err
	}
	req := &consistency.Read{}
	if err := proto.Unmarshal(query.([]byte), req); err != nil {
		return nil, err
	}
	resp := wrs.p.OnRead(req)
	data, err := proto.Marshal(&resp)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (wrs *wrapperRaftStateMachine) SaveSnapshot(io.Writer, statemachine.ISnapshotFileCollection,
	<-chan struct{}) error {
	if err := wrs.isClosed(); err != nil {
		return err
	}
	return nil
}

func (wrs *wrapperRaftStateMachine) RecoverFromSnapshot(io.Reader, []statemachine.SnapshotFile,
	<-chan struct{}) error {
	if err := wrs.isClosed(); err != nil {
		return err
	}
	return nil
}

func (wrs *wrapperRaftStateMachine) Close() error {
	atomic.StoreInt32(&wrs.closed, 1)
	return nil
}

func (wrs *wrapperRaftStateMachine) isClosed() error {
	if atomic.LoadInt32(&wrs.closed) == 1 {
		name, id := wrs.p.Group()
		return fmt.Errorf("[%s-%d] state-machine already closed", name, id)
	}
	return nil
}
