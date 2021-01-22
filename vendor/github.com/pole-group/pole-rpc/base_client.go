// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pole_rpc

import (
	"context"
	"net"
	"sync"
	"time"
)

type RpcClientContext interface {
	Send(resp *ServerRequest)
}

type RpcServerContext interface {

	GetReq() *ServerRequest

	Send(resp *ServerResponse)

	Complete()
}

type RequestResponseHandler func(cxt context.Context, rpcCtx RpcServerContext)

type RequestChannelHandler func(cxt context.Context, rpcCtx RpcServerContext)

const (
	RequestID = "Pole-Request-ID"
)

var EmptyBytes []byte

type ProxyConn struct {
	Target  net.Conn
	OnClose func(conn net.Conn)
}

func (c *ProxyConn) Read(b []byte) (n int, err error) {
	return c.Target.Read(b)
}

func (c *ProxyConn) Write(b []byte) (n int, err error) {
	return c.Target.Write(b)
}

func (c *ProxyConn) Close() error {
	c.OnClose(c.Target)
	return c.Target.Close()
}

func (c *ProxyConn) LocalAddr() net.Addr {
	return c.Target.LocalAddr()
}

func (c *ProxyConn) RemoteAddr() net.Addr {
	return c.Target.RemoteAddr()
}

func (c *ProxyConn) SetDeadline(t time.Time) error {
	return c.Target.SetDeadline(t)
}

func (c *ProxyConn) SetReadDeadline(t time.Time) error {
	return c.Target.SetReadDeadline(t)
}

func (c *ProxyConn) SetWriteDeadline(t time.Time) error {
	return c.Target.SetWriteDeadline(t)
}

type ConnectEventType int8

const (
	ConnectEventForConnected ConnectEventType = iota
	ConnectEventForDisConnected
)

type ConnectEventWatcher func(eventType ConnectEventType, con net.Conn)

type ConnectEvent struct {
	EventType ConnectEventType
	Conn      net.Conn
}

type BaseTransportClient struct {
	rwLock    sync.RWMutex
	EventChan chan ConnectEvent
	Watchers  []ConnectEventWatcher
	Filters   []func(req *ServerRequest)
	CancelFs  []context.CancelFunc
}

func NewBaseClient() *BaseTransportClient {
	return &BaseTransportClient{
		rwLock:    sync.RWMutex{},
		Watchers:  make([]ConnectEventWatcher, 0, 0),
		EventChan: make(chan ConnectEvent, 16),
		Filters:   make([]func(req *ServerRequest), 0, 0),
	}
}

func (btc *BaseTransportClient) startConnectEventListener() {
	ctx, cancelf := context.WithCancel(context.Background())
	btc.CancelFs = append(btc.CancelFs, cancelf)
	go func(ctx context.Context) {
		for {
			select {
			case element := <-btc.EventChan:
				handle := func(event ConnectEvent) {
					defer btc.rwLock.RUnlock()
					btc.rwLock.RLock()
					for _, watcher := range btc.Watchers {
						watcher(element.EventType, event.Conn)
					}
				}
				handle(element)
			case <-ctx.Done():
				return
			}
		}
	}(ctx)
}

func (btc *BaseTransportClient) AddWatcher(watcher ConnectEventWatcher) {
	defer btc.rwLock.Unlock()
	btc.rwLock.Lock()
	btc.Watchers = append(btc.Watchers, watcher)
}

func (btc *BaseTransportClient) AddChain(filter func(req *ServerRequest)) {
	defer btc.rwLock.Unlock()
	btc.rwLock.Lock()
	btc.Filters = append(btc.Filters, filter)
}

func (btc *BaseTransportClient) DoFilter(req *ServerRequest) {
	btc.rwLock.RLock()
	for _, filter := range btc.Filters {
		filter(req)
	}
	btc.rwLock.RUnlock()
}

type ConnectType string

const (
	ConnectTypeRSocket ConnectType = "RSocket"
	ConnectWebSocket   ConnectType = "WebSocket"
)
