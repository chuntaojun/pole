package core

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"strconv"
	
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
	
	"nacos-go/auth"
	"nacos-go/transport"
)

type RsocketServer struct {
	IsReady    chan struct{}
	Dispatcher *transport.Dispatcher
	security   *auth.SecurityCenter
}

func NewRsocketServer(label string, port int64, center *auth.SecurityCenter, openTSL bool) *RsocketServer {
	
	r := RsocketServer{
		IsReady:    make(chan struct{}),
		Dispatcher: transport.NewDispatcher(label),
		security:   center,
	}
	
	r.Dispatcher.RegisterFilter(func(payload payload.Payload) error {
		metadata, ok := payload.Metadata()
		if ok {
			var header map[string]string
			err := json.Unmarshal(metadata, &header)
			if err != nil {
				return err
			}
			
			if r.security != nil {
				ok, err := r.security.Filter(header)
				if !ok {
					return err
				}
			}
		}
		return nil
	})
	
	go func() {
		start := rsocket.Receive().
			OnStart(func() {
				close(r.IsReady)
			}).
			Acceptor(func(setup payload.SetupPayload, sendingSocket rsocket.CloseableRSocket) (socket rsocket.RSocket, err error) {
				return rsocket.NewAbstractSocket(r.Dispatcher.CreateRequestResponseSocket(), r.Dispatcher.CreateRequestChannelSocket()), nil
			}).
			Transport("tcp://0.0.0.0:"+strconv.FormatInt(port, 10))
		var err error
		if openTSL {
			err = start.ServeTLS(context.Background(), &tls.Config{})
		} else {
			err = start.Serve(context.Background())
		}
		if err != nil {
			panic(err)
		}
	}()
	
	
	return &r
}