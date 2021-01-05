// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package discovery

import (
	"fmt"
	"sync"

	"github.com/Conf-Group/pole/pojo"
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
	lock     sync.RWMutex
	services map[string]map[string]*Service
}

func (sm *ServiceManager) addInstance(req *pojo.InstanceRegister) *pojo.RestResult {
	namespaceId := req.NamespaceId
	serviceName := req.Instance.ServiceName
	groupName := req.Instance.Group

	name := fmt.Sprintf("%s@@%s", serviceName, groupName)
	sm.createServiceIfAbsent(namespaceId, name)

	clusterName := req.Instance.ClusterName
	key := fmt.Sprintf("%s#%s#%s#%s#%d", serviceName, groupName, clusterName, req.GetInstance().Ip,
		req.GetInstance().Port)

	instance, metadata := parseToInstance(key, req.Instance)

	cluster := sm.services[namespaceId][name].Clusters[clusterName]
	if cluster == nil {
		return &pojo.RestResult{
			Code: -1,
			Msg:  fmt.Sprintf("Cluster : %s not exist in service : %s, namespace : %s", clusterName, name, namespaceId),
		}
	}

	if ri := cluster.FindRandom(); !isEmptyInstance(ri) && ri.temporary != instance.temporary {
		return &pojo.RestResult{
			Code: 0,
			Msg:  "A service can only be a persistent or non-persistent instance",
		}
	}

	return sm.storeInstanceInfo(cluster, instance, metadata)
}

func (sm *ServiceManager) createServiceIfAbsent(namespaceId, serviceName string) {
	sm.lock.Lock()
	if _, exist := sm.services[namespaceId]; !exist {
		sm.services[namespaceId] = make(map[string]*Service)
		if _, exist := sm.services[namespaceId][serviceName]; !exist {
			sm.services[namespaceId][serviceName] = &Service{
				clusterLock: sync.RWMutex{},
				labelLock:   sync.RWMutex{},
				ServiceName: serviceName,
				Clusters:    make(map[string]*Cluster),
			}
		}
	}

	sm.lock.Unlock()
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

func (sm *ServiceManager) storeInstanceInfo(cluster *Cluster, instance Instance, metadata InstanceMetadata) *pojo.RestResult {
	cluster.AddInstance(instance, metadata)
	return nil
}

