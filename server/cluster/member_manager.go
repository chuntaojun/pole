// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cluster

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	polerpc "github.com/pole-group/pole-rpc"

	"github.com/pole-group/pole/common"
	"github.com/pole-group/pole/constants"
	"github.com/pole-group/pole/notify"
	"github.com/pole-group/pole/pojo"
	"github.com/pole-group/pole/server/sys"
	"github.com/pole-group/pole/utils"
)

const (
	Health = iota
	Down
	Impeach
)

type Member struct {
	lock           sync.Locker
	MemberID       uint64                 `json:"memberId"`
	Identifier     string                 `json:"omitempty"`
	Ip             string                 `json:"ip"`
	ExtensionPorts map[ProtocolPort]int32 `json:"extensionPorts"`
	MetaData       map[string]string      `json:"metadata"`
	Status         int
	accessFailCnt  int32
}

func (m *Member) GetIdentifier() string {
	if m.Identifier == "" {
		m.Identifier = fmt.Sprintf("%s:%d", m.Ip, m.GetExtensionPort(ServerPort))
	}
	return m.Identifier
}

func (m *Member) GetAddr() string {
	return fmt.Sprintf("%s:%d", m.Ip, m.GetExtensionPort(ServerPort))
}

func (m *Member) GetIp() string {
	return m.Ip
}

func (m *Member) GetExtensionPort(key ProtocolPort) int32 {
	return m.ExtensionPorts[key]
}

func (m *Member) GetMetaDataCopy() map[string]string {
	c := make(map[string]string)
	for k, v := range m.MetaData {
		c[k] = v
	}
	return c
}

func (m *Member) RemoveMetaData(key string) {
	defer m.lock.Unlock()
	m.lock.Lock()
	delete(m.MetaData, key)
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
	self         string
	ctx          *common.ContextPole
	lock         sync.RWMutex
	memberList   *polerpc.ConcurrentMap
	lookUp       MemberLookup
	reportFuture polerpc.Future
}

func NewServerClusterManager() *ServerClusterManager {
	mgn := &ServerClusterManager{
		ctx:        nil,
		lock:       sync.RWMutex{},
		memberList: &polerpc.ConcurrentMap{},
	}

	mgn.init()
	return mgn
}

func (s *ServerClusterManager) init() {
	defer func() {
		if err := recover(); err != nil {
			sys.CoreLogger.Error("server-cluster-manager init has error : %s", err)
			panic(err)
		}
	}()

	if err := notify.RegisterPublisher(&MemberChangeEvent{}, 64); err != nil {
		panic(err)
	}

	s.self = fmt.Sprintf("%s:%d", utils.FindSelfIP(), sys.GetEnvHolder().ServerPort)
	selfMember := SingParse(s.self)
	s.memberList.Put(selfMember.GetIdentifier(), selfMember)

	lookUp, err := CreateMemberLookup(s.ctx, s.MemberChange)
	if err != nil {
		panic(err)
	}
	s.lookUp = lookUp
}

// RegisterHttpHandler 注册用于终端控制台的Http请求处理器
func (s *ServerClusterManager) RegisterHttpHandler(httpServer *gin.Engine) {
	httpServer.GET(constants.MemberInfoReportPattern, func(c *gin.Context) {
		remoteMember := &Member{}
		err := c.BindJSON(remoteMember)
		if err != nil {
			c.JSON(http.StatusBadRequest, pojo.HttpResult{
				Code: http.StatusBadRequest,
				Msg:  err.Error(),
			})
		} else {
			originMember := s.memberList.Get(remoteMember.GetIdentifier()).(*Member)
			originMember.UpdateAllMetaData(remoteMember.MetaData)
			c.JSON(http.StatusOK, pojo.HttpResult{
				Code: http.StatusOK,
				Msg:  "success",
			})
		}
	})
	httpServer.GET(constants.MemberListPattern, func(c *gin.Context) {
		c.JSON(http.StatusOK, pojo.HttpResult{
			Code: http.StatusOK,
			Body: s.GetMemberList(),
			Msg:  "success",
		})
	})
	httpServer.GET(constants.MemberLookupNowPattern, func(c *gin.Context) {
		c.JSON(http.StatusOK, pojo.HttpResult{
			Code: http.StatusOK,
			Body: s.lookUp.Name(),
			Msg:  "success",
		})
	})
	httpServer.POST(constants.MemberLookupSwitchPattern, func(c *gin.Context) {
		params := make(map[string]string)
		err := c.BindJSON(&params)
		if err != nil {
			c.JSON(http.StatusBadRequest, pojo.HttpResult{
				Code: http.StatusBadRequest,
				Msg:  err.Error(),
			})
		} else {
			val := params["lookType"]
			newLookup, err := SwitchMemberLookupAndCloseOld(s.ctx, val, s.lookUp, s.MemberChange)
			if err != nil {
				c.JSON(http.StatusInternalServerError, pojo.HttpResult{
					Code: http.StatusInternalServerError,
					Msg:  err.Error(),
				})
			} else {
				s.lookUp = newLookup
				c.JSON(http.StatusOK, pojo.HttpResult{
					Code: http.StatusOK,
					Msg:  "success",
				})
			}
		}
	})
}

