// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transport

import (
	"context"
	"crypto/tls"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/jjeffcaii/reactor-go"
	flux2 "github.com/jjeffcaii/reactor-go/flux"
	mono2 "github.com/jjeffcaii/reactor-go/mono"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/core/transport"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx/flux"

	"github.com/Conf-Group/pole/constants"
	"github.com/Conf-Group/pole/pojo"
	"github.com/Conf-Group/pole/utils"
)

type RSocketClient struct {
	socket  rsocket.Client
	rwLock  sync.RWMutex
	filters []func(req *pojo.ServerRequest)
}

func newRSocketClient(serverAddr string, openTSL bool) *RSocketClient {

	client := &RSocketClient{
		filters: make([]func(req *pojo.ServerRequest), 0, 0),
	}

	ip, port := utils.AnalyzeIPAndPort(serverAddr)
	c, err := rsocket.Connect().
		OnClose(func(err error) {

		}).
		Transport(func(ctx context.Context) (*transport.Transport, error) {
			tcb := rsocket.TCPClient().SetHostAndPort(ip, int(port))
			if openTSL {
				tcb.SetTLSConfig(&tls.Config{})
			}
			return tcb.Build()(ctx)
		}).
		Start(context.Background())
	if err != nil {
		panic(err)
	}

	client.socket = c
	return client
}

func (c *RSocketClient) AddChain(filter func(req *pojo.ServerRequest)) {
	defer c.rwLock.Unlock()
	c.rwLock.Lock()
	c.filters = append(c.filters, filter)
}

func (c *RSocketClient) Request(req *pojo.ServerRequest) (mono2.Mono, error) {
	c.rwLock.RLock()
	for _, filter := range c.filters {
		filter(req)
	}
	c.rwLock.RUnlock()

	body, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	return c.socket.RequestResponse(payload.New(body, constants.EmptyBytes)).
		Raw().
		FlatMap(func(any mono2.Any) mono2.
		Mono {
			payLoad := any.(payload.Payload)
			resp := new(pojo.ServerResponse)
			if err := proto.Unmarshal(payLoad.Data(), resp); err != nil {
				return mono2.Error(err)
			}
			return mono2.Just(resp)
		}), nil
}

func (c *RSocketClient) RequestChannel(call func(resp *pojo.ServerResponse, err error)) flux2.Sink {
	var _sink flux2.Sink
	f := flux2.Create(func(ctx context.Context, sink flux2.Sink) {
		_sink = sink
	}).Map(func(any reactor.Any) (reactor.Any, error) {
		req := any.(*pojo.ServerRequest)
		c.rwLock.RLock()
		for _, filter := range c.filters {
			filter(req)
		}
		c.rwLock.RUnlock()
		body, err := proto.Marshal(req)
		if err != nil {
			return nil, err
		}
		return payload.New(body, constants.EmptyBytes), nil
	})
	c.socket.RequestChannel(flux.Raw(f)).Raw().DoOnNext(func(v reactor.Any) error {
		payLoad := v.(payload.Payload)
		resp := new(pojo.ServerResponse)
		if err := proto.Unmarshal(payLoad.Data(), resp); err != nil {
			return err
		}
		call(resp, nil)
		return nil
	}).DoOnError(func(e error) {
		call(nil, e)
	})
	return _sink
}
