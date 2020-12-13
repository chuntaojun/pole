// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package healthcheck

import (
	"github.com/Conf-Group/pole/common"
	"github.com/Conf-Group/pole/plugin"
)

type HealthCheckPlugin interface {
	plugin.Plugin
}

type HttpBeatHealthCheckPlugin struct {


}

func (h *HttpBeatHealthCheckPlugin) Name() string {
	return "health_check_http_beat"
}

func (h *HttpBeatHealthCheckPlugin)  Init(ctx *common.ContextPole) {

}

func (h *HttpBeatHealthCheckPlugin) Run() {

}

func (h *HttpBeatHealthCheckPlugin) Destroy() {

}

type TcpHealthCheckPlugin struct {

}


func (h *TcpHealthCheckPlugin) Name() string {
	return "health_check_tcp"
}

func (h *TcpHealthCheckPlugin)  Init(ctx *common.ContextPole) {

}

func (h *TcpHealthCheckPlugin) Run() {

}

func (h *TcpHealthCheckPlugin) Destroy() {

}