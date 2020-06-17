package transport

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx/mono"
	. "nacos-go/pojo"
	"sync"
)

type RsocketServer struct {

	Label	string

	IsReady chan struct{}

	Handler map[int32]func(req *any.Any) *any.Any

	lock sync.Locker
}

func CreateRsocketServer(label string, port int64) (*RsocketServer, error) {

	r := RsocketServer{
		IsReady: make(chan struct{}),
		Label:label,
	}

	err := rsocket.Receive().
			OnStart(func() {
				close(r.IsReady)
			}).
			Resume().
			Acceptor(func(setup payload.SetupPayload, sendingSocket rsocket.CloseableRSocket) (socket rsocket.RSocket, err error) {
				return rsocket.NewAbstractSocket(r.createOptAbstractSocket()), nil
			}).
			Transport("tcp://0.0.0.0:" + string(port)).
			ServeTLS(context.Background(), &tls.Config{})

	return &r, err
}

func (r *RsocketServer) createOptAbstractSocket() rsocket.OptAbstractSocket {
	return rsocket.RequestResponse(func(msg payload.Payload) mono.Mono {
		return mono.Create(func(ctx context.Context, sink mono.Sink) {
			metaData, ok := msg.Metadata()
			if ok {
				r.auth(metaData)
			}
			body := msg.Data()
			req := &BaseRequest{}
			err := proto.Unmarshal(body, req)
			if err != nil {
				sink.Error(err)
			} else {
				handler, ok := r.Handler[req.GetLabel()]
				if ok {
					resp := handler(req.GetBody())
					bytes, err := proto.Marshal(resp)
					if err != nil {
						sink.Error(err)
					} else {
						sink.Success(payload.New(bytes, metaData))
					}
				} else {
					sink.Error(fmt.Errorf("not implement"))
				}
			}
		})
	})
}

func (r *RsocketServer) RegisterHandler(key int32, handler func(req *any.Any) *any.Any) {
	defer r.lock.Unlock()
	r.lock.Lock()

	if _, ok := r.Handler[key]; ok {
		return
	}
	r.Handler[key] = handler
}

func (r *RsocketServer) auth(metadata []byte) bool {
	return false
}
