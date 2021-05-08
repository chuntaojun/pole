// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package discovery

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	polerpc "github.com/pole-group/pole-rpc"

	"github.com/pole-group/pole/notify"
	"github.com/pole-group/pole/pojo"
	"github.com/pole-group/pole/utils"
)

const (
	AND = "and"
	OR  = "or"
)

type ServiceChangeEvent struct {
}

func (se *ServiceChangeEvent) Name() string {
	return "ServiceChangeEvent"
}

func (se *ServiceChangeEvent) Sequence() int64 {
	return utils.GetCurrentTimeMs()
}

//DiscoveryCore 服务发现核心模块
type DiscoveryCore struct {
	serviceMgn      *ServiceManager
	subscriberMgn   *SubscriberManager
	storageOperator DiscoveryStorage
}

type QueryInstance struct {
	namespaceID string
	serviceName string
	clusterName string
	clientHost  string
}

type ServiceManager struct {
	lock       sync.RWMutex
	core       *DiscoveryCore
	lessHolder *Lessor
	services   map[string]map[string]*Service
}

func (sm *ServiceManager) addInstance(req *pojo.InstanceRegister, sink polerpc.RpcServerContext) {
	namespaceId := req.NamespaceId
	serviceName := req.Instance.ServiceName
	groupName := req.Instance.Group

	name := ParseServiceName(serviceName, groupName)
	service, err := sm.createServiceIfAbsent(namespaceId, serviceName, groupName)
	if err != nil {
		sink.Send(&polerpc.ServerResponse{
			Code: http.StatusInternalServerError,
			Msg:  err.Error(),
		})
		return
	}

	clusterName := req.Instance.ClusterName
	key := ParseInstanceKey(name, clusterName, req.GetInstance().Ip, int32(req.GetInstance().Port))

	instance, metadata := parseToInstance(key, req.Instance)

	// 注册一个 Instance, 会计算判断是否存在 cluster 信息，如果不存在会自动创建一个 Cluster
	sm.storeInstanceInfo(sink, service, instance, metadata)
}

func (sm *ServiceManager) removeInstance(req *pojo.InstanceDeregister, sink polerpc.RpcServerContext) {
	namespaceId := req.NamespaceId
	serviceName := req.Instance.ServiceName
	groupName := req.Instance.Group

	name := ParseServiceName(serviceName, groupName)
	defer sm.lock.RUnlock()
	sm.lock.RLock()

	namespace, exist := sm.services[namespaceId]
	if !exist {
		sink.Send(&polerpc.ServerResponse{
			Code: http.StatusInternalServerError,
			Msg:  fmt.Sprintf("namespace : %s not exist", namespaceId),
		})
		return
	}
	service, exist := namespace[name]
	if !exist {
		sink.Send(&polerpc.ServerResponse{
			Code: http.StatusInternalServerError,
			Msg:  fmt.Sprintf("service : %s not exist", name),
		})
		return
	}
	clusterName := req.Instance.ClusterName
	key := ParseInstanceKey(name, clusterName, req.GetInstance().Ip, int32(req.GetInstance().Port))

	instance, _ := parseToInstance(key, req.Instance)
	_, err := service.removeInstance(instance)
	if err != nil {
		sink.Send(&polerpc.ServerResponse{
			Code: http.StatusInternalServerError,
			Msg:  fmt.Sprintf("service : %s not exist", name),
		})
		return
	}
	sink.Send(&polerpc.ServerResponse{
		Code: http.StatusOK,
		Msg:  "success",
	})
}

func (sm *ServiceManager) createServiceIfAbsent(namespaceId, serviceName, groupName string) (*Service, error) {
	defer sm.lock.Unlock()
	sm.lock.Lock()
	if _, exist := sm.services[namespaceId]; !exist {
		sm.services[namespaceId] = make(map[string]*Service)
	}
	namespace := sm.services[namespaceId]
	name := ParseServiceName(serviceName, groupName)
	if service, exist := namespace[name]; exist {
		return service, nil
	}

	service := &Service{
		clusterLock: sync.RWMutex{},
		labelLock:   sync.RWMutex{},
		name:        ParseServiceName(serviceName, groupName),
		originService: &pojo.Service{
			ServiceName:      serviceName,
			Group:            groupName,
			ProtectThreshold: 0,
			Selector:         "",
		},
		Clusters: polerpc.ConcurrentMap{},
	}

	var sErr error

	sm.core.storageOperator.SaveService(context.Background(), service, func(err error, s *Service) {
		if err != nil {
			sErr = err
			return
		}
		sm.services[namespaceId][name] = service
	})
	if sErr != nil {
		return nil, sErr
	}
	return service, nil
}

