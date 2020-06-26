// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"nacos-go/constants"
	"nacos-go/notify"
	"nacos-go/pojo"
	"nacos-go/sys"
	"nacos-go/utils"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	Health = iota
	Down
	Impeach
)

type Member struct {
	Identifier     string         `json:"omitempty"`
	Ip             string         `json:"ip"`
	Port           int            `json:"port"`
	ExtensionPorts map[string]int `json:"extensionPorts"`
	lock           sync.Locker
	MetaData       map[string]string `json:"metadata"`
	Status         int
	accessFailCnt  int
}

func (m *Member) GetIdentifier() string {
	if m.Identifier == "" {
		m.Identifier = m.Ip + string(m.Port)
	}
	return m.Identifier
}

func (m *Member) GetIp() string {
	return m.Ip
}

func (m *Member) GetPort() int {
	return m.Port
}

func (m *Member) GetExtensionPort(key string) int {
	return m.ExtensionPorts[key]
}

func (m *Member) GetMetaDataCopy() map[string]string {
	c := make(map[string]string)
	for k, v := range m.MetaData {
		c[k] = v
	}
	return c
}

func (m *Member) UpdateMetaData(key, value string) {
	defer m.lock.Unlock()
	m.lock.Lock()

	m.MetaData[key] = value
}

func (m *Member) UpdateAllMetaData(newMetadata map[string]string) {
	defer m.lock.Unlock()
	m.lock.Lock()

	m.MetaData = newMetadata
}

type ServerClusterManager struct {
	self       string
	ctx        context.Context
	lock       sync.RWMutex
	memberList map[string]*Member
	lookUp     MemberLookup
	config     *sys.Config
}

func (s *ServerClusterManager) Init(config *sys.Config) {

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("server-cluster-manager init has error : %s\n", err)
			panic(err)
		}
	}()

	s.self = utils.FindSelfIp()

	utils.Runnable(func() error {
		return notify.RegisterPublisher(&MemberChangeEvent{}, 64)
	})

	s.lookUp = utils.Supplier(func() (i interface{}, err error) {
		return CreateMemberLookup(config, s.MemberChange)
	}).(MemberLookup)
}

// Register the handler for each request, just use for console
func (s *ServerClusterManager) RegisterHttpHandler(group *gin.RouterGroup) []gin.IRoutes {

	var routes []gin.IRoutes

	routes = append(routes,
		group.POST(constants.MemberInfoReportPattern, func(c *gin.Context) {
			remoteMember := &Member{}
			err := c.BindJSON(remoteMember)
			if err != nil {
				c.JSON(http.StatusBadRequest, pojo.RestResult{
					Code: http.StatusBadRequest,
					Msg:  err.Error(),
				})
			} else {
				originMember := s.memberList[remoteMember.GetIdentifier()]
				originMember.UpdateAllMetaData(remoteMember.MetaData)
				c.JSON(http.StatusOK, pojo.RestResult{
					Code: http.StatusOK,
					Msg:  "success",
				})
			}
		}),
		group.POST(constants.MemberListPattern, func(c *gin.Context) {
			c.JSON(http.StatusOK, pojo.RestResult{
				Code: http.StatusOK,
				Body: s.GetMemberList(),
				Msg:  "success",
			})
		}),
		group.POST(constants.MemberLookupNowPattern, func(c *gin.Context) {
			c.JSON(http.StatusOK, pojo.RestResult{
				Code: http.StatusOK,
				Body: s.lookUp.Name(),
				Msg:  "success",
			})
		}),
		group.POST(constants.MemberLookupSwitchPattern, func(c *gin.Context) {
			params := make(map[string]string)
			err := c.BindJSON(&params)
			if err != nil {
				c.JSON(http.StatusBadRequest, pojo.RestResult{
					Code: http.StatusBadRequest,
					Msg:  err.Error(),
				})
			} else {
				val := params["lookType"]
				newLookup, err := SwitchMemberLookupAndCloseOld(val, s.config, s.lookUp, s.MemberChange)
				if err != nil {
					c.JSON(http.StatusInternalServerError, pojo.RestResult{
						Code: http.StatusInternalServerError,
						Msg:  err.Error(),
					})
				} else {
					s.lookUp = newLookup
					c.JSON(http.StatusOK, pojo.RestResult{
						Code: http.StatusOK,
						Msg:  "success",
					})
				}
			}
		}),
	)

	return routes
}

func (s *ServerClusterManager) GetSelf() *Member {
	return nil
}

func (s *ServerClusterManager) FindMember(ip string, port int) (*Member, bool) {
	defer s.lock.RUnlock()
	s.lock.RLock()

	key := ip + string(port)
	m, exist := s.memberList[key]

	return m, exist
}

func (s *ServerClusterManager) GetMemberList() []*Member {
	list := make([]*Member, len(s.memberList))
	for _, v := range s.memberList {
		list = append(list, v)
	}
	return list
}

func (s *ServerClusterManager) MemberChange(newMembers []*Member) {
	defer s.lock.Unlock()
	s.lock.Lock()

	hasChanged := false

	for _, m := range newMembers {
		if v, exist := s.memberList[m.GetIdentifier()]; exist {
			v.UpdateAllMetaData(m.MetaData)
		} else {
			s.memberList[m.GetIdentifier()] = m
			hasChanged = true
		}
	}

	if hasChanged {
		var list []*Member
		for _, m := range s.memberList {
			list = append(list, m)
		}

		_, err := notify.PublishEvent(&MemberChangeEvent{
			newMembers: list,
		})

		if err != nil {
			fmt.Printf("notify member change failed : %s", err)
		}
	}

}

func (s *ServerClusterManager) openReportSelfInfoToOtherWork() {
	utils.DoTimerSchedule(func() {

		waitReport := KRandomMember(3, s.GetMemberList(), func(m *Member) bool {
			return m.Status == Health
		})

		for _, m := range waitReport {
			bytes, err := json.Marshal(*m)
			if err != nil {
				fmt.Printf("member info marshal to json has error : %s\n", err)
				continue
			}

			url := utils.BuildHttpUrl(m.Ip, m.Port, constants.MemberInfoReportPattern)
			resp, err := http.Post(url, "application/json;charset=utf-8", strings.NewReader(string(bytes)))
			if err != nil {
				fmt.Printf("report info to target member has error : %s\n", err)
				OnFail(m, err)
			} else {
				if http.StatusOK == resp.StatusCode {
					OnSuccess(m)
				} else {
					OnFail(m, nil)
				}
			}
		}

	}, time.Duration(2)*time.Second, func() time.Duration {
		return time.Duration(2) * time.Second
	}, s.ctx)
}

type MemberChangeEvent struct {
	newMembers []*Member
}

func (m *MemberChangeEvent) Name() string {
	return "MemberChangeEvent"
}

func (m *MemberChangeEvent) Sequence() int64 {
	return time.Now().Unix()
}
