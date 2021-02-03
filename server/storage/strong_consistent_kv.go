// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package storage

import (
	"sync"

	"github.com/pole-group/pole/common"
	cy "github.com/pole-group/pole/server/consistency"
	"github.com/pole-group/pole/server/sys"
)

type RangeRegionKVStorage struct {
	ctx         *common.ContextPole
	lock        sync.RWMutex
	kvType      KvType
	currentSize int64
	originKv    KVStorage
}

type KvOp int32

const (
	KvOpForRead        KvOp = 0xf11
	KvOpForBatchRead   KvOp = 0xf12
	KvOpForWrite       KvOp = 0xf13
	KvOpForBatchWrite  KvOp = 0xf14
	KvOpForDelete      KvOp = 0x15
	KvOpForBatchDelete KvOp = 0x16
)

type RaftKvStorage struct {
	rwLock      sync.RWMutex
	seqNo       uint64
	name        string
	kv          KVStorage
	raftService cy.Protocol
}

func newRaftKvStorage(ctx *common.ContextPole, name string, seqNo uint64, t KvType) (*RaftKvStorage, error) {

	kv, err := NewKVStorage(ctx, t)
	if err != nil {
		return nil, err
	}

	rKv := &RaftKvStorage{
		rwLock: sync.RWMutex{},
		name:   name,
		seqNo:  seqNo,
		kv:     kv,
	}
	return rKv, rKv.init(ctx)
}

func (rk *RaftKvStorage) init(cxt *common.ContextPole) error {
	if err := rk.raftService.AddProcessor(rk); err != nil {
		return err
	}
	return nil
}

func (rk *RaftKvStorage) RegisterHook(t HookType, h KvHook) {

}

func (rk *RaftKvStorage) Read(ctx *common.ContextPole, key []byte) ([]byte, error) {
	return nil, nil
}

func (rk *RaftKvStorage) ReadBatch(ctx *common.ContextPole, keys [][]byte) ([][]byte, error) {
	return nil, nil
}

func (rk *RaftKvStorage) Write(ctx *common.ContextPole, key []byte, value []byte) error {
	return nil
}

func (rk *RaftKvStorage) WriteBatch(ctx *common.ContextPole, keys [][]byte, values [][]byte) error {
	return nil
}

func (rk *RaftKvStorage) Delete(ctx *common.ContextPole, key []byte) error {
	return nil
}

func (rk *RaftKvStorage) DeleteBatch(ctx *common.ContextPole, keys [][]byte) error {
	return nil
}

func (rk *RaftKvStorage) Size() int64 {
	return 1
}

func (rk *RaftKvStorage) Destroy() error {
	return nil
}

// 处理读请求
func (rk *RaftKvStorage) OnRead(req *cy.Read) cy.Response {
	return cy.Response{}
}

// 处理写请求
func (rk *RaftKvStorage) OnWrite(req *cy.Write) cy.Response {
	return cy.Response{}
}

// 处理错误
func (rk *RaftKvStorage) OnError(err error) {
	sys.RaftKVLogger.Error("occur error in raft state machine : %s", err)
}

// 分组详细
func (rk *RaftKvStorage) Group() (string, uint64) {
	return rk.name, rk.seqNo
}

func (rk *RaftKvStorage) OnSnapshotSave() {

}

func (rk *RaftKvStorage) OnSnapshotLoad() {

}
