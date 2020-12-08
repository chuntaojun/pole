// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rsocket

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net"
	"reflect"
	"sync"

	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/core/transport"
	"github.com/rsocket/rsocket-go/payload"
)

type RSocketServer struct {
	IsReady    chan struct{}
	Dispatcher *Dispatcher
	ConnMgr    *ConnManager
}

func NewRSocketServer(ctx context.Context, label string, port int64, openTSL bool) *RSocketServer {
	subCtx, _ := context.WithCancel(ctx)

	r := RSocketServer{
		IsReady:    make(chan struct{}),
		Dispatcher: NewDispatcher(label),
	}

	r.Dispatcher.RegisterFilter(func(req RSocketRequest) error {
		metadata, ok := req.Msg.Metadata()
		if ok {
			var header map[string]string
			err := json.Unmarshal(metadata, &header)
			if err != nil {
				return err
			}
		}
		return nil
	})

	go func(rServer *RSocketServer) {
		server := rsocket.Receive().
			OnStart(func() {
				close(r.IsReady)
			}).
			Acceptor(func(setup payload.SetupPayload, sendingSocket rsocket.CloseableRSocket) (socket rsocket.RSocket, err error) {
				return rsocket.NewAbstractSocket(r.Dispatcher.CreateRequestResponseSocket(), r.Dispatcher.CreateRequestChannelSocket()), nil
			}).
			Transport(func(ctx context.Context) (transport.ServerTransport, error) {
				serverTransport := transport.NewTCPServerTransport(func(ctx context.Context) (net.Listener, error) {
					var listener net.Listener
					var err error
					listener, err = net.ListenTCP("tcp", &net.TCPAddr{
						IP:   net.ParseIP("0.0.0.0"),
						Port: int(port),
						Zone: "",
					})
					if err != nil {
						return nil, err
					}
					if openTSL {
						listener = tls.NewListener(listener, &tls.Config{})
					}
					return listener, err
				})
				return &poleServerTransport{rServer: rServer, target: serverTransport}, nil
			})

		if err := server.Serve(subCtx); err != nil {
			panic(err)
		}
	}(&r)

	return &r
}

type ConnManager struct {
	rwLock         sync.RWMutex
	connRepository map[string]net.Conn
}

func (cm *ConnManager) PutConn(conn transport.Conn) {
	d := reflect.ValueOf(conn)
	netCon := d.FieldByName("conn").Interface().(net.Conn)

	defer cm.rwLock.Unlock()
	cm.rwLock.Lock()
	cm.connRepository[netCon.RemoteAddr().String()] = netCon
}

func (cm *ConnManager) RemoveConn(conn transport.Conn) {
	d := reflect.ValueOf(conn)
	netCon := d.FieldByName("conn").Interface().(net.Conn)

	defer cm.rwLock.Unlock()
	cm.rwLock.Lock()
	delete(cm.connRepository, netCon.RemoteAddr().String())
}

type poleServerTransport struct {
	rServer *RSocketServer
	target  transport.ServerTransport
}

// Accept register incoming connection handler.
func (p *poleServerTransport) Accept(acceptor transport.ServerTransportAcceptor) {
	proxy := func(ctx context.Context, tp *transport.Transport, onClose func(*transport.Transport)) {
		p.rServer.ConnMgr.PutConn(tp.Connection())

		wrapperOnClose := func(tp *transport.Transport) {
			p.rServer.ConnMgr.PutConn(tp.Connection())
			onClose(tp)
		}

		acceptor(ctx, tp, wrapperOnClose)
	}
	p.target.Accept(proxy)
}

// Listen listens on the network address addr and handles requests on incoming connections.
// You can specify notifier chan, it'll be sent true/false when server listening success/failed.
func (p *poleServerTransport) Listen(ctx context.Context, notifier chan<- bool) error {
	return p.target.Listen(ctx, notifier)
}

func (p *poleServerTransport) Close() error {
	return p.target.Close()
}
