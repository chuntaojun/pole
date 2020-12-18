//  Copyright (c) 2020, Conf-Group. All rights reserved.
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.

package plugin

import (
	"fmt"
	"sync"

	"github.com/Conf-Group/pole/common"
	"github.com/Conf-Group/pole/utils"
)

type PluginCenter struct {
	rwLock           sync.RWMutex
	pluginRepository map[string]Plugin
}

var pluginCenter *PluginCenter

func init()  {
	pluginCenter = &PluginCenter{
		rwLock:           sync.RWMutex{},
		pluginRepository: make(map[string]Plugin),
	}
}

func RegisterPlugin(cxt *common.ContextPole, p Plugin) (bool, error) {
	defer pluginCenter.rwLock.Unlock()
	pluginCenter.rwLock.Lock()
	if _, exist := pluginCenter.pluginRepository[p.Name()]; exist {
		return false, fmt.Errorf("plugin %s exist", p.Name())
	}
	pluginCenter.pluginRepository[p.Name()] = p
	utils.Go(cxt, func(ctx *common.ContextPole) {
		p.Init(ctx)
		p.Run()
	})
	return true, nil
}

func DeregisterPlugin(p Plugin) (bool, error) {
	defer pluginCenter.rwLock.Unlock()
	pluginCenter.rwLock.Lock()
	if _, exist := pluginCenter.pluginRepository[p.Name()]; exist {
		return false, fmt.Errorf("plugin %s exist", p.Name())
	}
	pluginCenter.pluginRepository[p.Name()] = p
	p.Destroy()
	return true, nil
}

func GetPluginByName(name string) Plugin {
	defer pluginCenter.rwLock.RUnlock()
	pluginCenter.rwLock.RLock()
	return pluginCenter.pluginRepository[name]
}
