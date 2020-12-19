// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/golang/protobuf/ptypes"
	"github.com/jjeffcaii/reactor-go"
	"github.com/rsocket/rsocket-go/rx/mono"
	"github.com/stretchr/testify/assert"

	"github.com/Conf-Group/pole/common"
	"github.com/Conf-Group/pole/transport"

	"github.com/Conf-Group/pole/pojo"
)

const (
	RequestTestOne = "RequestTestOne"
	RequestTestTwo = "RequestTestTwo"
)

func Test_RSocket(t *testing.T) {
	rServer := createRSocketServer(9528)
	<-rServer.IsReady
	rClient := createRSocketClient("127.0.0.1:9528")

	rServer.RegisterRequestHandler(RequestTestOne, func(cxt *common.ContextPole, req *pojo.ServerRequest) *pojo.RestResult {
		fmt.Printf("receive req %+v\n", req)
		resp := &pojo.RestResult{
			Code: 0,
			Msg:  "this is test",
			Body: nil,
		}
		return resp
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

	req := &pojo.ServerRequest{
		Label: RequestTestOne,
		Header: map[string]string{
			"Name": "Liaochuntao",
		},
		Body: any,
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	m, err := rClient.Request(req)
	if err != nil {
		t.Error(err)
	} else {
		m.DoOnNext(func(v reactor.Any) error {
			resp := v.(*pojo.ServerResponse)
			if err != nil {
				panic(err)
			}
			fmt.Printf("receive resp : %s\n", resp)
			wg.Done()
			return nil
		}).Subscribe(context.TODO())
	}

	wg.Wait()

}

func createRSocketClient(serverAddr string) *transport.RSocketClient {
	return transport.newRSocketClient(serverAddr, false)
}

func createRSocketServer(port int) *transport.RSocketServer {
	return transport.NewRSocketServer(common.NewCtxPole(), "test", int64(port), false)
}

func Test_MonoCreateHasError(t *testing.T) {
	wait := sync.WaitGroup{}
	wait.Add(1)
	reference := atomic.Value{}
	m := mono.Create(func(ctx context.Context, sink mono.Sink) {
		panic(fmt.Errorf("test mono inner panic error"))
	}).DoOnError(func(e error) {
		fmt.Printf("has error : %s\n", e)
		reference.Store(e)
		wait.Done()
	})
	ctx := context.Background()
	m.Subscribe(ctx)
	wait.Wait()
	assert.NotNil(t, reference.Load(), "must not nil")
}
