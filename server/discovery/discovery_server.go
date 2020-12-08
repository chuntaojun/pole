// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package discovery

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/proto"
	"github.com/rsocket/rsocket-go/rx/mono"

	"github.com/Conf-Group/pole/common"
	"github.com/Conf-Group/pole/pojo"
	"github.com/Conf-Group/pole/server/auth"
	"github.com/Conf-Group/pole/server/sys"
	"github.com/Conf-Group/pole/transport/rsocket"
)

const (
	InstanceCommonPath = sys.APIVersion + "/discovery/instance/"
	InstanceRegister   = InstanceCommonPath + "register"
	InstanceDeregister = InstanceCommonPath + "deregister"
	InstanceUpdate     = InstanceCommonPath + "update"
	InstanceHeartbeat  = InstanceCommonPath + "heartbeat"
)

var (
	InstanceRegisterSupplier = func() proto.Message {
		return &pojo.InstanceRegister{}
	}
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
}

func (d *DiscoveryServer) Shutdown() {
	d.console.Shutdown()
	d.api.Shutdown()
}

type DiscoveryConsole struct {
	httpServer *gin.Engine
	ctx        context.Context
}

func newDiscoveryConsole(cfg sys.Properties, httpServer *gin.Engine) *DiscoveryConsole {
	return &DiscoveryConsole{
		httpServer: httpServer,
	}
}

func (nc *DiscoveryConsole) Init(ctx *common.ContextPole) {
	subCtx, _ := context.WithCancel(ctx)
	nc.ctx = subCtx

}

func (nc *DiscoveryConsole) Shutdown() {

}

type DiscoverySdkAPI struct {
	cfg    sys.Properties
	server *rsocket.RSocketServer
	ctx    context.Context
}

func newDiscoverySdkAPI(cfg sys.Properties) *DiscoverySdkAPI {
	return &DiscoverySdkAPI{
		cfg: cfg,
	}
}

// 注册请求处理器
func (na *DiscoverySdkAPI) Init(ctx *common.ContextPole) {
	subCtx, _ := context.WithCancel(ctx)
	na.server = rsocket.NewRSocketServer(subCtx, "POLE-DISCOVERY", na.cfg.DiscoveryPort, na.cfg.OpenSSL)

	na.server.Dispatcher.RegisterRequestResponseHandler(InstanceRegister, auth.WriteOnly, InstanceRegisterSupplier,
		func(ctx context.Context, req proto.Message, sink mono.Sink) {
			na.instanceRegister(ctx, req.(*pojo.InstanceRegister), sink)
		})
	na.server.Dispatcher.RegisterRequestResponseHandler(InstanceDeregister, auth.WriteOnly, InstanceRegisterSupplier,
		func(ctx context.Context, req proto.Message, sink mono.Sink) {
			na.instanceDeregister(ctx, req.(*pojo.InstanceDeregister), sink)
		})
	na.server.Dispatcher.RegisterRequestResponseHandler(InstanceHeartbeat, auth.WriteOnly, InstanceRegisterSupplier,
		func(ctx context.Context, req proto.Message, sink mono.Sink) {
			na.instanceDeregister(ctx, req.(*pojo.InstanceDeregister), sink)
		})
	na.server.Dispatcher.RegisterRequestResponseHandler(InstanceUpdate, auth.WriteOnly, InstanceRegisterSupplier,
		func(ctx context.Context, req proto.Message, sink mono.Sink) {
			na.instanceUpdate(ctx, req.(*pojo.InstanceRegister), sink)
		})
}

func (na *DiscoverySdkAPI) Shutdown() {
	na.ctx.Done()
}

func (na *DiscoverySdkAPI) instanceRegister(ctx context.Context, req *pojo.InstanceRegister, sink mono.Sink) {

}

func (na *DiscoverySdkAPI) instanceDeregister(ctx context.Context, req *pojo.InstanceDeregister, sink mono.Sink) {

}

func (na *DiscoverySdkAPI) instanceHeartbeat(ctx context.Context, req *pojo.InstanceHeartBeat, sink mono.Sink) {

}

func (na *DiscoverySdkAPI) instanceUpdate(ctx context.Context, req *pojo.InstanceRegister, sink mono.Sink) {

}
