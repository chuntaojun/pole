// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"

	"nacos-go/consistency"
	"nacos-go/utils"
)

type RaftProtocol struct {
	raftDir string

	lock sync.Locker

	servers map[string]*RaftServer

	Customer map[string]*consistency.LogProcessor4CP
}

type RaftServer struct {
	mu sync.RWMutex

	raft         *raft.Raft
	raftObserver raft.Observer
	raftTn       *raft.NetworkTransport
	raftLog      raft.LogStore
	raftStable   raft.StableStore
	boltStore    *raftboltdb.BoltStore

	ln net.Listener

	ShutdownOnRemove  bool
	SnapshotThreshold uint64
	SnapshotInterval  time.Duration
	HeartbeatTimeout  time.Duration
	ElectionTimeout   time.Duration
	ApplyTimeout      time.Duration
	RaftLogLevel      string
}

func newRaftServer(dir string, port int, c consistency.RaftConfig, l *consistency.LogProcessor) (*RaftServer, error) {

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("create raft-server has error : %s\n", utils.PrintStack())
		}
	}()

	ln := utils.BiFunction("tcp", "0.0.0.0:"+string(port), func(v, t interface{}) (i interface{}, err error) {
		return net.Listen(v.(string), t.(string))
	}).(net.Listener)

	raftTn := raft.NewNetworkTransport(NewTransport(ln), 30, time.Duration(10)*time.Second, nil)

	raftLog := utils.Function(filepath.Join(dir, "raft-log.bolt"), func(v interface{}) (i interface{}, err error) {
		return raftboltdb.NewBoltStore(v.(string))
	}).(*raftboltdb.BoltStore)

	stableStore := utils.Function(filepath.Join(dir, "raft-stable.bolt"), func(v interface{}) (i interface{}, err error) {
		return raftboltdb.NewBoltStore(v.(string))
	}).(*raftboltdb.BoltStore)

	snapshots := utils.Function(dir, func(v interface{}) (i interface{}, err error) {
		return raft.NewFileSnapshotStore(v.(string), 2, os.Stderr)
	}).(raft.SnapshotStore)

	r := &RaftServer{
		raftLog:    raftLog,
		raftStable: stableStore,
		raftTn:     raftTn,
	}

	config := r.raftConfig()
	config.LocalID = raft.ServerID((*l).Group() + string(port))

	ra, err := raft.NewRaft(config, r, r.raftLog, r.raftStable, snapshots, r.raftTn)
	r.raft = ra
	return r, err
}

// raftConfig returns a new Raft config for the store.
func (r *RaftServer) raftConfig() *raft.Config {
	config := raft.DefaultConfig()
	config.ShutdownOnRemove = r.ShutdownOnRemove
	config.LogLevel = r.RaftLogLevel
	if r.SnapshotThreshold != 0 {
		config.SnapshotThreshold = r.SnapshotThreshold
	}
	if r.SnapshotInterval != 0 {
		config.SnapshotInterval = r.SnapshotInterval
	}
	if r.HeartbeatTimeout != 0 {
		config.HeartbeatTimeout = r.HeartbeatTimeout
	}
	if r.ElectionTimeout != 0 {
		config.ElectionTimeout = r.ElectionTimeout
	}
	return config
}

func (r *RaftServer) Apply(log *raft.Log) interface{} {
	return nil
}

func (r *RaftServer) Snapshot() (raft.FSMSnapshot, error) {
	return nil, nil
}

func (r *RaftServer) Restore(io.ReadCloser) error {
	return nil
}

type raftSnapshot struct {
}

func (r *raftSnapshot) Persist(sink raft.SnapshotSink) error {
	return nil
}

func (r *raftSnapshot) Release() {

}

// Transport is the network service provided to Raft, and wraps a Listener.
type Transport struct {
	ln net.Listener
}

// NewTransport returns an initialized Transport.
func NewTransport(ln net.Listener) *Transport {
	return &Transport{
		ln: ln,
	}
}

// Dial creates a new network connection.
func (t *Transport) Dial(addr raft.ServerAddress, timeout time.Duration) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: timeout}
	conn, err := dialer.Dial("tcp", string(addr))
	return conn, err
}

// Accept waits for the next connection.
func (t *Transport) Accept() (net.Conn, error) {
	return t.ln.Accept()
}

// Close closes the transport
func (t *Transport) Close() error {
	return t.ln.Close()
}

// Addr returns the binding address of the transport.
func (t *Transport) Addr() net.Addr {
	return t.ln.Addr()
}
