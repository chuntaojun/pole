// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transport

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
	"github.com/rsocket/rsocket-go/rx"
	"github.com/rsocket/rsocket-go/rx/flux"
	"github.com/rsocket/rsocket-go/rx/mono"

	"nacos-go/pojo"
)

var (
	ErrorNotImplement = errors.New("not implement")
)

type Dispatcher struct {
	Label   string
	lock    sync.Mutex
	filters []func(payload payload.Payload) error

	reqRespHandler map[int32]struct {
		supplier func() proto.Message
		handler  func(payload payload.Payload, req proto.Message, sink mono.Sink)
	}
	reqChannelHandler map[int32]struct {
		supplier func() proto.Message
		handler  func(payload payload.Payload, req proto.Message, sink flux.Sink)
	}
}

func NewDispatcher(label string) *Dispatcher {
	return &Dispatcher{
		Label: label,
		reqRespHandler: make(map[int32]struct {
			supplier func() proto.Message
			handler  func(payload payload.Payload, req proto.Message, sink mono.Sink)
		}),
		reqChannelHandler: make(map[int32]struct {
			supplier func() proto.Message
			handler  func(payload payload.Payload, req proto.Message, sink flux.Sink)
		}),
	}
}

func (r *Dispatcher) RegisterFilter(chain ...func(payload payload.Payload) error) {
	r.filters = append(r.filters, chain...)
}

func (r *Dispatcher) CreateRequestResponseSocket() rsocket.OptAbstractSocket {
	return rsocket.RequestResponse(func(msg payload.Payload) mono.Mono {
		for _, filter := range r.filters {
			if err := filter(msg); err != nil {
				return mono.Error(err)
			}
		}
		body := msg.Data()
		req := &pojo.GrpcRequest{}
		err := proto.Unmarshal(body, req)
		if err != nil {
			return mono.Error(err)
		}
		return mono.Create(func(ctx context.Context, sink mono.Sink) {
			wrap, ok := r.reqRespHandler[req.GetLabel()]
			any := req.GetBody()
			pb := wrap.supplier()

			err := ptypes.UnmarshalAny(any, pb)

			if err != nil {
				sink.Error(err)
			} else {
				if ok {
					wrap.handler(msg, pb, sink)
				} else {
					sink.Error(ErrorNotImplement)
				}
			}
		}).DoOnError(func(e error) {
			fmt.Printf("an exception occurred while processing the request %s\n", err)
		})
	})
}

func (r *Dispatcher) CreateRequestChannelSocket() rsocket.OptAbstractSocket {
	return rsocket.RequestChannel(func(msgs rx.Publisher) flux.Flux {
		return flux.Create(func(ctx context.Context, sink flux.Sink) {
			msgs.(flux.Flux).SubscribeOn(scheduler.Elastic()).
				DoOnNext(func(input payload.Payload) {
					var err error
					for _, filter := range r.filters {
						e := filter(input)
						if err == nil {
							err = e
						}
					}

					if err != nil {
						sink.Error(err)
					} else {
						body := input.Data()
						req := &pojo.GrpcRequest{}
						err = proto.Unmarshal(body, req)

						if err != nil {
							sink.Error(err)
						} else {
							wrap, ok := r.reqChannelHandler[req.GetLabel()]
							if ok {
								wrap.handler(input, req.GetBody(), sink)
							} else {
								sink.Error(ErrorNotImplement)
							}
						}
					}

				}).
				DoOnError(func(e error) {
					fmt.Printf("an exception occurred while processing the request %s\n", e)
				}).
				Subscribe(context.Background())
		})
	})
}

func (r *Dispatcher) RegisterRequestResponseHandler(key int32, supplier func() proto.Message, handler func(input payload.Payload, req proto.Message, sink mono.Sink)) {
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
		handler  func(payload payload.Payload, req proto.Message, sink mono.Sink)
	}{supplier: supplier, handler: handler}
}

func (r *Dispatcher) RegisterRequestChannelHandler(key int32, supplier func() proto.Message, handler func(input payload.Payload, req proto.Message, sink flux.Sink)) {
	defer r.lock.Unlock()
	r.lock.Lock()

	if _, ok := r.reqChannelHandler[key]; ok {
		return
	}
	r.reqChannelHandler[key] = struct {
		supplier func() proto.Message
		handler  func(payload payload.Payload, req proto.Message, sink flux.Sink)
	}{supplier: supplier, handler: handler}
}
