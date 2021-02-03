// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package discovery

import (
	"fmt"
	"sync"

	polerpc "github.com/pole-group/pole-rpc"

	"github.com/pole-group/pole/pojo"
	"github.com/pole-group/pole/utils"
)

const (
	ServiceLink  = "@@"
	ClusterLink  = "$$"
	InstanceLink = "=>"
)

func ParseServiceName(name, group string) string {
	return fmt.Sprintf("%s"+ServiceLink+"%s", name, group)
}

func ParseClusterName(serviceName, clusterName string) string {
	return fmt.Sprintf("%s"+ClusterLink+"%s", serviceName, clusterName)
}

func ParseServiceClusterList(name string) string {
	return name + "_cluster_list"
}

func ParseServiceClusterInstanceList(name string) string {
	return name + "_instance_list"
}

func ParseInstanceKey(serviceName, clusterName, host string, port int32) string {
	return fmt.Sprintf("%s"+InstanceLink+"%s:%d", ParseClusterName(serviceName, clusterName), host, port)
}

type InstanceHealthCheckType int8

const (
	HealthCheckByHeartbeat InstanceHealthCheckType = iota
	HealthCheckByAgent
)

type Service struct {
	clusterLock   sync.RWMutex
	labelLock     sync.RWMutex
	name          string
	originService *pojo.Service
	Clusters      polerpc.ConcurrentMap // 仅仅是为了读写分离
}

func (s *Service) addInstance(instance Instance, metadata InstanceMetadata) (bool, error) {
	cluster := s.Clusters.Get(instance.originInstance.ClusterName).(*Cluster)
	cluster.repository.Put(instance.key, instance)
	return true, nil
}

func (s *Service) removeInstance(instance Instance) (bool, error) {
	cluster := s.Clusters.Get(instance.originInstance.ClusterName).(*Cluster)
	cluster.repository.Remove(instance.key)
	return true, nil
}

func (s *Service) findAllInstance() ([]Instance, error) {
	total := 0
	s.Clusters.ForEach(func(k, v interface{}) {
		total += v.(*Cluster).repository.Size()
	})
	allInstances := make([]Instance, total, total)
	index := 0
	s.Clusters.ForEach(func(k, v interface{}) {
		cluster := v.(*Cluster)
		cluster.repository.ForEach(func(k, v interface{}) {
			allInstances[index] = v.(Instance)
			index++
		})
	})
	return allInstances, nil
}

func (s *Service) findOneInstance(clusterName, key string) (Instance, error) {
	return s.Clusters.Get(clusterName).(*Cluster).FindInstance(key)
}

type ClusterMetadata struct {
	ClusterName string
	metadata    map[string]string
}

type CInstanceType int8

const (
	CInstanceUnKnow CInstanceType = iota
	CInstancePersist
	CInstanceTemporary
)

type Cluster struct {
	lock          sync.RWMutex
	name          string
	originCluster *pojo.Cluster
	instanceType  CInstanceType
	repository    polerpc.ConcurrentMap
}

var emptyInstance Instance = Instance{
	key: "",
	originInstance: &pojo.Instance{
		Ip:   "",
		Port: -1,
	},
	health: false,
}

func (c *Cluster) GetCInstanceType() CInstanceType {
	return c.instanceType
}

func (c *Cluster) FindInstance(key string) (Instance, error) {
	instance := c.repository.Get(key)
	if instance != nil {
		return instance.(Instance), nil
	}
	//TODO 需要直接读取KV的方式再一次读取数据，判断是否存在该实例信息
	return emptyInstance, fmt.Errorf("can't find instance by : %s", key)
}

func (c *Cluster) addInstance(instance Instance, metadata InstanceMetadata) (bool, error) {
	if c.instanceType == CInstanceUnKnow {
		c.lock.Lock()
		if c.instanceType == CInstanceUnKnow {
			c.instanceType = utils.IF(instance.IsTemporary(), CInstanceTemporary, CInstancePersist).(CInstanceType)
		} else {
			if c.instanceType != utils.IF(instance.IsTemporary(), CInstanceTemporary,
				CInstancePersist).(CInstanceType) {
				c.lock.Unlock()
				return false, fmt.Errorf("A service can only be a persistent or non-persistent instance")
			}
		}
		c.lock.Unlock()
	}
	c.repository.Put(instance.GetKey(), instance)
	return true, nil
}

func (c *Cluster) RemoveInstance(instance Instance) {
	c.repository.Remove(instance.GetKey())
}

func (c *Cluster) Seek(consumer func(instance Instance)) {
	c.repository.ForEach(func(k, v interface{}) {
		consumer(v.(Instance))
	})
}

func (c *Cluster) UpdateMetadata(metadata map[string]string) {
	defer c.lock.Unlock()
	c.lock.Lock()
	c.originCluster.Metadata.Metadata = metadata
}

func (c *Cluster) GetMetadata() map[string]string {
	return c.originCluster.Metadata.Metadata
}

type InstanceMetadata struct {
	key            string
	originMetadata *pojo.InstanceMetadata
}

func (i InstanceMetadata) GetKey() string {
	return i.key
}

func (i InstanceMetadata) UpdateMetadata(newMetadata map[string]string) {
	i.originMetadata.Metadata = newMetadata
}

func (i InstanceMetadata) GetMetadata(key string) string {
	return i.originMetadata.Metadata[key]
}

func isEmptyInstance(i Instance) bool {
	return i.key == "" && i.originInstance.Ip == "" && i.originInstance.Port == -1
}

type Instance struct {
	key            string
	health         bool
	originInstance *pojo.Instance
	HCType         InstanceHealthCheckType
}

func parseToInstance(key string, i *pojo.Instance) (Instance, InstanceMetadata) {
	instance := Instance{
		key:            key,
		originInstance: i,
		HCType: utils.IF(i.HealthCheckType == pojo.CheckType_HeartBeat, HealthCheckByHeartbeat,
			HealthCheckByAgent).(InstanceHealthCheckType),
	}

	metadata := InstanceMetadata{
		key:            key,
		originMetadata: i.Metadata,
	}

	return instance, metadata
}

func (i Instance) GetIP() string {
	return i.originInstance.Ip
}

func (i Instance) GetPort() int64 {
	return i.originInstance.Port
}

func (i Instance) SetWeight(wright float64) {
	i.originInstance.Weight = wright
}

func (i Instance) GetWeight() float64 {
	return i.originInstance.Weight
}

func (i Instance) SetEnabled(enabled bool) {
	i.originInstance.Enabled = enabled
}

func (i Instance) IsTemporary() bool {
	return i.originInstance.Ephemeral
}

func (i Instance) IsEnabled() bool {
	return i.originInstance.Enabled
}

func (i Instance) SetHealthy(healthy bool) {
	i.health = healthy
}

func (i Instance) IsHealthy() bool {
	return i.health
}

func (i Instance) GetKey() string {
	if i.key == "" {
		panic(fmt.Errorf("instance key must be init when create"))
	}
	return i.key
}
