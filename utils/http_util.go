// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/pole-group/pole/common"
	"github.com/pole-group/pole/pojo"
)

var defaultHttpClient *HttpClient

func init() {
	defaultHttpClient = NewHttpClient(false)
}

// http 客户端
type HttpClient struct {
	openSSL bool
	client  *http.Client
}

func NewDefaultClient() *HttpClient {
	return defaultHttpClient
}

func NewHttpClient(openSSL bool) *HttpClient {
	return &HttpClient{
		openSSL: openSSL,
		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:          1200,            // 连接池中最大连接数
				MaxIdleConnsPerHost:   300,             // 连接池中每个ip的最大连接数
				TLSHandshakeTimeout:   5 * time.Second, // 限制TLS握手的时间
				ResponseHeaderTimeout: 5 * time.Second, // 限制读取response header的超时时间
				IdleConnTimeout:       90 * time.Second,
			},
		}}
}

// post 请求
func (hc *HttpClient) Post(ctx *common.ContextPole, server, url string, body *pojo.ServerRequest) (*pojo.ServerResponse, error) {
	//add post body
	return hc.submit(ctx, server, url, http.MethodPost, body)
}

// put 请求
func (hc *HttpClient) Put(ctx *common.ContextPole, server, url string, body *pojo.ServerRequest) (*pojo.ServerResponse, error) {
	//add put body
	return hc.submit(ctx, server, url, http.MethodPut, body)
}

// post 以及 put 请求的真正实现
func (hc *HttpClient) submit(ctx *common.ContextPole, server, url, method string, body *pojo.ServerRequest) (*pojo.ServerResponse,
	error) {

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
	finalUrl := IFLazy(hc.openSSL, func() interface{} {
		return BuildHttpsUrl(server, url)
	}, func() interface{} {
		return BuildHttpUrl(server, url)
	}).(string)
	req, err := http.NewRequestWithContext(ctx, method, finalUrl, bytes.NewBuffer(bodyJson))
	if err != nil {
		return nil, fmt.Errorf("new request is fail: %v", err)
	}
	req.Header.Set("Content-type", "application/json")

	sReqBytes, _ := proto.Marshal(body)
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
func BuildServerAddr(host string, port int32) string {
	return fmt.Sprintf("%s:%d", host, port)
}
func AnalyzeIPAndPort(address string) (string, int32) {
	info := strings.Split(address, ":")
	ip := strings.TrimSpace(info[0])
	port, err := strconv.ParseInt(strings.TrimSpace(info[1]), 10, 32)
	if err != nil {
		panic(err)
	}
	return ip, int32(port)
}
