// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transport

import (
	"context"
	"net/url"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jjeffcaii/reactor-go"
	flux2 "github.com/jjeffcaii/reactor-go/flux"
	mono2 "github.com/jjeffcaii/reactor-go/mono"

	"github.com/Conf-Group/pole/common"
	"github.com/Conf-Group/pole/pojo"
	"github.com/Conf-Group/pole/utils"
)

type WebSocketClient struct {
	client        *websocket.Conn
	bc            *baseTransportClient
	future        map[string]mono2.Sink
	channelFuture map[string]func(resp *pojo.ServerResponse, err error)
	done          chan bool
}

func newWebSocketClient(serverAddr string, openTSL bool) (*WebSocketClient, error) {
	u := url.URL{Scheme: utils.IF(openTSL, func() interface{} {
		return "wss"
	}, func() interface{} {
		return "ws"
	}).(string), Host: serverAddr, Path: "/pole"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	wsc := &WebSocketClient{
		client:        c,
		bc:            newBaseClient(),
		future:        make(map[string]mono2.Sink),
		channelFuture: make(map[string]func(resp *pojo.ServerResponse, err error)),
		done:          make(chan bool),
	}

	wsc.AddChain(func(req *pojo.ServerRequest) {
		req.Header[common.RequestID] = uuid.New().String()
	})

	wsc.initServerRead()

	return wsc, nil
}

func (wsc *WebSocketClient) initServerRead() {
	utils.GoEmpty(func() {
		for {
			select {
			case <-wsc.done:
				return
			default:
				_, message, err := wsc.client.ReadMessage()
				if err != nil {
					return
				}
				resp := &pojo.ServerResponse{}
				if err := proto.Unmarshal(message, resp); err != nil {
					continue
				}
				requestId := resp.Header[common.RequestID]
				if v, exist := wsc.future[requestId]; exist {
					v.Success(resp)
					continue
				}
				if f, exist := wsc.channelFuture[requestId]; exist {
					f(resp, nil)
					continue
				}
			}
		}
	})
}

func (wsc *WebSocketClient) AddChain(filter func(req *pojo.ServerRequest)) {
	wsc.bc.AddChain(filter)
}

func (wsc *WebSocketClient) Request(req *pojo.ServerRequest) (mono2.Mono, error) {
	wsc.bc.DoFilter(req)

	return mono2.Create(func(ctx context.Context, s mono2.Sink) {
		body, err := proto.Marshal(req)
		if err != nil {
			s.Error(err)
		}

		wsc.future[req.Header[common.RequestID]] = s

		err = wsc.client.WriteMessage(websocket.TextMessage, body)
		if err != nil {
			s.Error(err)
		}

	}).Timeout(time.Duration(10) * time.Second), nil
}

func (wsc *WebSocketClient) RequestChannel(call func(resp *pojo.ServerResponse, err error)) flux2.Sink {
	var _sink flux2.Sink
	flux2.Create(func(ctx context.Context, sink flux2.Sink) {
		_sink = sink
	}).Map(func(any reactor.Any) (reactor.Any, error) {
		req := any.(*pojo.ServerRequest)
		wsc.bc.DoFilter(req)
		body, err := proto.Marshal(req)
		if err != nil {
			return nil, err
		}
		return body, nil
	}).Map(func(any reactor.Any) (reactor.Any, error) {
		body := any.([]byte)
		err := wsc.client.WriteMessage(websocket.TextMessage, body)
		if err != nil {
			call(nil, err)
		}
		return nil, err
	}).Subscribe(context.Background())
	return _sink
}

func (wsc *WebSocketClient) Close() error {
	close(wsc.done)
	return nil
}
