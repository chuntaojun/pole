// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package discovery

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/ptypes"
	"github.com/jjeffcaii/reactor-go/mono"
	pole_rpc "github.com/pole-group/pole-rpc"

	"github.com/pole-group/pole/common"
	"github.com/pole-group/pole/pojo"
	"github.com/pole-group/pole/server/sys"
)

const (
	InstanceCommonPath     = sys.APIVersion + "/discovery/instance/"
	InstanceRegister       = InstanceCommonPath + "register"
	InstanceDeregister     = InstanceCommonPath + "deregister"
	InstanceMetadataUpdate = InstanceCommonPath + "metadataUpdate"
	InstanceHeartbeat      = InstanceCommonPath + "heartbeat"
	InstanceSelect         = InstanceCommonPath + "select"
	InstanceDisabled       = InstanceCommonPath + "disabled"
)

type DiscoveryServer struct {
	console *DiscoveryConsole
	api     *DiscoverySdkAPI
}

func NewDiscoveryServer(cfg sys.Properties, ctx *common.ContextPole, httpServer *gin.Engine) *DiscoveryServer {
	server := &DiscoveryServer{
		console: newDiscoveryConsole(cfg, httpServer),
		api:     newDiscoverySdkAPI(cfg),
	}

	server.Init(ctx)
	return server
}

func (d *DiscoveryServer) Init(ctx *common.ContextPole) {
	d.console.Init(ctx)
	d.api.Init(ctx)
	if ok, err := InitDiscoveryStorage(); !ok || err != nil {
		panic(fmt.Errorf("init discovery storage plugin failed! err : %#v", err))
	}
}

func (d *DiscoveryServer) Shutdown() {
	d.console.Shutdown()
	d.api.Shutdown()
}

// DiscoveryConsole 用于控制台
type DiscoveryConsole struct {
	httpServer *gin.Engine
	core       *DiscoveryCore
	ctx        *common.ContextPole
}

func newDiscoveryConsole(cfg sys.Properties, httpServer *gin.Engine) *DiscoveryConsole {
	return &DiscoveryConsole{
		httpServer: httpServer,
	}
}

func (nc *DiscoveryConsole) Init(ctx *common.ContextPole) {
	nc.ctx = ctx.NewSubCtx()

}

func (nc *DiscoveryConsole) Shutdown() {

}

type DiscoverySdkAPI struct {
	cfg    sys.Properties
	ctx    *common.ContextPole
	server pole_rpc.TransportServer
	core   *DiscoveryCore
}

func newDiscoverySdkAPI(cfg sys.Properties) *DiscoverySdkAPI {
	return &DiscoverySdkAPI{
		cfg: cfg,
	}
}

// 注册请求处理器
func (na *DiscoverySdkAPI) Init(ctx *common.ContextPole) {
	subCtx := ctx.NewSubCtx()
	na.initRpcServer(subCtx)
	na.initHttpServer(subCtx)
}

func (na *DiscoverySdkAPI) initRpcServer(ctx *common.ContextPole) {
	na.server = pole_rpc.NewRSocketServer(context.Background(), "POLE-DISCOVERY", int32(na.cfg.DiscoveryPort), na.cfg.OpenSSL)
	na.server.RegisterRequestHandler(InstanceRegister, na.instanceRegister)
	na.server.RegisterRequestHandler(InstanceDeregister, na.instanceDeregister)
	na.server.RegisterRequestHandler(InstanceMetadataUpdate, na.instanceMetadataUpdate)
	na.server.RegisterRequestHandler(InstanceDisabled, na.instanceDisabled)
}

func (na *DiscoverySdkAPI) initHttpServer(ctx *common.ContextPole) {
}

func (na *DiscoverySdkAPI) Shutdown() {
	na.ctx.Cancel()
}

func (na *DiscoverySdkAPI) instanceRegister(cxt context.Context, rpcCtx pole_rpc.RpcServerContext) {
	registerReq := &pojo.InstanceRegister{}
	err := ptypes.UnmarshalAny(rpcCtx.GetReq().Body, registerReq)
	if err != nil {
		rpcCtx.Send(&pole_rpc.ServerResponse{
			Code: -1,
			Msg:  err.Error(),
		})
	}
	na.core.serviceMgn.addInstance(registerReq, rpcCtx)
}

func (na *DiscoverySdkAPI) instanceDeregister(cxt context.Context, rpcCtx pole_rpc.RpcServerContext) {
}

// just for Http API
func (na *DiscoverySdkAPI) instanceHeartbeat(cxt context.Context, sink mono.Sink) {
}

func (na *DiscoverySdkAPI) instanceDisabled(cxt context.Context, rpcCtx pole_rpc.RpcServerContext) {
	registerReq := &pojo.InstanceDisabled{}
	err := ptypes.UnmarshalAny(rpcCtx.GetReq().Body, registerReq)
	if err != nil {
		rpcCtx.Send(&pole_rpc.ServerResponse{
			Code: -1,
			Msg:  err.Error(),
		})
	}
}

func (na *DiscoverySdkAPI) instanceMetadataUpdate(cxt context.Context, rpcCtx pole_rpc.RpcServerContext) {
}
