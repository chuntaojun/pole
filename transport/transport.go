// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transport

import (
	flux2 "github.com/jjeffcaii/reactor-go/flux"
	mono2 "github.com/jjeffcaii/reactor-go/mono"

	"github.com/Conf-Group/pole/common"
	"github.com/Conf-Group/pole/pojo"
)

type ServerHandler func(cxt *common.ContextPole, req *pojo.ServerRequest) *pojo.RestResult

type TransportClient interface {
	AddChain(filter func(req *pojo.ServerRequest))

	Request(req *pojo.ServerRequest) (mono2.Mono, error)

	RequestChannel(call func(resp *pojo.ServerResponse, err error)) flux2.Sink
}

type TransportServer interface {
	RegisterRequestHandler(path string, handler ServerHandler)

	RegisterStreamRequestHandler(path string, handler ServerHandler)
}

func ParseErrorToResult(err *common.PoleError) *pojo.RestResult {
	return &pojo.RestResult{
		Code: int32(err.ErrCode()),
		Msg:  err.Error(),
	}
}

type ConnectType string

const (
	ConnectTypeRSocket ConnectType = "RSocket"
	ConnectTypeHttp ConnectType = "Http"
)

func NewTransportClient(t ConnectType, serverAddr string, openTSL bool) TransportClient {
	switch t {
	case ConnectTypeHttp:
		return newHttpClient(serverAddr, openTSL)
	case ConnectTypeRSocket:
		return newRSocketClient(serverAddr, openTSL)
	default:
		return nil
	}
}
