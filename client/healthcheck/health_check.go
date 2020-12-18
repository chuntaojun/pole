// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package healthcheck

import (
	"github.com/Conf-Group/pole/common"
	"github.com/Conf-Group/pole/plugin"
	"github.com/Conf-Group/pole/pojo"
)

const (
	HealthCheck_Customer  = "customer_health_check"
	HealthCheck_Tcp       = "tcp_health_check"
	HealthCheck_Http      = "http_health_check"
	HealthCheck_Heartbeat = "heartbeat_health_check" // default mode
)

type HealthCheckPlugin interface {
	plugin.Plugin

	AddTask(task HealthCheckTask) (bool, error)

	RemoveTask(task HealthCheckTask) (bool, error)
}

type HealthCheckTask interface {
}

type HttpHealthCheckTask struct {
	Instance   pojo.Instance
	CheckPath  string
	ExpectCode int32
}

type TcpHealthCheckTask struct {
	Instance pojo.Instance
}

type CustomerHealthCheckTask struct {
	Instance pojo.Instance
	F        CustomerHealthChecker
}

type HealthCheckManager struct {
	checkPlugins map[string]HealthCheckPlugin
}

func (hcm *HealthCheckManager) Init(ctx *common.ContextPole) {
	_, _ = plugin.RegisterPlugin(ctx, &TcpHealthCheckPlugin{})
	_, _ = plugin.RegisterPlugin(ctx, &HttpBeatHealthCheckPlugin{})
	_, _ = plugin.RegisterPlugin(ctx, &CustomHealthCheckPlugin{})
}

func (hcm *HealthCheckManager) RegisterHealthCheckTask(task HealthCheckTask) (bool, error) {
	var checker HealthCheckPlugin
	switch task.(type) {
	case HttpHealthCheckTask:
		checker = plugin.GetPluginByName(HealthCheck_Http).(*HttpBeatHealthCheckPlugin)
	case TcpHealthCheckTask:
		checker = plugin.GetPluginByName(HealthCheck_Tcp).(*TcpHealthCheckPlugin)
	case CustomerHealthCheckTask:
		checker = plugin.GetPluginByName(HealthCheck_Customer).(*CustomHealthCheckPlugin)
	default:
		// 走 Heartbeat 的模式
		checker = plugin.GetPluginByName(HealthCheck_Heartbeat).(*HttpBeatHealthCheckPlugin)
	}
	return checker.AddTask(task)
}

func (hcm *HealthCheckManager) Shutdown() {

}
