// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	polerpc "github.com/pole-group/pole-rpc"

	"github.com/pole-group/pole/common"
	"github.com/pole-group/pole/server/cluster"
	"github.com/pole-group/pole/server/config"
	"github.com/pole-group/pole/server/discovery"
	"github.com/pole-group/pole/server/sys"
)

var (
	PoleLogger = polerpc.NewTestLogger("pole-server")
)

type Pole struct {
	ctx             *common.ContextPole
	clusterMgn      *cluster.ServerClusterManager
	configModule    *config.ConfigServer
	discoveryModule *discovery.DiscoveryServer
	httpEngine      *gin.Engine
}

func startPoleServer() {
	n := new(Pole)
	n.Start()
}

func (n *Pole) Start() {
	n.ctx = common.NewCtxPole()
	n.httpEngine = gin.New()
	sys.InitConf()
	n.initServer()
	n.initConfig()
	n.initDiscovery()

	// 只有上述的动作都做好了，才能真正的启动 http-server
	if err := n.httpEngine.Run(fmt.Sprintf(":%d", sys.GetEnvHolder().ServerPort)); err != nil {
		panic(err)
	}
}

func (n *Pole) initServer() {
	PoleLogger.Info("start pole-server cluster manager")
	n.clusterMgn = cluster.NewServerClusterManager()
	n.clusterMgn.RegisterHttpHandler(n.httpEngine)
}

func (n *Pole) initConfig() {
	PoleLogger.Info("start pole-server config module")
	n.configModule = config.NewConfig(n.ctx.NewSubCtx(), n.clusterMgn, n.httpEngine)
}

func (n *Pole) initDiscovery() {
	PoleLogger.Info("start pole-server discovery module")
	n.discoveryModule = discovery.NewDiscoveryServer(n.ctx.NewSubCtx(), n.clusterMgn, n.httpEngine)
}
