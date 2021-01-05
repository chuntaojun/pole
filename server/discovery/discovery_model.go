// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package discovery

import (
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/Conf-Group/pole/pojo"
)

type ServiceMetadata struct {
	ServiceName string
	metadata    map[string]string
	// consumer.label.{key}=provider.label.{key}
	LabelRules map[string]string
}

type Service struct {
	clusterLock sync.RWMutex
	labelLock   sync.RWMutex
	ServiceName string
	Clusters    map[string]*Cluster
}

func (s *Service) FindInstance(clusterName, key string) (Instance, error) {
	return s.Clusters[clusterName].FindInstance(key)
}

type ClusterMetadata struct {
	ClusterName string
	metadata    map[string]string
}

type Cluster struct {
	lock      sync.RWMutex
	Name      string
	Instances map[string]Instance
	metadata  atomic.Value
}

var emptyInstance Instance = Instance{
	key:       "",
	host:      "",
	port:      -1,
	weight:    -1,
	enabled:   false,
	healthy:   false,
	temporary: false,
}

func (c *Cluster) FindInstance(key string) (Instance, error) {
	if instance, isOk := c.Instances[key]; isOk {
		return instance, nil
	}
	return emptyInstance, fmt.Errorf("can't find instance by : %s", key)
}

func (c *Cluster) AddInstance(instance Instance, metadata InstanceMetadata) {
	defer c.lock.Unlock()
	c.lock.Lock()
	c.Instances[instance.GetKey()] = instance
}

func (c *Cluster) RemoveInstance(instance Instance) {
	defer c.lock.Unlock()
	c.lock.Lock()
	delete(c.Instances, instance.GetKey())
}

func (c *Cluster) Seek(consumer func(instance Instance)) {
	defer c.lock.RUnlock()
	c.lock.RLock()
	for _, v := range c.Instances {
		consumer(v)
	}
}

func (c *Cluster) FindRandom() Instance {
	defer c.lock.RUnlock()
	c.lock.RLock()
	for _, v := range c.Instances {
		return v
	}
	return emptyInstance
}

func (c *Cluster) UpdateMetadata(metadata map[string]string) {
	c.metadata.Store(metadata)
}

func (c *Cluster) GetMetadata() map[string]string {
	return c.metadata.Load().(map[string]string)
}


type InstanceMetadata struct {
	key      string
	metadata map[string]string
}

func (i InstanceMetadata) GetKey() string {
	return i.key
}

func (i InstanceMetadata) UpdateMetadata(newMetadata map[string]string) {
	i.metadata = newMetadata
}

func (i InstanceMetadata) GetMetadata(key string) string {
	return i.metadata[key]
}

func isEmptyInstance(i Instance) bool {
	return i.key == "" && i.host == "" && i.port == -1
}

type Instance struct {
	key       string
	host      string
	port      int64
	weight    float64
	enabled   bool
	healthy   bool
	temporary bool
}

func parseToInstance(key string, i *pojo.Instance) (Instance, InstanceMetadata) {
	instance := Instance{
		key:       key,
		host:      i.Ip,
		port:      i.Port,
		weight:    i.Weight,
		enabled:   i.Enabled,
		healthy:   true,
		temporary: i.Ephemeral,
	}

	metadata := InstanceMetadata{
		key:      key,
		metadata: i.Metadata,
	}

	return instance, metadata
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

func (i Instance) GetKey() string {
	if i.key == "" {
		i.key = i.host + ":" + strconv.FormatInt(i.port, 10)
	}
	return i.key
}

