// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/ptypes"
	polerpc "github.com/pole-group/pole-rpc"

	"github.com/pole-group/pole/common"
	"github.com/pole-group/pole/pojo"
	"github.com/pole-group/pole/server/cluster"
	"github.com/pole-group/pole/server/sys"
)

const (
	ConfigCommonPath = sys.APIVersion + "/config/"
	PublishConfig    = ConfigCommonPath + "publish"
	CreateConfig     = ConfigCommonPath + "create"
	ModifyConfig     = ConfigCommonPath + "modify"
	DeleteConfig     = ConfigCommonPath + "delete"
	WatchConfig      = ConfigCommonPath + "watch"
)

type ConfigServer struct {
	console *ConfConsole
	api     *ConfAPI
}

//NewConfig 创建Config模块
func NewConfig(ctx *common.ContextPole, mgn *cluster.ServerClusterManager, httpServer *gin.Engine) *ConfigServer {
	return &ConfigServer{
		console: newConfigConsole(ctx.NewSubCtx(), mgn, httpServer),
		api:     newConfAPI(ctx.NewSubCtx(), mgn),
	}
}

//Init 初始化 ConfigServer
func (c *ConfigServer) Init(ctx *common.ContextPole) {
	c.console.Init(ctx)
	c.api.Init(ctx)
}

//Shutdown 关闭 ConfigServer
func (c *ConfigServer) Shutdown() {
	if c.console != nil {
		c.console.Shutdown()
	}
	if c.api != nil {
		c.api.Shutdown()
	}
}

//ConfConsole 控制台相关接口
type ConfConsole struct {
	ctx        *common.ContextPole
	httpServer *gin.Engine
	clusterMgn *cluster.ServerClusterManager
}

//newConfigConsole 创建控制台模块
func newConfigConsole(ctx *common.ContextPole, clusterMgn *cluster.ServerClusterManager, httpServer *gin.Engine) *ConfConsole {
	return &ConfConsole{
		ctx:        ctx,
		httpServer: httpServer,
		clusterMgn: clusterMgn,
	}
}

//Init 初始化配置模块的控制台
func (cc *ConfConsole) Init(ctx *common.ContextPole) {
}

//Shutdown 控制台模块关闭
func (cc *ConfConsole) Shutdown() {
	cc.ctx.Cancel()
}

type ConfAPI struct {
	ctx        *common.ContextPole
	clusterMgn *cluster.ServerClusterManager
	server     polerpc.TransportServer
	cfgCore    *ConfigCore
}

//newConfAPI 创建Config的API模块，用于和客户端交互
func newConfAPI(ctx *common.ContextPole, clusterMgn *cluster.ServerClusterManager) *ConfAPI {
	return &ConfAPI{
		clusterMgn: clusterMgn,
		ctx:        ctx,
		cfgCore:    newConfigCore(),
	}
}

//Init 初始化Config的API模块，主要用于和客户端交互
func (ca *ConfAPI) Init(ctx *common.ContextPole) {
	server, err := polerpc.NewTransportServer(ctx, func(opt *polerpc.ServerOption) {
		opt.ConnectType = polerpc.ConnectTypeRSocket
		opt.Label = "CONF-CONFIG"
		opt.Port = ca.clusterMgn.GetSelf().GetExtensionPort(cluster.ConfigPort)
		opt.OpenTSL = sys.GetEnvHolder().OpenSSL
	})
	if err != nil {
		panic(err)
	}
	ca.server = server

	ca.server.RegisterRequestHandler(PublishConfig, ca.publishConfig)
	ca.server.RegisterRequestHandler(CreateConfig, ca.createConfig)
	ca.server.RegisterRequestHandler(ModifyConfig, ca.modifyConfig)
	ca.server.RegisterRequestHandler(DeleteConfig, ca.deleteConfig)
	ca.server.RegisterChannelRequestHandler(WatchConfig, ca.watch)
}

//publishConfig 允许配置文件可以被客户端访问查询
func (ca *ConfAPI) publishConfig(cxt context.Context, rpcCtx polerpc.RpcServerContext) {
	req := rpcCtx.GetReq()
	cfgReq := new(pojo.ConfigRequest)
	if err := ptypes.UnmarshalAny(req.Body, cfgReq); err != nil {
		_ = rpcCtx.Send(&polerpc.ServerResponse{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}
	ca.cfgCore.operateConfig(OpForPublishConfig, cfgReq, rpcCtx)
}

//createConfig 创建一个配置文件
func (ca *ConfAPI) createConfig(cxt context.Context, rpcCtx polerpc.RpcServerContext) {
	req := rpcCtx.GetReq()
	cfgReq := new(pojo.ConfigRequest)
	if err := ptypes.UnmarshalAny(req.Body, cfgReq); err != nil {
		_ = rpcCtx.Send(&polerpc.ServerResponse{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}
	ca.cfgCore.operateConfig(OpForCreateConfig, cfgReq, rpcCtx)
}

//modifyConfig 修改配置文件
func (ca *ConfAPI) modifyConfig(cxt context.Context, rpcCtx polerpc.RpcServerContext) {
	req := rpcCtx.GetReq()
	cfgReq := new(pojo.ConfigRequest)
	if err := ptypes.UnmarshalAny(req.Body, cfgReq); err != nil {
		_ = rpcCtx.Send(&polerpc.ServerResponse{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}
	ca.cfgCore.operateConfig(OpForModifyConfig, cfgReq, rpcCtx)
}

//deleteConfig 删除配置文件
func (ca *ConfAPI) deleteConfig(cxt context.Context, rpcCtx polerpc.RpcServerContext) {
	req := rpcCtx.GetReq()
	cfgReq := new(pojo.ConfigRequest)
	if err := ptypes.UnmarshalAny(req.Body, cfgReq); err != nil {
		_ = rpcCtx.Send(&polerpc.ServerResponse{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}
	ca.cfgCore.operateConfig(OpForDeleteConfig, cfgReq, rpcCtx)
}

//watch 监听一个配置文件
func (ca *ConfAPI) watch(cxt context.Context, rpcCtx polerpc.RpcServerContext) {
	req := rpcCtx.GetReq()
	cfgReq := new(pojo.ConfigWatchRequest)
	if err := ptypes.UnmarshalAny(req.Body, cfgReq); err != nil {
		_ = rpcCtx.Send(&polerpc.ServerResponse{
			Code: -1,
			Msg:  err.Error(),
		})
		return
	}
	ca.cfgCore.listenConfig(cfgReq, rpcCtx)
}

//Shutdown 关闭API模块
func (ca *ConfAPI) Shutdown() {
	ca.ctx.Cancel()
}
