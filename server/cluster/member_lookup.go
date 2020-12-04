// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cluster

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Conf-Group/pole/notify"
	"github.com/Conf-Group/pole/server/sys"
	"github.com/Conf-Group/pole/utils"
)

var (
	ErrorNotSupportMode = errors.New("only support in cluster mode")
)

type MemberLookup interface {
	Start() error

	Observer(observer func(newMembers []*Member))

	Shutdown()

	Name() string
}

const (
	typeForStandaloneMemberLookup    = "standalone"
	typeForFileMemberLookup          = "file"
	typeForAddressServerMemberLookup = "address-server"
	typeForKubernetesMemberLookup    = "kubernetes"
)

func CreateMemberLookup(ctx context.Context, config *sys.Properties, observer func(newMembers []*Member)) (MemberLookup,
	error) {
	var lookup MemberLookup
	if config.IsStandaloneMode() {
		lookup = &standaloneMemberLookup{}
	} else {
		val := sys.GetEnvHolder().ClusterCfg.LookupCfg.MemberLookupType
		switch val {
		case typeForFileMemberLookup:
			lookup = &FileMemberLookup{
				BaseMemberLookup{
					ctx:    context.Background(),
					config: config,
				},
			}
		case typeForAddressServerMemberLookup:
			lookup = &addressServerMemberLookup{
				BaseMemberLookup: BaseMemberLookup{
					ctx: ctx,
				},
				addressServer: "",
				addressPort:   0,
				urlPath:       "",
				isShutdown:    false,
			}
		case typeForKubernetesMemberLookup:
			return nil, nil
		default:
			panic(fmt.Errorf("this member-lookup [%s] unrealized", val))
		}
	}

	lookup.Observer(observer)
	err := lookup.Start()
	return lookup, err
}

func SwitchMemberLookupAndCloseOld(ctx context.Context, name string, config *sys.Properties, oldLookup MemberLookup, observer func(newMembers []*Member)) (MemberLookup, error) {

	if config.IsStandaloneMode() {
		return nil, ErrorNotSupportMode
	}

	oldLookup.Shutdown()
	var newLookup MemberLookup

	switch name {
	case typeForFileMemberLookup:
		newLookup = &FileMemberLookup{
			BaseMemberLookup{
				ctx:    context.Background(),
				config: config,
			},
		}
	case typeForAddressServerMemberLookup:
		newLookup = &addressServerMemberLookup{
			BaseMemberLookup: BaseMemberLookup{
				ctx: ctx,
			},
			addressServer: "",
			addressPort:   0,
			urlPath:       "",
			isShutdown:    false,
		}
	case typeForKubernetesMemberLookup:
		return nil, nil
	default:
		panic(fmt.Errorf("this member-lookup [%s] unrealized", name))
	}

	newLookup.Observer(observer)
	err := newLookup.Start()
	return newLookup, err
}

type BaseMemberLookup struct {
	ctx      context.Context
	config   *sys.Properties
	observer func(newMembers []*Member)
}

// standaloneMemberLookup start
type standaloneMemberLookup struct {
	BaseMemberLookup
	ip string
}

func (s *standaloneMemberLookup) Start() error {
	s.ip = utils.FindSelfIP()
	return nil
}

func (s *standaloneMemberLookup) Observer(observer func(newMembers []*Member)) {
	s.observer = observer
}

func (s *standaloneMemberLookup) Shutdown() {
}

func (s *standaloneMemberLookup) Name() string {
	return typeForStandaloneMemberLookup
}

// ==================== standaloneMemberLookup end ====================

// ==================== KubernetesMemberLookup start ====================

type KubernetesMemberLookup struct {
	BaseMemberLookup
}

// ==================== KubernetesMemberLookup end ====================

// ==================== FileMemberLookup start ====================

type FileMemberLookup struct {
	BaseMemberLookup
}

func (s *FileMemberLookup) Start() error {
	return notify.AddWatcher(s.config.GetConfPath(), s)
}

func (s *FileMemberLookup) Observer(observer func(newMembers []*Member)) {
	s.observer = observer
}

func (s *FileMemberLookup) Shutdown() {
	_ = notify.RemoveWatcher(s.config.GetConfPath(), s)
}

func (s *FileMemberLookup) Name() string {
	return typeForFileMemberLookup
}

func (s *FileMemberLookup) OnEvent(event notify.FileChangeEvent) {
	path := s.config.GetClusterConfPath()
	s.observer(MultiParse(utils.ReadFileAllLine(path)...))
}

func (s *FileMemberLookup) FileName() string {
	return "cluster.conf"
}

// ==================== FileMemberLookup end ====================

// ==================== addressServerMemberLookup start ====================

type addressServerMemberLookup struct {
	BaseMemberLookup
	addressServer string
	addressPort   uint64
	urlPath       string
	isShutdown    bool
}

func (s *addressServerMemberLookup) Start() error {
	s.addressServer = sys.GetEnvHolder().ClusterCfg.LookupCfg.AddressLookupCfg.ServerAddr
	s.addressPort = sys.GetEnvHolder().ClusterCfg.LookupCfg.AddressLookupCfg.ServerPort
	s.urlPath = sys.GetEnvHolder().ClusterCfg.LookupCfg.AddressLookupCfg.ServerPath
	utils.DoTickerSchedule(s.ctx, func() {
		url := utils.BuildHttpsUrl(s.addressServer, s.urlPath, s.addressPort)
		for {
			if s.isShutdown {
				s.ctx.Done()
			} else {
				resp, err := http.Get(url)
				if err != nil {
					fmt.Printf("get cluster.conf from address-server has error  : %s\n", err)
				} else {
					if resp.StatusCode == http.StatusOK {
						ss := utils.ReadAllLines(resp.Body)
						s.observer(MultiParse(ss...))
					} else {
						fmt.Printf("get cluster.conf from address-server failed : %s\n", utils.ReadContent(resp.Body))
					}
				}
			}
		}
	}, time.Duration(5)*time.Second)
	return nil
}

func (s *addressServerMemberLookup) Observer(observer func(newMembers []*Member)) {
	s.observer = observer
}

func (s *addressServerMemberLookup) Shutdown() {
	s.isShutdown = true
}

func (s *addressServerMemberLookup) Name() string {
	return typeForAddressServerMemberLookup
}

// ==================== addressServerMemberLookup end ====================

// ==================== kubernetesMemberLookup start ====================

type kubernetesMemberLookup struct {
}

func (kml *kubernetesMemberLookup) Start() error {
	return nil
}

func (kml *kubernetesMemberLookup) Observer(observer func(newMembers []*Member)) {

}

func (kml *kubernetesMemberLookup) Shutdown() {

}

func (kml *kubernetesMemberLookup) Name() string {
	return typeForKubernetesMemberLookup
}

// ==================== kubernetesMemberLookup end ====================
