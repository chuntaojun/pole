// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rsocket

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

	"github.com/Conf-Group/pole/constants"
	"github.com/Conf-Group/pole/pojo"
	"github.com/Conf-Group/pole/utils"
)

var (
	ServerNotFount = errors.New("target server not found")
)

type RSocketClient struct {
	sockets    map[string]rsocket.Client
	dispatcher *Dispatcher
	token      string
}

func NewRSocketClient(label, token string, serverAddr []string, openTSL bool) *RSocketClient {

	client := &RSocketClient{
		sockets:    make(map[string]rsocket.Client),
		token:      token,
		dispatcher: NewDispatcher(label),
	}

	for _, address := range serverAddr {
		ip, port := utils.AnalyzeIPAndPort(address)
		start := rsocket.Connect().
			OnClose(func(err error) {

			}).
			Acceptor(func(socket rsocket.RSocket) rsocket.RSocket {
				return rsocket.NewAbstractSocket(client.dispatcher.CreateRequestResponseSocket(), client.dispatcher.CreateRequestChannelSocket())
			}).
			Transport("tcp://" + ip + ":" + strconv.FormatInt(int64(port), 10))
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

func (c *RSocketClient) SendRequest(serverAddr string, req *pojo.GrpcRequest) mono.Mono {
	header := map[string]string{
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
