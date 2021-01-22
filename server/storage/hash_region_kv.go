// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package storage

import (
	"github.com/pole-group/pole/common"
	"github.com/pole-group/pole/utils"
)

type HashRegionKVStorage struct {
	hash *utils.ConsistentHash
}

func NewHashRegionKVStorage(ctx *common.ContextPole, kvType KvType, slot int32) (*HashRegionKVStorage, error) {
	rks := &HashRegionKVStorage{
		hash: utils.NewConsistentHash(slot),
	}
	return rks, rks.init()
}

func (hrk *HashRegionKVStorage) init() error {
	return nil
}
