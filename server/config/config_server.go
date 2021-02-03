// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"github.com/gin-gonic/gin"
	polerpc "github.com/pole-group/pole-rpc"

	"github.com/pole-group/pole/common"
	"github.com/pole-group/pole/server/cluster"
	"github.com/pole-group/pole/server/sys"
)

type ConfigServer struct {
	console *ConfConsole
	api     *ConfAPI
}

func NewConfig(ctx *common.ContextPole, mgn *cluster.ServerClusterManager, httpServer *gin.Engine) *ConfigServer {
	return &ConfigServer{
		console: newConfigConsole(ctx.NewSubCtx(), mgn, httpServer),
		api:     newConfAPI(ctx.NewSubCtx(), mgn),
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
	ctx        *common.ContextPole
	httpServer *gin.Engine
	clusterMgn *cluster.ServerClusterManager
}

func newConfigConsole(ctx *common.ContextPole, clusterMgn *cluster.ServerClusterManager, httpServer *gin.Engine) *ConfConsole {
	return &ConfConsole{
		ctx:        ctx,
		httpServer: httpServer,
		clusterMgn: clusterMgn,
	}
}

func (cc *ConfConsole) Init(ctx *common.ContextPole) {
}

func (cc *ConfConsole) Shutdown() {
	cc.ctx.Cancel()
}

type ConfAPI struct {
	ctx        *common.ContextPole
	clusterMgn *cluster.ServerClusterManager
	server     polerpc.TransportServer
}

func newConfAPI(ctx *common.ContextPole, clusterMgn *cluster.ServerClusterManager) *ConfAPI {
	return &ConfAPI{
		clusterMgn: clusterMgn,
		ctx:        ctx,
	}
}

func (ca *ConfAPI) Init(ctx *common.ContextPole) {
	ca.server = polerpc.NewRSocketServer(ctx, "CONF-CONFIG", ca.clusterMgn.GetSelf().GetExtensionPort(cluster.ConfigPort),
		sys.GetEnvHolder().OpenSSL)
}

func (ca *ConfAPI) Shutdown() {
	ca.ctx.Cancel()
}
