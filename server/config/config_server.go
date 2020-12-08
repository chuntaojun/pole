// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/Conf-Group/pole/common"
	"github.com/Conf-Group/pole/server/sys"
	"github.com/Conf-Group/pole/transport/rsocket"
)

type ConfigServer struct {
	console *ConfConsole
	api     *ConfAPI
}

func NewConfig(cfg sys.Properties, ctx context.Context, httpServer *gin.Engine) *ConfigServer {
	return &ConfigServer{
		console: newConfigConsole(cfg, httpServer),
		api:     newConfAPI(cfg, ctx),
	}
}

func (c *ConfigServer) Init(ctx *common.ContextPole) {
	c.console.Init(ctx)
	c.api.Init(ctx)
}

func (c *ConfigServer) Shutdown() {
	if c.console != nil {
		c.console.Shutdown()
	}
	if c.api != nil {
		c.api.Shutdown()
	}
}

type ConfConsole struct {
	httpServer *gin.Engine
}

func newConfigConsole(cfg sys.Properties, httpServer *gin.Engine) *ConfConsole {
	return &ConfConsole{
		httpServer: httpServer,
	}
}

func (cc *ConfConsole) Init(ctx *common.ContextPole) {
}

func (cc *ConfConsole) Shutdown() {

}

type ConfAPI struct {
	server *rsocket.RSocketServer
	ctx	context.Context
}

func newConfAPI(cfg sys.Properties, ctx	context.Context) *ConfAPI {
	subCtx, _ := context.WithCancel(ctx)
	return &ConfAPI{
		server: rsocket.NewRSocketServer(subCtx, "CONF-CONFIG", cfg.ConfigPort, cfg.OpenSSL),
		ctx:    subCtx,
	}
}

func (ca *ConfAPI) Init(ctx *common.ContextPole) {

}

func (ca *ConfAPI) Shutdown() {

}
