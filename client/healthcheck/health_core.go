// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package healthcheck

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Conf-Group/pole/common"
	"github.com/Conf-Group/pole/pojo"
	"github.com/Conf-Group/pole/transport"
	"github.com/Conf-Group/pole/utils"
)

type CustomerHealthChecker func() bool

type customerTimeF struct {
	task	*CustomerHealthCheckTask
}

func (ctf *customerTimeF) Run()  {
	if !ctf.task.F() {
		// TODO 当前实例的健康状态为非健康
	}
}

type CustomHealthCheckPlugin struct {
	rwLock         sync.RWMutex
	htw            *utils.HashTimeWheel
	taskRepository map[string]*CustomerHealthCheckTask
}

func (h *CustomHealthCheckPlugin) Name() string {
	return HealthCheck_Customer
}

func (h *CustomHealthCheckPlugin) Init(ctx *common.ContextPole) {
	h.taskRepository = make(map[string]*CustomerHealthCheckTask)
	h.htw = utils.NewTimeWheel(time.Duration(5) * time.Second, 16)
}

func (h *CustomHealthCheckPlugin) Run() {

}

func (h *CustomHealthCheckPlugin) AddTask(task HealthCheckTask) (bool, error) {
	cTask := task.(*CustomerHealthCheckTask)
	defer h.rwLock.Unlock()
	h.rwLock.Lock()
	h.taskRepository[fmt.Sprintf("%s:%d", cTask.Instance.Ip, cTask.Instance.Port)] = cTask
	h.htw.AddTask(&customerTimeF{task: cTask}, time.Duration(5) * time.Second)
	return true, nil
}

func (h *CustomHealthCheckPlugin) RemoveTask(task HealthCheckTask) (bool, error) {
	cTask := task.(*CustomerHealthCheckTask)
	defer h.rwLock.Unlock()
	h.rwLock.Lock()
	delete(h.taskRepository, fmt.Sprintf("%s:%d", cTask.Instance.Ip, cTask.Instance.Port))
	return true, nil
}

func (h *CustomHealthCheckPlugin) Destroy() {

}

type HttpBeatHealthCheckPlugin struct {
	client         *transport.HttpClient
	rwLock         sync.RWMutex
	htw            *utils.HashTimeWheel
	taskRepository map[string]*HttpHealthCheckTask
}

func (h *HttpBeatHealthCheckPlugin) Name() string {
	return HealthCheck_Http
}

func (h *HttpBeatHealthCheckPlugin) Init(ctx *common.ContextPole) {
	h.client = transport.NewHttpClient()
	h.taskRepository = make(map[string]*HttpHealthCheckTask)
	h.htw = utils.NewTimeWheel(time.Duration(5) * time.Second, 16)
}

func (h *HttpBeatHealthCheckPlugin) Run() {

}

func (h *HttpBeatHealthCheckPlugin) AddTask(task HealthCheckTask) (bool, error) {
	httpTask := task.(*HttpHealthCheckTask)
	defer h.rwLock.Unlock()
	h.rwLock.Lock()
	h.taskRepository[fmt.Sprintf("%s:%d", httpTask.Instance.Ip, httpTask.Instance.Port)] = httpTask
	return true, nil
}

func (h *HttpBeatHealthCheckPlugin) RemoveTask(task HealthCheckTask) (bool, error) {
	httpTask := task.(*HttpHealthCheckTask)
	defer h.rwLock.Unlock()
	h.rwLock.Lock()
	delete(h.taskRepository, fmt.Sprintf("%s:%d", httpTask.Instance.Ip, httpTask.Instance.Port))
	return true, nil
}

func (h *HttpBeatHealthCheckPlugin) Destroy() {

}

type TcpHealthCheckPlugin struct {
	rwLock         sync.RWMutex
	htw            *utils.HashTimeWheel
	taskRepository map[string]*tcpClient
}

type tcpClient struct {
	serverAddr string
	conn       *net.TCPConn
	instance   *pojo.Instance
}

func (h *TcpHealthCheckPlugin) Name() string {
	return HealthCheck_Tcp
}

func (h *TcpHealthCheckPlugin) Init(ctx *common.ContextPole) {
	h.taskRepository = make(map[string]*tcpClient)
	h.htw = utils.NewTimeWheel(time.Duration(5) * time.Second, 16)
}

func (h *TcpHealthCheckPlugin) Run() {

}

func (h *TcpHealthCheckPlugin) AddTask(task HealthCheckTask) (bool, error) {
	tcpTask := task.(*TcpHealthCheckTask)
	servAddr := fmt.Sprintf("%s:%d", tcpTask.Instance.Ip, tcpTask.Instance.Port)
	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
	if err != nil {
		return false, err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return false, err
	}

	defer h.rwLock.Unlock()
	h.rwLock.Lock()
	h.taskRepository[servAddr] = &tcpClient{
		serverAddr: servAddr,
		conn:       conn,
		instance:   &tcpTask.Instance,
	}

	return true, nil
}

func (h *TcpHealthCheckPlugin) RemoveTask(task HealthCheckTask) (bool, error) {
	tcpTask := task.(*TcpHealthCheckTask)
	servAddr := fmt.Sprintf("%s:%d", tcpTask.Instance.Ip, tcpTask.Instance.Port)
	defer h.rwLock.Unlock()
	h.rwLock.Lock()
	delete(h.taskRepository, servAddr)
	return true, nil
}

func (h *TcpHealthCheckPlugin) Destroy() {

}
