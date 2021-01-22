// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pole_rpc

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/flux"
	"github.com/jjeffcaii/reactor-go/mono"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type WebSocketClient struct {
	repository    *EndpointRepository
	lock          sync.RWMutex
	clientMap     map[string]*websocket.Conn
	supplier      func(endpoint Endpoint) (*websocket.Conn, error)
	bc            *BaseTransportClient
	future        map[string]mono.Sink
	channelFuture map[string]func(resp *ServerResponse, err error)
	done          chan bool
}

func newWebSocketClient(openTSL bool) (*WebSocketClient, error) {

	wsc := &WebSocketClient{
		lock:          sync.RWMutex{},
		clientMap:     make(map[string]*websocket.Conn),
		bc:            NewBaseClient(),
		future:        make(map[string]mono.Sink),
		channelFuture: make(map[string]func(resp *ServerResponse, err error)),
		done:          make(chan bool),
	}

	supplier := func(endpoint Endpoint) (*websocket.Conn, error) {
		host := fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
		var u url.URL
		if openTSL {
			u = url.URL{Scheme: "wss", Host: host, Path: "/pole"}
		} else {
			u = url.URL{Scheme: "ws", Host: host, Path: "/pole"}
		}
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		return c, err
	}
	wsc.supplier = supplier

	wsc.AddChain(func(req *ServerRequest) {
		req.Header[RequestID] = uuid.New().String()
	})

	return wsc, nil
}

func (wsc *WebSocketClient) initServerRead(client *websocket.Conn) {
	go func() {
		for {
			select {
			case <-wsc.done:
				return
			default:
				_, message, err := client.ReadMessage()
				if err != nil {
					return
				}
				resp := &ServerResponse{}
				if err := proto.Unmarshal(message, resp); err != nil {
					continue
				}
				requestId := resp.Header[RequestID]
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
	}()
}

func (wsc *WebSocketClient) RegisterConnectEventWatcher(watcher func(eventType ConnectEventType, conn net.Conn)) {
	wsc.bc.AddWatcher(watcher)
}

func (wsc *WebSocketClient) AddChain(filter func(req *ServerRequest)) {
	wsc.bc.AddChain(filter)
}

func (wsc *WebSocketClient) Request(ctx context.Context, name string, req *ServerRequest) (mono.Mono, error) {
	wsc.bc.DoFilter(req)

	conn, err := wsc.computeIfAbsent(name)
	if err != nil {
		return nil, err
	}

	return mono.Create(func(ctx context.Context, s mono.Sink) {
		body, err := proto.Marshal(req)
		if err != nil {
			s.Error(err)
		}
		wsc.future[req.Header[RequestID]] = s
		err = conn.WriteMessage(websocket.TextMessage, body)
		if err != nil {
			s.Error(err)
		}

	}).Timeout(time.Duration(10) * time.Second), nil
}

func (wsc *WebSocketClient) RequestChannel(ctx context.Context, name string, call func(resp *ServerResponse, err error)) flux.Sink {

	conn, err := wsc.computeIfAbsent(name)
	if err != nil {
		return nil
	}

	var sink flux.Sink
	flux.
		Create(func(ctx context.Context, s flux.Sink) {
			sink = s
		}).
		Map(func(any reactor.Any) (reactor.Any, error) {
			req := any.(*ServerRequest)
			wsc.bc.DoFilter(req)
			body, err := proto.Marshal(req)
			if err != nil {
				call(nil, err)
			} else {
				err = conn.WriteMessage(websocket.TextMessage, body)
				if err != nil {
					call(nil, err)
				}
			}
			return nil, nil
		}).
		SubscribeOn(scheduler.NewElastic(1)).
		Subscribe(ctx)
	return sink
}

func (wsc *WebSocketClient) Close() error {
	close(wsc.done)
	return nil
}

func (wsc *WebSocketClient) computeIfAbsent(name string) (*websocket.Conn, error) {
	success, endpoint := wsc.repository.SelectOne(name)
	if !success {
		return nil, fmt.Errorf("")
	}

	var rClient *websocket.Conn
	wsc.lock.RLock()
	if v, exist := wsc.clientMap[endpoint.GetKey()]; exist {
		rClient = v
		wsc.lock.RUnlock()
	} else {
		wsc.lock.RUnlock()
		defer wsc.lock.Unlock()
		wsc.lock.Lock()
		client, err := wsc.supplier(endpoint)
		if err != nil {
			return nil, err
		}
		wsc.clientMap[endpoint.GetKey()] = client
		rClient = client
		wsc.initServerRead(rClient)
	}
	return rClient, nil
}
