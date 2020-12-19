// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package healthcheck

import (
	"fmt"

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

type HttpCodeCheckTask struct {
	Instance   pojo.Instance
	CheckPath  string
	ExpectCode int32
}

type HeartbeatCheckTask struct {
	Instance pojo.Instance
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
	_, _ = plugin.RegisterPlugin(ctx, &ConnectionBeatHealthCheckPlugin{})
	_, _ = plugin.RegisterPlugin(ctx, &HttpCodeHealthCheckPlugin{})
	_, _ = plugin.RegisterPlugin(ctx, &CustomHealthCheckPlugin{})
}

func (hcm *HealthCheckManager) RegisterHealthCheckTask(task HealthCheckTask) (bool, error) {
	var checker HealthCheckPlugin
	switch task.(type) {
	case HttpCodeCheckTask:
		checker = plugin.GetPluginByName(HealthCheck_Http).(*HttpCodeHealthCheckPlugin)
	case TcpHealthCheckTask:
		checker = plugin.GetPluginByName(HealthCheck_Tcp).(*TcpHealthCheckPlugin)
	case CustomerHealthCheckTask:
		checker = plugin.GetPluginByName(HealthCheck_Customer).(*CustomHealthCheckPlugin)
	case HeartbeatCheckTask:
		// 走 Heartbeat 的模式
		checker = plugin.GetPluginByName(HealthCheck_Heartbeat).(*ConnectionBeatHealthCheckPlugin)
	default:
		return false, fmt.Errorf("unsupport this check task : %s", task)
	}
	return checker.AddTask(task)
}

func (hcm *HealthCheckManager) Shutdown() {

}
