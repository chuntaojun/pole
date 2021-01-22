// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package client

type RunModeType int8

type TransportType int8

const (
	SdkRunMode RunModeType = iota
	AgentRunMode

	HttpTransport TransportType = iota
	RSocketTransport
)

type Config struct {
	Token         string        `yaml:"token"`          // token，用于认证识别
	NamespaceId   string        `yaml:"namespaceId"`    // 命名空间
	DiscoveryAddr []string      `yaml:"discovery-addr"` // 注册中心的地址
	ConfigAddr    []string      `yaml:"config-addr"`    // 配置中心的地址
	ModeType      RunModeType   `yaml:"mode-type"`      // SDK 的运行模式, 默认为SdkRunMode
	TransportMode TransportType `yaml:"transport-type"` // 通行模式选择, 默认为RSocketTransport
}
