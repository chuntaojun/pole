// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pole_rpc

import (
	"fmt"
	"math/rand"
	"sync"
)

//Endpoint 实例的链接信息
type Endpoint struct {
	name string
	Key  string
	Host string
	Port int32
}

func (e Endpoint) GetKey() string {
	if e.Key == "" {
		e.Key = fmt.Sprintf("%s_%s_%d", e.name, e.Host, e.Port)
	}
	return e.Key
}

//EndpointRepository 管理实例的仓库
type EndpointRepository struct {
	rwLock        sync.RWMutex
	serviceIndex map[string]map[string]Endpoint
	serviceMap   map[string][]string
}

func NewEndpointRepository() *EndpointRepository {
	return &EndpointRepository{
		rwLock:      sync.RWMutex{},
		serviceIndex: make(map[string]map[string]Endpoint),
		serviceMap: make(map[string][]string),
	}
}

//SelectOne 选择一个服务的实例进行随机访问
func (erp *EndpointRepository) SelectOne(name string) (bool, Endpoint) {
	defer erp.rwLock.RUnlock()
	erp.rwLock.RLock()
	if v, exist := erp.serviceMap[name]; exist {
		index := rand.Int31n(int32(len(v)))
		return true, erp.serviceIndex[name][v[index]]
	}
	return false, Endpoint{}
}

//Put 为某个服务添加一个服务实例
func (erp *EndpointRepository) Put(name string, endpoint Endpoint) {
	defer erp.rwLock.Unlock()
	erp.rwLock.Lock()
	endpoint.name = name
	if _, exist := erp.serviceIndex[name]; !exist {
		erp.serviceIndex[name] = make(map[string]Endpoint)
		erp.serviceMap[name] = make([]string, 0, 0)
	}

	instance := erp.serviceIndex[name]
	if _, exist := instance[endpoint.GetKey()]; !exist {
		li := erp.serviceMap[name]
		li = append(li, endpoint.GetKey())
		erp.serviceMap[name] = li
	}
	instance[endpoint.GetKey()] = endpoint
}
