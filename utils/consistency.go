// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utils

import (
	"hash/crc32"
	"sort"
	"strings"
	"sync"
)

type ConsistentHash struct {
	rwLock         sync.RWMutex
	sortNode       []uint32
	virtualNodeNum int32
	nodeMap        map[string]bool
	circleMap      map[uint32]string
}

func NewConsistentHash(virtualNodeNum int32) *ConsistentHash {
	ch := &ConsistentHash{
		rwLock:         sync.RWMutex{},
		sortNode:       make([]uint32, 0, 0),
		virtualNodeNum: virtualNodeNum,
		nodeMap:        make(map[string]bool),
		circleMap:      make(map[uint32]string),
	}
	return ch
}

func (ch *ConsistentHash) AddNode(n string) {
	defer ch.rwLock.Unlock()
	ch.rwLock.Lock()
	if _, exist := ch.nodeMap[n]; !exist {
		ch.nodeMap[n] = true
		for i := int32(0); i < ch.virtualNodeNum; i++ {
			for {
				position := crc32.ChecksumIEEE([]byte(n))
				if _, exist := ch.circleMap[position]; !exist {
					ch.circleMap[position] = n
					ch.sortNode = append(ch.sortNode, position)
					break
				}
			}
		}
	}

	sort.Slice(ch.sortNode, func(i, j int) bool {
		return ch.sortNode[i] < ch.sortNode[j]
	})
}

func (ch *ConsistentHash) RemoveNode(n string) {
	defer ch.rwLock.Unlock()
	ch.rwLock.Lock()
	if _, exist := ch.nodeMap[n]; !exist {
		delete(ch.nodeMap, n)

		newCircle := make(map[uint32]string)
		newSortNode := make([]uint32, 0, 0)

		for k, v := range ch.circleMap {
			if strings.Compare(v, n) == 0 {
				continue
			}
			newCircle[k] = v
			newSortNode = append(newSortNode, k)
		}

		ch.sortNode = newSortNode
		ch.circleMap = newCircle

		sort.Slice(ch.sortNode, func(i, j int) bool {
			return ch.sortNode[i] < ch.sortNode[j]
		})
	}
}

func (ch *ConsistentHash) GetNode(key string) string {
	defer ch.rwLock.RUnlock()
	ch.rwLock.RLock()

	code := crc32.ChecksumIEEE([]byte(key))

	position := sort.Search(len(ch.sortNode), func(i int) bool {
		return ch.sortNode[i] >= code
	})

	if position < len(ch.sortNode) {
		p := IF(position == len(ch.sortNode)-1, 0, position).(int)
		return ch.circleMap[ch.sortNode[p]]
	} else {
		return ch.circleMap[ch.sortNode[len(ch.sortNode)-1]]
	}

}