func (s *ServerClusterManager) GetSelf() *Member {
	return s.memberList.Get(s.self).(*Member)
}

func (s *ServerClusterManager) FindMember(ip string, port int) (*Member, bool) {
	m := s.memberList.Get(fmt.Sprintf("%s:%d", ip, port)).(*Member)
	return m, m != nil
}

func (s *ServerClusterManager) GetMemberList() []*Member {
	list := make([]*Member, s.memberList.Size())
	i := 0
	s.memberList.ForEach(func(k, v interface{}) {
		list[i] = v.(*Member)
		i++
	})
	return list
}

func (s *ServerClusterManager) UpdateMember(newMember *Member) {
	defer s.lock.Unlock()
	s.lock.Lock()
	key := newMember.GetIdentifier()
	oldMember := s.memberList.Get(key)
	if oldMember == nil {
		return
	}
	s.memberList.Put(newMember.GetIdentifier(), newMember)
	s.notifySubscriber()
}

func (s *ServerClusterManager) MemberChange(newMembers []*Member) {
	defer s.lock.Unlock()
	s.lock.Lock()

	// 如果节点数都不一样，则一定发生了节点的变化
	hasChanged := len(newMembers) == s.memberList.Size()
	newMemberList := &polerpc.ConcurrentMap{}

	for _, member := range newMembers {
		address := member.GetIdentifier()

		if s.memberList.Get(address) == nil {
			hasChanged = true
		}

		newMemberList.Put(address, member)
	}

	s.memberList = newMemberList

	if hasChanged {
		s.notifySubscriber()
	}

}

func (s *ServerClusterManager) notifySubscriber() {
	list := s.GetMemberList()
	_, err := notify.PublishEvent(&MemberChangeEvent{
		newMembers: list,
	})

	if err != nil {
		sys.CoreLogger.Info("notify member change failed : %s", err)
	}
}

func (s *ServerClusterManager) openReportSelfInfoToOtherWork() {
	s.reportFuture = polerpc.DoTimerSchedule(func() {

		waitReport := KRandomMember(3, s.GetMemberList(), func(m *Member) bool {
			return m.Status == Health
		})

		for _, m := range waitReport {
			bytes, err := json.Marshal(*m)
			if err != nil {
				sys.CoreLogger.Error("member info marshal to json has error : %s\n", err)
				continue
			}
			url := utils.BuildHttpUrl(utils.BuildServerAddr(m.Ip, m.GetExtensionPort(ServerPort)),
				constants.MemberInfoReportPattern)
			resp, err := http.Post(url, "application/json;charset=utf-8", strings.NewReader(string(bytes)))
			if err != nil {
				sys.CoreLogger.Error("report info to target member has error : %s\n", err)
				OnFail(s, m, err)
			} else {
				if http.StatusOK == resp.StatusCode {
					OnSuccess(s, m)
				} else {
					OnFail(s, m, nil)
				}
			}
		}

	}, time.Duration(2)*time.Second, func() time.Duration {
		return time.Duration(2) * time.Second
	})
}

type MemberChangeEvent struct {
	newMembers []*Member
}

func (m *MemberChangeEvent) Name() string {
	return "MemberChangeEvent"
}

func (m *MemberChangeEvent) Sequence() int64 {
	return utils.GetCurrentTimeMs()
}
