// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package discovery

import (
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"

	"github.com/Conf-Group/pole/common"
	"github.com/Conf-Group/pole/pojo"
	"github.com/Conf-Group/pole/server/sys"
	"github.com/Conf-Group/pole/transport"
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
	server *transport.RSocketServer
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
	na.server = transport.NewRSocketServer(ctx, "POLE-DISCOVERY", na.cfg.DiscoveryPort, na.cfg.OpenSSL)
	na.server.RegisterRequestHandler(InstanceRegister, na.instanceRegister)
	na.server.RegisterRequestHandler(InstanceDeregister, na.instanceDeregister)
	na.server.RegisterRequestHandler(InstanceMetadataUpdate, na.instanceMetadataUpdate)
	na.server.RegisterRequestHandler(InstanceDisabled, na.instanceDisabled)
}

func (na *DiscoverySdkAPI) initHttpServer(ctx *common.ContextPole) {

}

func (na *DiscoverySdkAPI) Shutdown() {
	na.ctx.Done()
}

func (na *DiscoverySdkAPI) instanceRegister(cxt *common.ContextPole, req *pojo.ServerRequest) *pojo.RestResult {
	registerReq := &pojo.InstanceRegister{}
	err := ptypes.UnmarshalAny(req.Body, registerReq)
	if err != nil {
		return transport.ParseErrorToResult(&common.PoleError{
			ErrMsg: err.Error(),
			Code:   -1,
		})
	}
	return na.core.serviceMgn.addInstance(registerReq)
}

func (na *DiscoverySdkAPI) instanceDeregister(cxt *common.ContextPole, req *pojo.ServerRequest) *pojo.RestResult {
	return nil
}

// just for Http API
func (na *DiscoverySdkAPI) instanceHeartbeat(cxt *common.ContextPole, req *pojo.ServerRequest) *pojo.RestResult {
	return nil
}

func (na *DiscoverySdkAPI) instanceDisabled(ctx *common.ContextPole, req *pojo.ServerRequest) *pojo.RestResult {
	registerReq := &pojo.InstanceDisabled{}
	err := ptypes.UnmarshalAny(req.Body, registerReq)
	if err != nil {
		return transport.ParseErrorToResult(&common.PoleError{
			ErrMsg: err.Error(),
			Code:   -1,
		})
	}
	return nil
}

func (na *DiscoverySdkAPI) instanceMetadataUpdate(cxt *common.ContextPole, req *pojo.ServerRequest) *pojo.RestResult {
	return nil
}
