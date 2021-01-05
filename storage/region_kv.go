// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package storage

import (
	"fmt"
	"strings"
	"sync"

	"github.com/Conf-Group/pole/common"
	"github.com/Conf-Group/pole/utils"
)

type proxyKvTree struct {
	lock sync.RWMutex
	root *kvNode
	size int64
}

type kvNode struct {
	val    proxyKVStorage
	parent *kvNode
	left   *kvNode
	right  *kvNode
}

func newProxyKvTree() *proxyKvTree {
	return &proxyKvTree{
		lock: sync.RWMutex{},
		root: nil,
		size: 0,
	}
}

func (kt *proxyKvTree) Find(v string) *kvNode {
	defer kt.lock.RUnlock()
	kt.lock.RLock()
	return kt.findNode(v, kt.root)
}

func (kt *proxyKvTree) findNode(v string, root *kvNode) *kvNode {
	if root == nil {
		return nil
	}
	if strings.Compare(v, root.val.minKey) >= 0 && strings.Compare(v, root.val.maxKey) < 0 {
		return root
	}
	if n := kt.findNode(v, root.left); n != nil {
		return n
	}
	if n := kt.findNode(v, root.right); n != nil {
		return n
	}
	return nil
}

func (kt *proxyKvTree) FindMax() *kvNode {
	defer kt.lock.RUnlock()
	kt.lock.RLock()
	return kt.findMaxNode(kt.root)
}

func (kt *proxyKvTree) findMaxNode(root *kvNode) *kvNode {
	if root == nil {
		return nil
	}
	if root.left == nil && root.right == nil {
		return root
	}
	return kt.findMaxNode(root.right)
}

func (kt *proxyKvTree) FindMin() *kvNode {
	defer kt.lock.RUnlock()
	kt.lock.RLock()
	return kt.findMinNode(kt.root)
}

func (kt *proxyKvTree) findMinNode(root *kvNode) *kvNode {

	if root == nil {
		return nil
	}
	if root.left == nil && root.right == nil {
		return root
	}
	return kt.findMinNode(root.left)
}

func (kt *proxyKvTree) Insert(v proxyKVStorage) {
	defer kt.lock.Unlock()
	kt.lock.Lock()
	kt.root = kt.insertVal(v, kt.root, kt.root)
	kt.size++
}

func (kt *proxyKvTree) insertVal(v proxyKVStorage, root *kvNode, parent *kvNode) *kvNode {
	if root == nil {
		return &kvNode{
			val:    v,
			parent: parent,
			left:   nil,
			right:  nil,
		}
	}
	if compareKvNode(v, parent.val) < 0 {
		root.left = kt.insertVal(v, root.left, root)
	}
	if compareKvNode(v, parent.val) > 0 {
		root.right = kt.insertVal(v, root.right, root)
	}
	return root
}

func (kt *proxyKvTree) Delete(v proxyKVStorage) {
	defer kt.lock.Unlock()
	kt.lock.Lock()
	kt.deleteVal(v, kt.root)
	kt.size--
}

func (kt *proxyKvTree) deleteVal(v proxyKVStorage, root *kvNode) *kvNode {
	if root == nil {
		return nil
	}
	if compareKvNode(v, root.val) < 0 {
		root.left = kt.deleteVal(v, root.left)
		root.left.parent = root
	} else if compareKvNode(v, root.val) > 0 {
		root.right = kt.deleteVal(v, root.right)
		root.right.parent = root
	} else if root.left != nil && root.right != nil {
		rMin := kt.findMinNode(root.right)
		root.val = rMin.val
		root.right = kt.deleteVal(rMin.val, root.right)
	} else {
		if root.left == nil {
			root = root.right
		} else if root.right != nil {
			root = root.left
		}
	}
	return root
}

func (kt *proxyKvTree) Seek(start, end *kvNode, call func(n *kvNode) error) error {
	defer kt.lock.RUnlock()
	kt.lock.RLock()
	return kt.rangeVal(kt.root, call)
}

func (kt *proxyKvTree) Range(call func(n *kvNode) error) error {
	defer kt.lock.RUnlock()
	kt.lock.RLock()
	return kt.rangeVal(kt.root, call)
}

func (kt *proxyKvTree) rangeVal(root *kvNode, call func(n *kvNode) error) error {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("err : %#v", err)
		}
	}()
	if root != nil {
		if err := call(root); err != nil {
			return err
		}
		if err := kt.rangeVal(root.left, call); err != nil {
			return err
		}
		if err := kt.rangeVal(root.right, call); err != nil {
			return err
		}
	}
	return nil
}

// 一定不能出现范围交错在一起的现象出现！
func compareKvNode(a, b proxyKVStorage) int {
	if strings.Compare(a.maxKey, b.minKey) < 0 {
		return -1
	}
	if strings.Compare(a.minKey, b.maxKey) > 0 {
		return 1
	}
	return 0
}

const maxValueThresholdNumber int64 = 16384

type proxyKVStorage struct {
	owner  *RegionKVStorage
	minKey string
	maxKey string
	kv     KVStorage
	size   int64
}

func (sk proxyKVStorage) destroy() {
	_ = sk.kv.Destroy()
}

