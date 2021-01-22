package storage

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/dgraph-io/badger/v2"

	"github.com/pole-group/pole/common"
	"github.com/pole-group/pole/utils"
)

type HookType int32

type KvHook func(key string, value []byte)

const (
	HookForBeforeWrite HookType = iota
	HookForAfterWrite
	HookForBeforeDelete
	HookForAfterDelete
)

type kvHookHolder struct {
	hLock sync.RWMutex
	hooks map[HookType]utils.Set
}

func newKvHookHolder() *kvHookHolder {
	return &kvHookHolder{
		hLock: sync.RWMutex{},
		hooks: make(map[HookType]utils.Set),
	}
}

func (kh *kvHookHolder) RegisterHook(t HookType, h KvHook) {
	defer kh.hLock.Unlock()
	kh.hLock.Lock()

	if _, exist := kh.hooks[t]; !exist {
		kh.hooks[t] = utils.NewSet()
	}
	kh.hooks[t].Add(h)
}

func (kh *kvHookHolder) RemoveHook(t HookType, h KvHook) {
	defer kh.hLock.Unlock()
	kh.hLock.Lock()
	if hooks, exist := kh.hooks[t]; exist {
		hooks.Remove(h)
	}
}

func (kh *kvHookHolder) BeforeWrite(key string, value []byte) {
	kh.execHook(HookForBeforeWrite, key, value)
}

func (kh *kvHookHolder) AfterWrite(key string, value []byte) {
	kh.execHook(HookForAfterWrite, key, value)
}

func (kh *kvHookHolder) BeforeDelete(key string, value []byte) {
	kh.execHook(HookForBeforeDelete, key, value)
}

func (kh *kvHookHolder) AfterDelete(key string, value []byte) {
	kh.execHook(HookForAfterDelete, key, value)
}

func (kh *kvHookHolder) execHook(t HookType, k string, v []byte) {
	kh.hLock.RLock()
	hooks := kh.hooks[t]
	kh.hLock.RUnlock()
	hooks.Range(func(h interface{}) {
		h.(KvHook)(k, v)
	})
}

type KvType int16

const (
	MemoryKv KvType = iota
	BadgerKv
	BadgerMemoryKv
)

func NewKVStorage(ctx *common.ContextPole, t KvType) (KVStorage, error) {
	switch t {
	case MemoryKv:
		kv := &memoryKVStorage{}
		return kv, kv.init(ctx)
	case BadgerKv:
		kv := &badgerKVStorage{}
		return kv, kv.init(ctx)
	default:
		return nil, nil
	}
}

type KVStorage interface {
	init(cxt *common.ContextPole) error

	RegisterHook(t HookType, h KvHook)

	Read(ctx *common.ContextPole, key []byte) ([]byte, error)

	ReadBatch(ctx *common.ContextPole, keys [][]byte) ([][]byte, error)

	Write(ctx *common.ContextPole, key []byte, value []byte) error

	WriteBatch(ctx *common.ContextPole, keys [][]byte, values [][]byte) error

	Delete(ctx *common.ContextPole, key []byte) error

	DeleteBatch(ctx *common.ContextPole, keys [][]byte) error

	Size() int64

	Destroy() error
}

type keyTree struct {
	lock sync.RWMutex
	root *node
	size int64
}

type node struct {
	val    string
	parent *node
	left   *node
	right  *node
}

func newKeyTree() *keyTree {
	return &keyTree{
		lock: sync.RWMutex{},
		root: nil,
		size: 0,
	}
}

func (kt *keyTree) SeekLevel() [][]*node {
	defer kt.lock.RUnlock()
	kt.lock.RLock()

	if kt.root == nil {
		return nil
	}

	ans := make([][]*node, 0, 0)
	tmp := make([]*node, 0, 0)
	_stack := make([]*node, 0, 0)
	_stack = append(_stack, kt.root)
	nowNodeSize := len(_stack)
	for len(_stack) != 0 {
		if nowNodeSize == 0 {
			ans = append(ans, tmp)
			tmp = make([]*node, 0, 0)
			nowNodeSize = len(_stack)
		}
		p := _stack[0]
		_stack = _stack[1:]
		nowNodeSize --
		tmp = append(tmp, p)
		if p.left != nil {
			_stack = append(_stack, p.left)
		}
		if p.right != nil {
			_stack = append(_stack, p.right)
		}
	}
	ans = append(ans, tmp)
	return ans
}

//					5
//				  /   \
//				 3     8
//				/ \   / \
//             1   4 7   9
//
// if you find 5, will return 3, if find 1, will return nil
func (kt *keyTree) FindNearbyLeft(v string) *node {
	defer kt.lock.RUnlock()
	kt.lock.RLock()
	return kt.findNearbyLeftNode(v, kt.root)
}

