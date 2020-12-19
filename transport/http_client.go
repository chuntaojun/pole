package transport

import (
	"context"
	"sync"

	flux2 "github.com/jjeffcaii/reactor-go/flux"
	mono2 "github.com/jjeffcaii/reactor-go/mono"

	"github.com/Conf-Group/pole/common"
	"github.com/Conf-Group/pole/pojo"
)

type HttpTransportClient struct {
	serverAddr string
	openTSL    bool
	client     *HttpClient
	rwLock     sync.RWMutex
	filters    []func(req *pojo.ServerRequest)
}

func newHttpClient(serverAddr string, openTSL bool) *HttpTransportClient {
	return &HttpTransportClient{
		serverAddr: serverAddr,
		openTSL:    openTSL,
		client:     NewHttpClient(openTSL),
		rwLock:     sync.RWMutex{},
		filters:    nil,
	}
}

func (c *HttpTransportClient) AddChain(filter func(req *pojo.ServerRequest)) {
	defer c.rwLock.Unlock()
	c.rwLock.Lock()
	c.filters = append(c.filters, filter)
}

func (c *HttpTransportClient) Request(req *pojo.ServerRequest) (mono2.Mono, error) {
	c.rwLock.RLock()
	for _, filter := range c.filters {
		filter(req)
	}
	c.rwLock.RUnlock()

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
	return nil
}
