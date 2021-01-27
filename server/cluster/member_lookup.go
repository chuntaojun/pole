// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cluster

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	polerpc "github.com/pole-group/pole-rpc"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	watch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	v12 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/pole-group/pole/common"
	"github.com/pole-group/pole/notify"
	"github.com/pole-group/pole/server/sys"
	"github.com/pole-group/pole/utils"
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
		l, err := createLookupByName(ctx, val, config)
		if err != nil {
			return nil, err
		}
		lookup = l
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

	newLookup, err := createLookupByName(ctx, name, config)
	if err != nil {
		return nil, err
	}

	newLookup.Observer(observer)
	err = newLookup.Start()
	return newLookup, err
}

func createLookupByName(ctx context.Context, name string, config *sys.Properties) (MemberLookup, error) {
	var newLookup MemberLookup
	switch name {
	case typeForFileMemberLookup:
		newLookup = &FileMemberLookup{
			BaseMemberLookup{
				ctx:    ctx,
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
		newLookup = &kubernetesMemberLookup{
			BaseMemberLookup: BaseMemberLookup{
				ctx: ctx,
			},
			k8sCfg: config.ClusterCfg.LookupCfg.K8sLookupCfg,
		}
	default:
		return nil, fmt.Errorf("this member-lookup [%s] unrealized", name)
	}
	return newLookup, nil
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

// ==================== kubernetesMemberLookup start ====================

type kubernetesMemberLookup struct {
	BaseMemberLookup
	k8sCfg      sys.KubernetesLookupConfig
	endpointsOp v12.EndpointsInterface
	watcher     watch.Interface
}

func (s *kubernetesMemberLookup) Start() error {
	clientSet := &kubernetes.Clientset{}
	ctx := context.Background()

	s.endpointsOp = clientSet.CoreV1().Endpoints(s.k8sCfg.Namespace)
	watcher, err := s.endpointsOp.Watch(ctx, metav1.ListOptions{
		Watch: true,
	})
	if err != nil {
		return err
	}
	s.watcher = watcher
	s.startListener()
	return nil
}

func (s *kubernetesMemberLookup) startListener() {
	polerpc.Go(common.NewCtxPole(), func(cxt context.Context) {
		for e := range s.watcher.ResultChan() {
			f := func(e watch.Event) {
				defer func() {
					if err := recover(); err != nil {
						sys.LookupLogger.Error(cxt, "%#v", err)
					}
				}()

				endpoints := e.Object.DeepCopyObject().(*v1.Endpoints)
				sets := endpoints.Subsets
				utils.RequireTrue(len(sets) == 1, "only one service can be returned, "+
					"and the result returned is not equal to 1")

				addresses := sets[0].Addresses
				ports := sets[0].Ports

				newMembers := make([]*Member, 0, 0)

				for _, address := range addresses {
					memberHost := address.Hostname
					var httpPort int32
					extensionPorts := make(map[string]int32)

					for _, port := range ports {
						if strings.Compare(port.Name, PoleHttpPort) == 0 {
							httpPort = port.Port
						}
						extensionPorts[port.Name] = port.Port
					}

					newMember := &Member{
						Ip:             memberHost,
						Port:           httpPort,
						ExtensionPorts: extensionPorts,
						Status:         Health,
					}
					newMembers = append(newMembers, newMember)
				}

				s.observer(newMembers)
			}
			f(e)
		}
	})
}

func (s *kubernetesMemberLookup) Observer(observer func(newMembers []*Member)) {
	s.observer = observer
}

func (s *kubernetesMemberLookup) Shutdown() {
	s.watcher.Stop()
}

func (s *kubernetesMemberLookup) Name() string {
	return typeForKubernetesMemberLookup
}

// ==================== kubernetesMemberLookup end ====================

// ==================== FileMemberLookup start ====================

type FileMemberLookup struct {
	BaseMemberLookup
}

func (s *FileMemberLookup) Start() error {
	return notify.RegisterFileWatcher(s.config.GetConfPath(), s)
}

func (s *FileMemberLookup) Observer(observer func(newMembers []*Member)) {
	s.observer = observer
}

func (s *FileMemberLookup) Shutdown() {
	_ = notify.RemoveFileWatcher(s.config.GetConfPath(), s)
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
	addressPort   int32
	urlPath       string
	isShutdown    bool
}

func (s *addressServerMemberLookup) Start() error {
	s.addressServer = sys.GetEnvHolder().ClusterCfg.LookupCfg.AddressLookupCfg.ServerAddr
	s.addressPort = sys.GetEnvHolder().ClusterCfg.LookupCfg.AddressLookupCfg.ServerPort
	s.urlPath = sys.GetEnvHolder().ClusterCfg.LookupCfg.AddressLookupCfg.ServerPath

	return nil
}

func (s *addressServerMemberLookup) startWatchAddressServer() {
	polerpc.DoTickerSchedule(s.ctx, func() {
		url := utils.BuildHttpsUrl(utils.BuildServerAddr(s.addressServer, s.addressPort), s.urlPath)
		for {
			if s.isShutdown {
				s.ctx.Done()
			} else {
				s.getClusterInfoFromServer(url)
			}
		}
	}, time.Duration(5)*time.Second)
}

func (s *addressServerMemberLookup) getClusterInfoFromServer(url string) {
	resp, err := http.Get(url)
	if err != nil {
		sys.LookupLogger.Error(s.ctx, "get cluster.conf from address-server has error  : %s", err)
	} else {
		if resp.StatusCode == http.StatusOK {
			ss := utils.ReadAllLines(resp.Body)
			s.observer(MultiParse(ss...))
		} else {
			sys.LookupLogger.Error(s.ctx, "get cluster.conf from address-server failed : %s",
				utils.ReadContent(resp.Body))
		}
	}
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