func (kt *keyTree) findNearbyLeftNode(v string, root *node) *node {
	if root == nil {
		return nil
	}
	if strings.Compare(v, root.val) <= 0 {
		if root.left != nil {
			if strings.Compare(v, root.left.val) > 0 {
				return root.left
			} else {
				return kt.findNearbyLeftNode(v, root.left)
			}
		}
		return root
	} else {
		return kt.findNearbyLeftNode(v, root.right)
	}
}

//					5
//				  /   \
//				 3     8
//				/ \   / \
//             1   4 7   9
//
// if you find 5, will return 7, if find 8, will return 9
func (kt *keyTree) FindNearbyRight(v string) *node {
	defer kt.lock.RUnlock()
	kt.lock.RLock()
	return kt.findMaxNode(kt.root)
}

func (kt *keyTree) FindMax() *node {
	defer kt.lock.RUnlock()
	kt.lock.RLock()
	return kt.findMaxNode(kt.root)
}

func (kt *keyTree) findMaxNode(root *node) *node {
	if root == nil {
		return nil
	}
	if root.left == nil && root.right == nil {
		return root
	}
	return kt.findMaxNode(root.left)
}

func (kt *keyTree) FindMin() *node {
	defer kt.lock.RUnlock()
	kt.lock.RLock()
	return kt.findMinNode(kt.root)
}

func (kt *keyTree) findMinNode(root *node) *node {
	if root == nil {
		return nil
	}
	if root.left == nil && root.right == nil {
		return root
	}
	return kt.findMinNode(root.left)
}

func (kt *keyTree) Insert(v string) {
	defer kt.lock.Unlock()
	kt.lock.Lock()
	kt.root = kt.insertVal(v, kt.root, kt.root)
	kt.size++
}

func (kt *keyTree) insertVal(v string, root *node, parent *node) *node {
	if root == nil {
		return &node{
			val:    v,
			parent: parent,
			left:   nil,
			right:  nil,
		}
	}
	if strings.Compare(v, root.val) < 0 {
		root.left = kt.insertVal(v, root.left, root)
	} else if strings.Compare(v, root.val) > 0 {
		root.right = kt.insertVal(v, root.right, root)
	}
	return root
}

func (kt *keyTree) Delete(v string) {
	defer kt.lock.Unlock()
	kt.lock.Lock()
	kt.deleteVal(v, kt.root)
	kt.size--
}

func (kt *keyTree) deleteVal(v string, root *node) *node {
	if root == nil {
		return nil
	}
	if strings.Compare(v, root.val) < 0 {
		root.left = kt.deleteVal(v, root.left)
		root.left.parent = root
	} else if strings.Compare(v, root.val) > 0 {
		root.right = kt.deleteVal(v, root.right)
		root.right.parent = root
	} else if root.left != nil && root.right != nil {
		rMin := kt.findMinNode(root.right)
		root.val = rMin.val
		root.right = kt.deleteVal(rMin.val, root.right)
	} else {
		tmpCell := root
		if root.left == nil {
			root = root.right
		} else if root.right != nil {
			root = root.left
		}

		tmpCell.left = nil
		tmpCell.right = nil
		tmpCell.parent = nil
	}
	return root
}

func (kt *keyTree) Range(call func(n *node) error) error {
	defer kt.lock.RUnlock()
	kt.lock.RLock()
	return kt.rangeVal(kt.root, call)
}