func (sm *ServiceManager) findService(namespaceID, serviceName string) *Service {
	defer sm.lock.RUnlock()
	sm.lock.RLock()
	if services, isOk := sm.services[namespaceID]; isOk {
		if service, isOk := services[serviceName]; isOk {
			return service
		}
	}
	return nil
}

// 查询某个服务下的实例信息，不支持模糊查询的操作
func (sm *ServiceManager) queryInstances(query QueryInstance) []Instance {
	return nil
}

func (sm *ServiceManager) storeInstanceInfo(sink polerpc.RpcServerContext, service *Service, instance Instance,
	metadata InstanceMetadata) {
	originClusterName := instance.originInstance.ClusterName
	clusterName := ParseClusterName(service.name, originClusterName)
	cluster := service.Clusters.Get(clusterName)
	if cluster == nil {
		cluster = &Cluster{
			lock: sync.RWMutex{},
			name: clusterName,
			originCluster: &pojo.Cluster{
				ServiceName: service.originService.ServiceName,
				Group:       service.originService.Group,
				ClusterName: instance.originInstance.ClusterName,
				Metadata:    nil,
			},
			instanceType: utils.IF(instance.IsTemporary(), CInstanceTemporary, CInstancePersist).(CInstanceType),
			repository:   polerpc.ConcurrentMap{},
		}
		service.Clusters.Put(originClusterName, cluster)
	}
	// TODO 2021-01-30 这里是有BUG的，需要修复下，可以改为 Cluster 有一个标记信息，存储着当前Cluster对应的Instance 是持久化的还是非持久化的，但是这个信息只能在运行态下存在，不得被跟随Cluster一起持久化
	// FIXME 2021-01-31 已经修复此问题
	if t := cluster.(*Cluster).GetCInstanceType(); t != CInstanceUnKnow && (t != utils.IF(instance.IsTemporary(),
		CInstanceTemporary, CInstancePersist).(CInstanceType)) {
		sink.Send(&polerpc.ServerResponse{
			Code: http.StatusInternalServerError,
			Msg:  "A service can only be a persistent or non-persistent instance",
		})
		return
	}

	// 先存入到 KV 中，然后通过回调的方式来通知上层该做哪一些相应的处理
	sm.core.storageOperator.SaveInstance(context.Background(), &instance, func(err error, instance *Instance) {
		if err != nil {
			sink.Send(&polerpc.ServerResponse{
				Code: http.StatusInternalServerError,
				Msg:  err.Error(),
			})
			return
		}
		if _, err := service.addInstance(*instance, metadata); err != nil {
			sink.Send(&polerpc.ServerResponse{
				Code: http.StatusInternalServerError,
				Msg:  err.Error(),
			})
			return
		}

		if instance.HCType == HealthCheckByHeartbeat {
			sm.lessHolder.GrantLess(*instance)
		}
		sink.Send(&polerpc.ServerResponse{
			Code: http.StatusOK,
			Msg:  "success",
		})
	})
}

//AddressMapService 索引管理，管理着 host:port => service 的映射信息
type AddressMapService struct {
}

//Subscriber 服务订阅，使用流式推送的方式进行处理
type Subscriber struct {
	rpcCtx polerpc.RpcServerContext
}

type SubscriberManager struct {
	lock       sync.RWMutex
	subscribes map[string]polerpc.ConcurrentMap
}

func newSubscriberMgn() *SubscriberManager {
	mgn := &SubscriberManager{
		lock:       sync.RWMutex{},
		subscribes: make(map[string]polerpc.ConcurrentMap),
	}

	mgn.init()
	return mgn
}

func (mgn *SubscriberManager) init() {
	if err := notify.RegisterSubscriber(mgn); err != nil {
		// impossible to run here
		panic(err)
	}
}

func (mgn *SubscriberManager) addSubscriber(name string, subscriber *Subscriber) {

}

func (mgn *SubscriberManager) removeSubscriber(name string, subscribe *Subscriber) {

}

func (mgn *SubscriberManager) OnEvent(event notify.Event) {

}

func (mgn *SubscriberManager) SubscribeType() notify.Event {
	return &ServiceChangeEvent{}
}

func (mgn *SubscriberManager) IgnoreExpireEvent() bool {
	return false
}

func (mgn *SubscriberManager) listSubscriber() map[string]map[string]*Subscriber {
	return nil
}
