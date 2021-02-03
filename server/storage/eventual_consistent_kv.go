// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package storage

import (
	polerpc "github.com/pole-group/pole-rpc"

	"github.com/pole-group/pole/common"
)

type DataChangeListener func(isRemote bool, op DataOp, key, value []byte)

type APKVStorage struct {
	listeners *polerpc.ConcurrentSlice
}

func NewAPKVStorage(ctx *common.ContextPole, kvType KvType, slot int32) (*APKVStorage, error) {
	rks := &APKVStorage{
		listeners: polerpc.NewConcurrentSlice(func(opts *polerpc.CSliceOptions) {
		}),
	}
	return rks, rks.init()
}

func (apKv *APKVStorage) RegisterListener(listener DataChangeListener) {
	apKv.listeners.Add(listener)
}

func (apKv *APKVStorage) init() error {
	return nil
}