func (kt *keyTree) rangeVal(root *node, call func(n *node) error) error {
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

type memoryKVStorage struct {
	lock      sync.RWMutex
	kh        *kvHookHolder
	sortKeys  *keyTree
	reference atomic.Value // map[string][]byte
}

func (mks *memoryKVStorage) init(cxt *common.ContextPole) error {
	mks.lock = sync.RWMutex{}
	mks.reference = atomic.Value{}
	mks.reference.Store(make(map[string][]byte))
	mks.kh = newKvHookHolder()
	mks.sortKeys = newKeyTree()

	mks.RegisterHook(HookForBeforeWrite, func(key string, value []byte) {
		mks.sortKeys.Insert(key)
	})

	mks.RegisterHook(HookForBeforeDelete, func(key string, value []byte) {
		mks.sortKeys.Delete(key)
	})

	return nil
}

func (mks *memoryKVStorage) RegisterHook(t HookType, h KvHook) {
	mks.kh.RegisterHook(t, h)
}

func (mks *memoryKVStorage) RemoveHook(t HookType, h KvHook) {
	mks.kh.RegisterHook(t, h)
}

func (mks *memoryKVStorage) Read(ctx *common.ContextPole, key []byte) ([]byte, error) {
	defer mks.lock.RUnlock()
	mks.lock.RLock()
	return mks.reference.Load().(map[string][]byte)[string(key)], nil
}

func (mks *memoryKVStorage) ReadBatch(ctx *common.ContextPole, keys [][]byte) ([][]byte, error) {
	defer mks.lock.RUnlock()
	mks.lock.RLock()
	values := make([][]byte, len(keys), len(keys))
	for i, key := range keys {
		values[i] = mks.reference.Load().(map[string][]byte)[string(key)]
	}
	return values, nil
}

func (mks *memoryKVStorage) Write(ctx *common.ContextPole, key []byte, value []byte) error {
	defer mks.lock.Unlock()
	mks.lock.Lock()
	k := string(key)
	mks.kh.BeforeWrite(k, value)
	mks.reference.Load().(map[string][]byte)[k] = value
	mks.kh.AfterWrite(k, value)
	return nil
}

func (mks *memoryKVStorage) WriteBatch(ctx *common.ContextPole, keys [][]byte, values [][]byte) error {
	defer mks.lock.Unlock()
	mks.lock.Lock()
	for i, key := range keys {
		k := string(key)
		v := values[i]
		mks.kh.BeforeWrite(k, v)
		mks.reference.Load().(map[string][]byte)[k] = values[i]
		mks.kh.AfterWrite(k, v)
	}
	return nil
}

func (mks *memoryKVStorage) Delete(ctx *common.ContextPole, key []byte) error {
	defer mks.lock.Unlock()
	mks.lock.Lock()
	delete(mks.reference.Load().(map[string][]byte), string(key))
	return nil
}

func (mks *memoryKVStorage) DeleteBatch(ctx *common.ContextPole, keys [][]byte) error {
	defer mks.lock.Unlock()
	mks.lock.Lock()
	for _, key := range keys {
		delete(mks.reference.Load().(map[string][]byte), string(key))
	}
	return nil
}

func (mks *memoryKVStorage) Size() int64 {
	return mks.sortKeys.size
}

func (mks *memoryKVStorage) Destroy() error {
	mks.reference = atomic.Value{}
	mks.kh = nil
	return nil
}

type badgerKVStorage struct {
	kv *badger.DB
	kh *kvHookHolder
}

func (bks *badgerKVStorage) init(cxt *common.ContextPole) error {
	bks.kh = newKvHookHolder()
	opt := badger.
		DefaultOptions(cxt.Value("kv_dir").(string)).
		WithMaxTableSize(1 << 15). // Force more compaction.
		WithLevelOneSize(4 << 15). // Force more compaction.
		WithSyncWrites(false).
		WithBlockCacheSize(10 << 20)
	kv, err := badger.Open(opt)
	bks.kv = kv
	return err
}

func (bks *badgerKVStorage) RegisterHook(t HookType, h KvHook) {
	bks.kh.RegisterHook(t, h)
}

func (bks *badgerKVStorage) Read(ctx *common.ContextPole, key []byte) ([]byte, error) {
	txn := bks.kv.NewTransaction(false)
	item, err := txn.Get(key)
	if err != nil {
		return nil, err
	}
	v := make([]byte, 0, 0)
	return item.ValueCopy(v)
}

func (bks *badgerKVStorage) ReadBatch(ctx *common.ContextPole, keys [][]byte) ([][]byte, error) {
	vs := make([][]byte, 0, 0)
	txn := bks.kv.NewTransaction(false)
	for _, key := range keys {
		item, err := txn.Get(key)
		if err != nil {
			return nil, err
		}
		v := make([]byte, 0, 0)
		v, err = item.ValueCopy(v)
		if err != nil {
			return nil, err
		}
		vs = append(vs, v)
	}
	return vs, nil
}

func (bks *badgerKVStorage) Write(ctx *common.ContextPole, key []byte, value []byte) error {
	txn := bks.kv.NewWriteBatch()
	if err := txn.Set(key, value); err != nil {
		txn.Cancel()
		return err
	} else {
		return txn.Flush()
	}
}

func (bks *badgerKVStorage) WriteBatch(ctx *common.ContextPole, keys [][]byte, values [][]byte) error {
	txn := bks.kv.NewWriteBatch()
	for i, k := range keys {
		if err := txn.Set(k, values[i]); err != nil {
			txn.Cancel()
			return err
		}
	}
	return txn.Flush()
}

func (bks *badgerKVStorage) Delete(ctx *common.ContextPole, key []byte) error {
	txn := bks.kv.NewTransaction(true)
	if err := txn.Delete(key); err != nil {
		txn.Discard()
		return err
	}
	return txn.Commit()
}

func (bks *badgerKVStorage) DeleteBatch(ctx *common.ContextPole, keys [][]byte) error {
	txn := bks.kv.NewTransaction(true)
	for _, key := range keys {
		if err := txn.Delete(key); err != nil {
			txn.Discard()
			return err
		}
	}
	return txn.Commit()
}

func (bks *badgerKVStorage) Size() int64 {
	lsm, vlog := bks.kv.Size()
	return lsm + vlog
}

func (bks *badgerKVStorage) Destroy() error {
	return bks.kv.Close()
}
