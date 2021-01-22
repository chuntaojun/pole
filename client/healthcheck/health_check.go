// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package healthcheck

import (
	"fmt"

	"github.com/pole-group/pole/common"
	"github.com/pole-group/pole/plugin"
	"github.com/pole-group/pole/pojo"
)

const (
	// 处于agent模式的实例健康检查模式
	healthcheckCustomer = "customer_health_check"
	HealthcheckTcp      = "tcp_health_check"
	HealthcheckHttp     = "http_health_check"
	// 默认健康检查模式，采用心跳包检查的方式
	HealthcheckHeartbeat = "heartbeat_health_check" // default mode
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
	ExpectCode int
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
		checker = plugin.GetPluginByName(HealthcheckHttp).(*HttpCodeHealthCheckPlugin)
	case TcpHealthCheckTask:
		checker = plugin.GetPluginByName(HealthcheckTcp).(*TcpHealthCheckPlugin)
	case CustomerHealthCheckTask:
		checker = plugin.GetPluginByName(healthcheckCustomer).(*CustomHealthCheckPlugin)
	case HeartbeatCheckTask:
		// 走 Heartbeat 的模式
		checker = plugin.GetPluginByName(HealthcheckHeartbeat).(*ConnectionBeatHealthCheckPlugin)
	default:
		return false, fmt.Errorf("unsupport this check task : %s", task)
	}
	return checker.AddTask(task)
}

func (hcm *HealthCheckManager) Shutdown() {

}
