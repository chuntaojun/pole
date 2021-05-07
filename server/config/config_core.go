// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"context"
	"net/http"
	"sync"

	"github.com/golang/protobuf/ptypes"
	polerpc "github.com/pole-group/pole-rpc"

	"github.com/pole-group/pole/notify"
	"github.com/pole-group/pole/pojo"
	"github.com/pole-group/pole/server/sys"
	"github.com/pole-group/pole/utils"
)

//ConfigCore
type ConfigCore struct {
	filterChain *HandlerChain
	watcherMgn  *WatcherManager
	storageOp   StorageOperator
}

func newConfigCore() *ConfigCore {
	return &ConfigCore{}
}

type ConfigOpType int32

const (
	OpForPublishConfig ConfigOpType = iota
	OpForCreateConfig
	OpForModifyConfig
	OpForDeleteConfig
)

func (c *ConfigCore) operateConfig(op ConfigOpType, request *pojo.ConfigRequest, rpcCtx polerpc.RpcServerContext) {
	cFile := &ConfigFile{
		FType:   request.FileType,
		Version: utils.GetCurrentTimeMs(),
	}
	err := c.filterChain.Do(context.TODO(), cFile)
	if err != nil {
		rpcCtx.Send(&polerpc.ServerResponse{
			Code: http.StatusInternalServerError,
			Msg:  err.Error(),
		})
		return
	}
	if request.BetaInfo.Open {
		c.operateBetaConfig(op, &ConfigBetaFile{
			Cfg:         cFile,
			BetaClients: request.BetaInfo.ClientIds,
		})
	} else {
		c.operateNormalConfig(op, cFile)
	}

}

func (c *ConfigCore) operateNormalConfig(op ConfigOpType, cfg *ConfigFile) {
	if op == OpForCreateConfig || op == OpForModifyConfig {
		c.storageOp.SaveConfig(cfg)
	}

}

func (c *ConfigCore) operateBetaConfig(op ConfigOpType, cfg *ConfigBetaFile) {
	if op == OpForCreateConfig || op == OpForModifyConfig {
	}

}

func (c *ConfigCore) listenConfig(request *pojo.ConfigWatchRequest, rpcCtx polerpc.RpcServerContext) {

}

//WatcherManager
type WatcherManager struct {
	lock              sync.RWMutex
	watcherRepository map[string]*aggWatcher // <namespace, aggWatcher>
	scheduler         polerpc.RoutinePool
}

func (mgn *WatcherManager) Init() {
	if err := notify.RegisterSubscriber(mgn); err != nil {
		panic(err)
	}
}

//AddWatcher
func (mgn *WatcherManager) AddWatcher(req *pojo.ConfigWatchRequest, rpcCtx polerpc.RpcServerContext) (bool, error) {
	id := req.Id
	namespace := req.Namespace

	mgn.lock.Lock()
	if _, exist := mgn.watcherRepository[namespace]; !exist {
		mgn.watcherRepository[namespace] = newAggWatcher()
	}
	mgn.lock.Unlock()

	ids := mgn.watcherRepository[namespace]

	defer ids.lock.Unlock()
	ids.lock.Lock()
	if w, exist := ids.watcherMap[id]; exist {
		w.merge(req.WatchItemMap)
	} else {
		w := &watcher{
			lock:    sync.RWMutex{},
			id:      req.Id,
			itemMap: make(map[string]*watchKey),
			sink:    rpcCtx,
		}

		itemMap := make(map[string]*watchKey)
		for group, wi := range req.WatchItemMap {
			it := &watchKey{
				group:     group,
				filenames: polerpc.NewSet(),
			}
			for _, filename := range wi.FileName {
				it.filenames.Add(filename)
			}
			itemMap[group] = it
		}
		w.itemMap = itemMap
		ids.watcherMap[w.id] = w
	}

	return false, nil
}

//OnEvent 处理配置变更事件
//case one: 处理正常的配置变更事件
//case two: 处理灰度的配置变更事件
func (mgn *WatcherManager) OnEvent(event notify.Event) {
	switch e := event.(type) {
	case *ConfigChangeEvent:
		namespace := e.Namespace

		mgn.lock.RLock()

		ids, exist := mgn.watcherRepository[namespace]
		if !exist {
			sys.ConfigWatchLogger.Error("namespace %s don't exist", namespace)
			mgn.lock.RUnlock()
			return
		}
		mgn.lock.RUnlock()

		group := e.Group

		defer ids.lock.RUnlock()
		ids.lock.RLock()

		for _, watcher := range ids.watcherMap {
			if _, ok := watcher.itemMap[group]; ok {
				polerpc.Go(e, func(v interface{}) {
					event := v.(*ConfigChangeEvent)
					watcher.onConfigChange(event)
				})
			}
		}
	case *ConfigBetaChangeEvent:

	default:
	}
}

//SubscribeTypes 获取订阅的事件类型列表
func (mgn *WatcherManager) SubscribeTypes() []notify.Event {
	return []notify.Event{&ConfigBetaChangeEvent{}, &ConfigChangeEvent{}}
}

//IgnoreExpireEvent 是否忽略过期的事件
func (mgn *WatcherManager) IgnoreExpireEvent() bool {
	return false
}

type aggWatcher struct {
	lock       sync.RWMutex
	watcherMap map[string]*watcher //<id, watcher>
}

//newAggWatcher 创建一个聚合 watcher
func newAggWatcher() *aggWatcher {
	return &aggWatcher{
		lock:       sync.RWMutex{},
		watcherMap: make(map[string]*watcher),
	}
}

//watchKey
type watchKey struct {
	group     string
	filenames *polerpc.Set
}

//watcher
type watcher struct {
	lock    sync.RWMutex
	id      string
	itemMap map[string]*watchKey //<group, watchKey>
	sink    polerpc.RpcServerContext
}

//merge 合并监听信息数据
func (w *watcher) merge(items map[string]*pojo.WatchItem) {
	defer w.lock.Unlock()
	w.lock.Lock()
	for group, watchItem := range items {
		filenames := watchItem.FileName
		if i, exist := w.itemMap[group]; exist {
			// 监听文件列表的聚合操作
			for _, filename := range filenames {
				i.filenames.Add(filename)
			}
		} else {
			w.itemMap[group] = &watchKey{
				group:     group,
				filenames: polerpc.NewSetWithValues(filenames),
			}
		}
	}
}

//onConfigChange 相关配置文件变化，需要调用此方法，由每个 watcher 自己去决定需不需要将配置通知到 client 侧
func (w *watcher) onConfigChange(event *ConfigChangeEvent) {
	filename := event.FileName
	items := w.itemMap[event.Group]
	if items != nil && items.filenames.Contain(filename) {
		cfgFile := &pojo.ConfigFile{
			Meta: &pojo.ConfigMeta{
				NamespaceId: event.Namespace,
				Group:       event.Group,
				Encrypt:     event.IsEncrypt,
				FileType:    event.FType,
			},
			FileName: event.FileName,
			Content:  event.Content,
			Version:  event.Version,
		}

		any, err := ptypes.MarshalAny(cfgFile)
		if err != nil {
			sys.ConfigWatchLogger.Error("marshal proto.Message to proto.Any failed : %s", err)
			return
		}

		resp := &polerpc.ServerResponse{
			Code: 0,
			Body: any,
			Msg:  "success",
		}

		w.sink.Send(resp)
	}
}
