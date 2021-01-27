// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package discovery

import (
	"context"
	"fmt"
	"sync"

	polerpc "github.com/pole-group/pole-rpc"

	"github.com/pole-group/pole/pojo"
)

const (
	AND = "and"
	OR  = "or"
)

type DiscoveryCore struct {
	serviceMgn      *ServiceManager
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

	name := fmt.Sprintf("%s@@%s", serviceName, groupName)
	if err := sm.createServiceIfAbsent(namespaceId, serviceName, groupName); err != nil {
		sink.Send(&polerpc.ServerResponse{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}

	clusterName := req.Instance.ClusterName
	key := fmt.Sprintf("%s#%s#%s#%s#%d", serviceName, groupName, clusterName, req.GetInstance().Ip,
		req.GetInstance().Port)

	instance, metadata := parseToInstance(key, req.Instance)

	cluster := sm.services[namespaceId][name].Clusters[clusterName]
	if cluster == nil {
		sink.Send(&polerpc.ServerResponse{
			Code: -1,
			Msg:  fmt.Sprintf("Cluster : %s not exist in service : %s, namespace : %s", clusterName, name, namespaceId),
		})
		return
	}

	if ri := cluster.FindRandom(); !isEmptyInstance(ri) && ri.IsTemporary() != instance.IsTemporary() {
		sink.Send(&polerpc.ServerResponse{
			Code: -1,
			Msg:  "A service can only be a persistent or non-persistent instance",
		})
	} else {
		sm.storeInstanceInfo(sink, cluster, instance, metadata)
	}
}

func (sm *ServiceManager) createServiceIfAbsent(namespaceId, serviceName, groupName string) error {
	defer sm.lock.Unlock()
	sm.lock.Lock()
	if _, exist := sm.services[namespaceId]; !exist {
		sm.services[namespaceId] = make(map[string]*Service)
		if _, exist := sm.services[namespaceId][serviceName]; !exist {

			service := &Service{
				clusterLock: sync.RWMutex{},
				labelLock:   sync.RWMutex{},
				name:        fmt.Sprintf("%s@@%s", serviceName, groupName),
				originService: &pojo.Service{
					ServiceName:      serviceName,
					Group:            groupName,
					ProtectThreshold: 0,
					Selector:         "",
				},
				Clusters: make(map[string]*Cluster),
			}

			sm.core.storageOperator.SaveService(context.Background(), service, func(err error, s *Service) {
				sm.services[namespaceId][serviceName] = service
			})
		}
	}

	return nil
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

// 查询某个服务下的实例信息，支持模糊查询的操作
func (sm *ServiceManager) SelectInstances(query QueryInstance) []Instance {
	return nil
}

func (sm *ServiceManager) storeInstanceInfo(sink polerpc.RpcServerContext, cluster *Cluster, instance Instance,
	metadata InstanceMetadata) {
	sm.core.storageOperator.SaveInstance(context.Background(), &instance, func(err error, instance *Instance) {
		if err != nil {
			sink.Send(&polerpc.ServerResponse{
				Code: 0,
				Msg:  err.Error(),
			})
			return
		}
		cluster.AddInstance(*instance, metadata)

		if instance.HCType == HealthCheckByHeartbeat {
			sm.lessHolder.GrantLess(*instance)
		}

		sink.Send(&polerpc.ServerResponse{
			Code: 0,
			Msg:  "success",
		})
	})
}
