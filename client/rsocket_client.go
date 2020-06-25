package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"strconv"
	
	"github.com/golang/protobuf/proto"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx/mono"
	
	"nacos-go/constants"
	"nacos-go/pojo"
	"nacos-go/transport"
	"nacos-go/utils"
)

var (
	ServerNotFount = errors.New("target server not found")
)

type RsocketClient struct {
	sockets    map[string]rsocket.Client
	dispatcher *transport.Dispatcher
	token      string
}

func NewRsocketClient(label, token string, serverAddr []string, openTSL bool) *RsocketClient {
	
	client := &RsocketClient{
		sockets:    make(map[string]rsocket.Client),
		token:      token,
		dispatcher: transport.NewDispatcher(label),
	}
	
	for _, address := range serverAddr {
		ip, port := utils.AnalyzeIpAndPort(address)
		start := rsocket.Connect().
			OnClose(func(err error) {
			
			}).
			Acceptor(func(socket rsocket.RSocket) rsocket.RSocket {
				return rsocket.NewAbstractSocket(client.dispatcher.CreateRequestResponseSocket(), client.dispatcher.CreateRequestChannelSocket())
			}).
			Transport("tcp://"+ip+":"+strconv.FormatInt(int64(port), 10))
		var err error
		var c rsocket.Client
		
		if openTSL {
			c, err = start.StartTLS(context.Background(), &tls.Config{})
		} else {
			c, err = start.Start(context.Background())
		}
		
		if err != nil {
			panic(err)
		}
		client.sockets[address] = c
	}
	
	return client
}

func (c *RsocketClient) SendRequest(serverAddr string, req *pojo.GrpcRequest) mono.Mono {
	header := map[string]string {
		constants.TokenKey: c.token,
	}
	
	hb, err := json.Marshal(header)
	if err != nil {
		return mono.Error(err)
	}
	
	if rc, ok := c.sockets[serverAddr]; ok {
		body, err := proto.Marshal(req)
		if err != nil {
			panic(err)
		}
		return rc.RequestResponse(payload.New(body, hb))
	} else {
		panic(ServerNotFount)
	}
}
