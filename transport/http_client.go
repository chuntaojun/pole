package transport

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	flux2 "github.com/jjeffcaii/reactor-go/flux"
	mono2 "github.com/jjeffcaii/reactor-go/mono"

	"github.com/Conf-Group/pole/common"
	"github.com/Conf-Group/pole/pojo"
)

type HttpTransportClient struct {
	serverAddr string
	openTSL    bool
	client     *HttpClient
	bc         *baseTransportClient
}

func newHttpClient(serverAddr string, openTSL bool) (*HttpTransportClient, error) {
	return &HttpTransportClient{
		serverAddr: serverAddr,
		openTSL:    openTSL,
		client:     NewHttpClient(openTSL),
		bc:         newBaseClient(),
	}, nil
}

func (c *HttpTransportClient) AddChain(filter func(req *pojo.ServerRequest)) {
	c.bc.AddChain(filter)
}

func (c *HttpTransportClient) Request(req *pojo.ServerRequest) (mono2.Mono, error) {
	c.bc.DoFilter(req)

	return mono2.Create(func(ctx context.Context, s mono2.Sink) {
		resp, err := c.client.Put(common.EmptyContext, c.serverAddr, req.Label, req)
		if err != nil {
			s.Error(err)
		} else {
			s.Success(resp)
		}
	}), nil
}

func (c *HttpTransportClient) RequestChannel(call func(resp *pojo.ServerResponse, err error)) flux2.Sink {
	var _sink flux2.Sink
	flux2.Create(func(ctx context.Context, sink flux2.Sink) {
		_sink = sink
	}).Map(func(any reactor.Any) (reactor.Any, error) {
		req := any.(*pojo.ServerRequest)
		c.bc.DoFilter(req)
		resp, err := c.client.Put(common.EmptyContext, c.serverAddr, req.Label, req)
		call(resp, err)
		return nil, nil
	}).Subscribe(context.Background())
	return _sink
}

func (c *HttpTransportClient) Close() error {
	return nil
}
