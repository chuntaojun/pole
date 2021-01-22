// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pole_rpc

import (
	"context"
	"net"
)

type TransportClient interface {
	RegisterConnectEventWatcher(watcher func(eventType ConnectEventType, conn net.Conn))

	AddChain(filter func(req *ServerRequest))

	Request(ctx context.Context, name string, req *ServerRequest) (*ServerResponse, error)

	RequestChannel(ctx context.Context, name string, call func(resp *ServerResponse, err error)) (RpcClientContext, error)

	Close() error
}

type TransportServer interface {
	RegisterRequestHandler(path string, handler RequestResponseHandler)

	RegisterChannelRequestHandler(path string, handler RequestChannelHandler)
}

func NewTransportClient(t ConnectType, repository *EndpointRepository, openTSL bool) (TransportClient,
	error) {
	switch t {
	case ConnectTypeRSocket:
		return NewRSocketClient(openTSL, repository)
	default:
		return nil, nil
	}
}
