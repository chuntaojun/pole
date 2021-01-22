// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package discovery

import (
	"fmt"
	"sync"
)

type testMemoryDiscoveryStorage struct {
	rwLock    sync.RWMutex
	memoryMap map[string]*Service
}

func (tmStorage *testMemoryDiscoveryStorage) SaveService(service *Service, call func(err error, service *Service)) {
	defer tmStorage.rwLock.Unlock()
	tmStorage.rwLock.Lock()
	if _, exist := tmStorage.memoryMap[service.name]; exist {
		call(fmt.Errorf("service : %s already exist", service.name), nil)
	}
	tmStorage.memoryMap[service.name] = service
	call(nil, service)
}

func (tmStorage *testMemoryDiscoveryStorage) BatchUpdateService(services []*Service,
	call func(err error, services []*Service)) {

}

func (tmStorage *testMemoryDiscoveryStorage) RemoveService(service *Service, call func(err error, service *Service)) {

}

func (tmStorage *testMemoryDiscoveryStorage) SaveCluster(cluster *Cluster, call func(err error, cluster *Cluster)) {

}

func (tmStorage *testMemoryDiscoveryStorage) BatchUpdateCluster(clusters []*Cluster,
	call func(err error, clusters []*Cluster)) {

}

func (tmStorage *testMemoryDiscoveryStorage) RemoveCluster(cluster *Cluster, call func(err error, cluster *Cluster)) {

}

func (tmStorage *testMemoryDiscoveryStorage) SaveInstance(instance Instance, call func(err error, instance Instance)) {

}

func (tmStorage *testMemoryDiscoveryStorage) BatchUpdateInstance(instances []Instance,
	call func(err error, instances []Instance)) {

}

func (tmStorage *testMemoryDiscoveryStorage) RemoveInstance(instance Instance, call func(err error, instance Instance)) {

}
