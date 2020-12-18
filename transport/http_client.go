package transport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"

	"github.com/Conf-Group/pole/common"
	"github.com/Conf-Group/pole/pojo"
)

var defaultHttpClient *HttpClient

func init() {
	defaultHttpClient = NewHttpClient()
}

// http 客户端
type HttpClient struct {
	client *http.Client
}

func NewDefaultClient() *HttpClient {
	return defaultHttpClient
}

func NewHttpClient() *HttpClient {
	return &HttpClient{client: &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:          1200,            // 连接池中最大连接数
			MaxIdleConnsPerHost:   300,             // 连接池中每个ip的最大连接数
			TLSHandshakeTimeout:   5 * time.Second, // 限制TLS握手的时间
			ResponseHeaderTimeout: 5 * time.Second, // 限制读取response header的超时时间
			IdleConnTimeout:       90 * time.Second,
		},
	}}
}

// get 请求
func (hc *HttpClient) Get(ctx *common.ContextPole, server, url string, params proto.Message,
	headers map[string]string) (*pojo.ServerResponse, error) {

	//new request
	finalUrl := BuildHttpUrl(server, url)
	req, err := http.NewRequest(http.MethodGet, finalUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("new request fail for : %s", url)
	}
	any, _ := ptypes.MarshalAny(params)
	sReq := &pojo.ServerRequest{
		Label:     ctx.Value(common.ModuleLabel).(string),
		Body:      any,
		RequestId: ctx.Value(common.RequestID).(string),
		Header:    headers,
	}

	sReqBytes, _ := proto.Marshal(sReq)
	if err := req.Write(bytes.NewBuffer(sReqBytes)); err != nil {
		return nil, err
	}
	//http client
	return hc.sendRequest(req)
}

// post 请求
func (hc *HttpClient) Post(ctx *common.ContextPole, server, url string, body *pojo.ServerRequest,
	headers map[string]string) (*pojo.ServerResponse, error) {
	//add post body
	return hc.submit(ctx, server, url, http.MethodPost, body, headers)
}

// put 请求
func (hc *HttpClient) Put(ctx *common.ContextPole, server, url string,
	body proto.Message, headers map[string]string) (*pojo.ServerResponse, error) {
	//add put body
	return hc.submit(ctx, server, url, http.MethodPut, body, headers)
}

// 同时向多个 server 顺序发起 get 请求，其中一个成功就立即结束
func (hc *HttpClient) GetWithServerList(ctx *common.ContextPole, servers []string, url string, params proto.Message,
	headers map[string]string) (*pojo.ServerResponse, error) {
	for _, server := range servers {
		resp, err := hc.Get(ctx, server, url, params, headers)
		if err != nil {
			continue
		}
		return resp, nil
	}
	return nil, fmt.Errorf("the current server list is not available : %v", servers)
}

// 同时向多个 server 顺序发起 post 请求，其中一个成功就立即结束
func (hc *HttpClient) PostWithServerList(ctx *common.ContextPole, servers []string, url string, body *pojo.ServerRequest, headers map[string]string) (*pojo.ServerResponse, error) {
	for _, server := range servers {
		resp, err := hc.Post(ctx, server, url, body, headers)
		if err != nil {
			continue
		}
		return resp, nil
	}
	return nil, fmt.Errorf("the current server list is not available : %v", servers)
}

// 同时向多个 server 顺序发起 put 请求，其中一个成功就立即结束
func (hc *HttpClient) PutWithServerList(ctx *common.ContextPole, servers []string, url string, body proto.Message, headers map[string]string) (*pojo.ServerResponse, error) {
	for _, server := range servers {
		resp, err := hc.Put(ctx, server, url, body, headers)
		if err != nil {
			continue
		}
		return resp, nil
	}
	return nil, fmt.Errorf("the current server list is not available : %v", servers)
}

// post 以及 put 请求的真正实现
func (hc *HttpClient) submit(ctx *common.ContextPole, server, url, method string, body proto.Message,
	headers map[string]string) (*pojo.ServerResponse, error) {

	//add post body
	var bodyJson []byte
	var req *http.Request
	if body != nil {
		var err error
		bodyJson, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("http post body to json failed")
		}
	}
	finalUrl := BuildHttpUrl(server, url)
	req, err := http.NewRequestWithContext(ctx, method, finalUrl, bytes.NewBuffer(bodyJson))
	if err != nil {
		return nil, fmt.Errorf("new request is fail: %v", err)
	}
	req.Header.Set("Content-type", "application/json")
	any, _ := ptypes.MarshalAny(body)
	sReq := &pojo.ServerRequest{
		Label:     ctx.Value(common.ModuleLabel).(string),
		Body:      any,
		RequestId: ctx.Value(common.RequestID).(string),
		Header:    headers,
	}

	sReqBytes, _ := proto.Marshal(sReq)
	if err := req.Write(bytes.NewBuffer(sReqBytes)); err != nil {
		return nil, err
	}
	return hc.sendRequest(req)
}

func (hc *HttpClient) sendRequest(req *http.Request) (*pojo.ServerResponse, error) {
	//http client
	hResp, err := hc.client.Do(req)
	if err != nil {
		return nil, err
	}
	resp := &pojo.ServerResponse{}
	result, err := ioutil.ReadAll(hResp.Body)
	if err != nil {
		return nil, err
	}
	if err := proto.Unmarshal(result, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// 构建非TLS的http请求路径
func BuildHttpUrl(server, path string) string {
	if strings.HasPrefix(path, "/") {
		return "http://" + server + path
	}
	return "http://" + server + "/" + path
}

// 构建TLS的http请求路径
func BuildHttpsUrl(server, path string) string {
	if strings.HasPrefix(path, "/") {
		return "https://" + server + path
	}
	return "https://" + server + "/" + path
}

// 构建IP:PORT
func BuildServerAddr(host string, port int64) string {
	return fmt.Sprintf("%s:%d", host, port)
}
