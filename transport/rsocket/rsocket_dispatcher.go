// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rsocket

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/jjeffcaii/reactor-go/scheduler"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx/flux"
	"github.com/rsocket/rsocket-go/rx/mono"

	"github.com/Conf-Group/pole/auth"
	"github.com/Conf-Group/pole/pojo"
)

var (
	ErrorNotImplement = errors.New("not implement")
)

type reqRespHandler struct {
	supplier func() proto.Message
	handler  func(ctx context.Context, req proto.Message, sink mono.Sink)
	op       auth.OperationType
}

type reqChannelHandler struct {
	supplier func() proto.Message
	handler  func(ctx context.Context, req proto.Message, sink flux.Sink)
	op       auth.OperationType
}

type Dispatcher struct {
	Label   string
	lock    sync.Mutex
	filters []func(req RSocketRequest) error

	reqRespHandler    map[string]reqRespHandler
	reqChannelHandler map[string]reqChannelHandler
}

type RSocketRequest struct {
	Op  auth.OperationType
	Msg payload.Payload
	Req *pojo.GrpcRequest
}

func NewDispatcher(label string) *Dispatcher {
	return &Dispatcher{
		Label:             label,
		reqRespHandler:    make(map[string]reqRespHandler),
		reqChannelHandler: make(map[string]reqChannelHandler),
	}
}

func (r *Dispatcher) RegisterFilter(chain ...func(req RSocketRequest) error) {
	r.filters = append(r.filters, chain...)
}

func (r *Dispatcher) CreateRequestResponseSocket() rsocket.OptAbstractSocket {
	return rsocket.RequestResponse(func(msg payload.Payload) mono.Mono {
		body := msg.Data()
		gRPCRep := &pojo.GrpcRequest{}
		err := proto.Unmarshal(body, gRPCRep)
		if err != nil {
			return mono.Error(err)
		}

		if wrap, ok := r.reqRespHandler[gRPCRep.GetLabel()]; ok {
			req := RSocketRequest{
				Op:  wrap.op,
				Msg: msg,
				Req: gRPCRep,
			}

			for _, filter := range r.filters {
				if err := filter(req); err != nil {
					return mono.Error(err)
				}
			}

			return mono.Create(func(ctx context.Context, sink mono.Sink) {
				any := gRPCRep.GetBody()
				pb := wrap.supplier()

				err := ptypes.UnmarshalAny(any, pb)

				if err != nil {
					sink.Error(err)
				} else {
					wrap.handler(context.WithValue(ctx, "Trace-ID", gRPCRep.RequestId), pb, sink)
				}
			}).DoOnError(func(e error) {
				fmt.Printf("an exception occurred while processing the request %s\n", err)
			})
		}
		return mono.Error(ErrorNotImplement)
	})
}

func (r *Dispatcher) CreateRequestChannelSocket() rsocket.OptAbstractSocket {
	return rsocket.RequestChannel(func(requests flux.Flux) (responses flux.Flux) {
		return flux.Create(func(ctx context.Context, sink flux.Sink) {
			requests.SubscribeOn(scheduler.Elastic()).
				DoOnNext(func(input payload.Payload) error {
					body := input.Data()
					gRPCRep := &pojo.GrpcRequest{}
					err := proto.Unmarshal(body, gRPCRep)
					if err != nil {
						panic(err)
					}
					if wrap, ok := r.reqChannelHandler[gRPCRep.GetLabel()]; ok {
						req := RSocketRequest{
							Op:  wrap.op,
							Msg: input,
							Req: gRPCRep,
						}

						for _, filter := range r.filters {
							if err := filter(req); err != nil {
								return err
							}
						}
						wrap.handler(context.WithValue(ctx, "Trace-ID", gRPCRep.RequestId), gRPCRep.GetBody(), sink)
						return nil
					} else {
						return ErrorNotImplement
					}
				}).
				DoOnError(func(e error) {
					fmt.Printf("an exception occurred while processing the request %s\n", e)
				}).
				Subscribe(ctx)
		})
	})
}

func (r *Dispatcher) RegisterRequestResponseHandler(key string, op auth.OperationType, supplier func() proto.Message,
	handler func(ctx context.Context, req proto.Message, sink mono.Sink)) {
	defer func() {
		r.lock.Unlock()
		if err := recover(); err != nil {
			fmt.Printf("register rep&resp handler has error %s\n", err)
		}
	}()
	r.lock.Lock()

	if _, ok := r.reqRespHandler[key]; ok {
		return
	}
	r.reqRespHandler[key] = struct {
		supplier func() proto.Message
		handler  func(ctx context.Context, req proto.Message, sink mono.Sink)
		op       auth.OperationType
	}{supplier: supplier, handler: handler, op: op}
}

func (r *Dispatcher) RegisterRequestChannelHandler(key string, op auth.OperationType, supplier func() proto.Message,
	handler func(ctx context.Context, req proto.Message, sink flux.Sink)) {
	defer r.lock.Unlock()
	r.lock.Lock()

	if _, ok := r.reqChannelHandler[key]; ok {
		return
	}
	r.reqChannelHandler[key] = struct {
		supplier func() proto.Message
		handler  func(ctx context.Context, req proto.Message, sink flux.Sink)
		op       auth.OperationType
	}{supplier: supplier, handler: handler, op: op}
}