func newProxyKVStorage(kv KVStorage, rkv *RegionKVStorage) proxyKVStorage {
	sk := proxyKVStorage{
		owner:  rkv,
		kv:     kv,
		minKey: "",
		maxKey: "",
	}

	sk.kv.RegisterHook(HookForBeforeWrite, func(key string, value []byte) {
		sk.minKey = utils.IF(strings.Compare(key, sk.minKey) < 0, func() interface{} {
			return key
		}, func() interface{} {
			return sk.minKey
		}).(string)
		sk.maxKey = utils.IF(strings.Compare(key, sk.maxKey) > 0, func() interface{} {
			return key
		}, func() interface{} {
			return sk.maxKey
		}).(string)
	})

	sk.kv.RegisterHook(HookForBeforeWrite, func(key string, value []byte) {
		sk.size++
		if sk.size > maxValueThresholdNumber {
			utils.GoEmpty(func() {
				sk.owner.needSplitSign <- sk
			})
		}
	})

	return sk
}

func (sk proxyKVStorage) canProcess(k []byte) bool {
	return strings.Compare(string(k), sk.minKey) >= 0 && strings.Compare(string(k), sk.maxKey) <= 0
}

type RegionKVStorage struct {
	ctx           *common.ContextPole
	lock          sync.RWMutex
	kvType        KvType
	currentSize   int64
	originKv      KVStorage
	regions       *proxyKvTree
	needSplitSign chan proxyKVStorage
}

func NewRegionKVStorage(ctx *common.ContextPole, kvType KvType) (*RegionKVStorage, error) {
	kv, err := NewKVStorage(ctx, kvType)

	if err != nil {
		return nil, err
	}

	rks := &RegionKVStorage{
		lock:          sync.RWMutex{},
		kvType:        kvType,
		currentSize:   0,
		originKv:      kv,
		regions:       newProxyKvTree(),
		needSplitSign: make(chan proxyKVStorage, 4),
	}
	return rks, rks.init()
}

func (rks *RegionKVStorage) init() error {
	rks.regions.Insert(newProxyKVStorage(rks.originKv, rks))

	utils.GoEmpty(func() {
		for pKv := range rks.needSplitSign {
			if err := rks.split(pKv); err != nil {
				rks.needSplitSign <- pKv
			}
		}
	})

	return nil
}

func (rks *RegionKVStorage) RegisterHook(t HookType, h KvHook) {
	rks.originKv.RegisterHook(t, h)
}

func (rks *RegionKVStorage) Read(key []byte) ([]byte, error) {
	defer rks.lock.RUnlock()
	rks.lock.RLock()

	node := rks.regions.Find(string(key))
	if node == nil {
		return nil, fmt.Errorf("can't find target response proxy for key : %s", string(key))
	}
	return node.val.kv.Read(key)
}

func (rks *RegionKVStorage) ReadBatch(keys [][]byte) ([][]byte, error) {
	defer rks.lock.RUnlock()
	rks.lock.RLock()

	values := make([][]byte, len(keys), len(keys))
	for i, key := range keys {
		node := rks.regions.Find(string(key))
		if node == nil {
			return nil, fmt.Errorf("can't find target response proxy for key : %s", string(key))
		}
		v, err := node.val.kv.Read(key)
		if err != nil {
			return nil, err
		}
		values[i] = v
	}
	return values, nil
}

func (rks *RegionKVStorage) Write(key []byte, value []byte) error {
	defer rks.lock.RUnlock()
	rks.lock.RLock()

	node := rks.regions.Find(string(key))
	if node == nil {
		return fmt.Errorf("can't find target response proxy for key : %s", string(key))
	}
	return node.val.kv.Write(key, value)
}

func (rks *RegionKVStorage) WriteBatch(keys [][]byte, values [][]byte) error {
	defer rks.lock.RUnlock()
	rks.lock.RLock()

	for i, key := range keys {
		node := rks.regions.Find(string(key))
		if node == nil {
			return fmt.Errorf("can't find target response proxy for key : %s", string(key))
		}
		if err := node.val.kv.Write(key, values[i]); err != nil {
			return err
		}
	}

	return nil
}

func (rks *RegionKVStorage) Delete(key []byte) error {
	defer rks.lock.RUnlock()
	rks.lock.RLock()

	node := rks.regions.Find(string(key))
	if node == nil {
		return fmt.Errorf("can't find target response proxy for key : %s", string(key))
	}
	return node.val.kv.Delete(key)
}

func (rks *RegionKVStorage) DeleteBatch(keys [][]byte) error {
	defer rks.lock.RUnlock()
	rks.lock.RLock()

	for _, key := range keys {
		node := rks.regions.Find(string(key))
		if node == nil {
			return fmt.Errorf("can't find target response proxy for key : %s", string(key))
		}
		if err := node.val.kv.Delete(key); err != nil {
			return err
		}
	}

	return nil
}

func (rks *RegionKVStorage) merge(pkvL, pkvR proxyKVStorage) error {
	return nil
}

func (rks *RegionKVStorage) split(pkv proxyKVStorage) error {
	defer rks.lock.Unlock()
	rks.lock.Lock()

	currentSize := pkv.size
	_ = currentSize / 2

	pKvR := newProxyKVStorage(rks.originKv, rks)

	pKvR.maxKey = pkv.maxKey


	return nil
}
