// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package discovery

import (
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/Conf-Group/pole/utils"
)

const (
	AND = "and"
	OR  = "or"
)

type ServiceManager struct {
	lock     sync.RWMutex
	services map[string]map[string]*Service
}

func (sm *ServiceManager) FindService(namespaceID, serviceName string) *Service {
	defer sm.lock.RUnlock()
	sm.lock.RLock()
	if services, isOk := sm.services[namespaceID]; isOk {
		if service, isOk := services[serviceName]; isOk {
			return service
		}
	}
	return nil
}

func (sm *ServiceManager) Select(namespaceID, cServiceName, clusterName, clientHost, tServiceName string) []Instance {
	return nil
}

type Service struct {
	clusterLock sync.RWMutex
	labelLock   sync.RWMutex

	ServiceName string
	Clusters    map[string]*Cluster
	// consumer.label.{key}=provider.label.{key}
	LabelRules map[string]string
}

func (s *Service) FindInstance(clusterName, key string) Instance {
	return s.Clusters[clusterName].FindInstance(key)
}

type Cluster struct {
	tmpLock    sync.RWMutex
	perLock    sync.RWMutex
	Name       string
	Temporary  map[string]Instance
	Persistent map[string]Instance
	metadata   atomic.Value
}

func (c *Cluster) FindInstance(key string) Instance {
	if instance, isOk := c.Temporary[key]; isOk {
		return instance
	}
	return c.Persistent[key]
}

func (c *Cluster) AddInstance(isTemp bool, instance Instance) {
	lock := utils.IF(isTemp, &c.tmpLock, &c.perLock).(*sync.RWMutex)
	instances := utils.IF(isTemp, &c.Temporary, &c.Persistent).(map[string]Instance)
	defer lock.Unlock()
	lock.Lock()
	instances[instance.GetKey()] = instance
}

func (c *Cluster) RemoveInstance(isTemp bool, instance Instance) {
	lock := utils.IF(isTemp, &c.tmpLock, &c.perLock).(*sync.RWMutex)
	instances := utils.IF(isTemp, &c.Temporary, &c.Persistent).(map[string]Instance)
	defer lock.Unlock()
	lock.Lock()
	delete(instances, instance.GetKey())
}

func (c *Cluster) Seek(isTemp bool, consumer func(instance Instance)) {
	lock := utils.IF(isTemp, &c.tmpLock, &c.perLock).(*sync.RWMutex)
	defer lock.RUnlock()
	lock.RLock()
	for _, v := range utils.IF(isTemp, &c.Temporary, &c.Persistent).(map[string]Instance) {
		consumer(v)
	}
}

func (c *Cluster) UpdateMetadata(metadata map[string]string) {
	c.metadata.Store(metadata)
}

func (c *Cluster) GetMetadata() map[string]string {
	return c.metadata.Load().(map[string]string)
}

type Instance struct {
	key       string
	host      string
	port      int64
	weight    float64
	enabled   bool
	healthy   bool
	temporary bool
	metadata  map[string]string
}

func (i Instance) GetIP() string {
	return i.host
}

func (i Instance) GetPort() int64 {
	return i.port
}

func (i Instance) SetWeight(wright float64) {
	i.weight = wright
}

func (i Instance) GetWeight() float64 {
	return i.weight
}

func (i Instance) SetEnabled(enabled bool) {
	i.enabled = enabled
}

func (i Instance) IsTemporary() bool {
	return i.temporary
}

func (i Instance) IsEnabled() bool {
	return i.enabled
}

func (i Instance) SetHealthy(healthy bool) {
	i.healthy = healthy
}

func (i Instance) IsHealthy() bool {
	return i.healthy
}

func (i Instance) UpdateMetadata(newMetadata map[string]string) {
	i.metadata = newMetadata
}

func (i Instance) GetMetadata(key string) string {
	return i.metadata[key]
}

func (i Instance) GetKey() string {
	if i.key == "" {
		i.key = i.host + ":" + strconv.FormatInt(i.port, 10)
	}
	return i.key
}
