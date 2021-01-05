// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transport

import (
	"sync"

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

	Close() error
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

type baseTransportClient struct {
	rwLock     sync.RWMutex
	filters    []func(req *pojo.ServerRequest)
}

func newBaseClient() *baseTransportClient {
	return &baseTransportClient{
		rwLock:  sync.RWMutex{},
		filters: make([]func(req *pojo.ServerRequest), 0, 0),
	}
}

func (btc *baseTransportClient) AddChain(filter func(req *pojo.ServerRequest))  {
	defer btc.rwLock.Unlock()
	btc.rwLock.Lock()
	btc.filters = append(btc.filters, filter)
}

func (btc *baseTransportClient) DoFilter(req *pojo.ServerRequest)  {
	btc.rwLock.RLock()
	for _, filter := range btc.filters {
		filter(req)
	}
	btc.rwLock.RUnlock()
}

type ConnectType string

const (
	ConnectTypeRSocket ConnectType = "RSocket"
	ConnectTypeHttp ConnectType = "Http"
	ConnectWebSocket ConnectType = "WebSocket"
)

func NewTransportClient(t ConnectType, serverAddr string, openTSL bool) (TransportClient, error) {
	switch t {
	case ConnectTypeHttp:
		return newHttpClient(serverAddr, openTSL)
	case ConnectTypeRSocket:
		return newRSocketClient(serverAddr, openTSL)
	case ConnectWebSocket:
		return newWebSocketClient(serverAddr, openTSL)
	default:
		return nil, nil
	}
}
