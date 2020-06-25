// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx/mono"
	
	"nacos-go/client"
	"nacos-go/core"
	"nacos-go/pojo"
)

const (
	RequestTestOne = iota
	RequestTestTwo
)

func Test_Rsocket(t *testing.T) {
	rServer := createRsocketServer(9528)
	<- rServer.IsReady
	rClient := createRsocketClient([]string{"127.0.0.1:9528"})
	
	rServer.Dispatcher.RegisterRequestResponseHandler(RequestTestOne, func() proto.Message {
		return &pojo.Instance{}
	}, func(input payload.Payload, req proto.Message, sink mono.Sink) {
		fmt.Printf("receive req %+v\n", req)
		resp := &pojo.GrpcResponse{
			Label:RequestTestOne,
			Header: map[string]string{
				"Name": "Liaochuntao",
			},
		}
		
		body, err := proto.Marshal(resp)
		if err != nil {
			panic(err)
		}
		
		sink.Success(payload.New(body, []byte("lessspring")))
	})
	
	instance := &pojo.Instance{
		ServiceName: "elastic-search",
		Group:       "DEFAULT_GROUP",
		Ip:          "127.0.0.1",
		Port:        80,
		ClusterName: "DEFAULT_CLUSTER",
		Weight:      0,
		Metadata:    nil,
		Ephemeral:   true,
		Enabled:     false,
	}
	
	any, _ := ptypes.MarshalAny(instance)
	
	req := &pojo.GrpcRequest{
		Label:RequestTestOne,
		Header: map[string]string{
			"Name": "Liaochuntao",
		},
		Body:any,
	}
	
	wg := sync.WaitGroup{}
	wg.Add(1)
	
	rClient.SendRequest("127.0.0.1:9528", req).DoOnSuccess(func(input payload.Payload) {
		resp := &pojo.GrpcResponse{}
		err := proto.Unmarshal(input.Data(), resp)
		if err != nil {
			panic(err)
		}
		fmt.Printf("receive resp : %s\n", resp)
		wg.Done()
	}).DoOnError(func(e error) {
		fmt.Printf("receive resp has error %s\n", e)
	}).Subscribe(context.Background())
	
	wg.Wait()
}

func createRsocketClient(serverAddr []string) *client.RsocketClient {
	return client.NewRsocketClient("test", "lessspring", serverAddr, false)
}

func createRsocketServer(port int) *core.RsocketServer {
	return core.NewRsocketServer("test", int64(port), nil, false)
}
