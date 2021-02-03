// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cluster

import (
	"math/rand"
	"strconv"
	"strings"

	"github.com/pole-group/pole/server/sys"
	"github.com/pole-group/pole/utils"
)

type ProtocolPort string

const (
	ServerPort    ProtocolPort = "server-port"
	DiscoveryPort ProtocolPort = "discovery-port"
	ConfigPort    ProtocolPort = "config-port"
	CPPort        ProtocolPort = "cp-port"
	APPort        ProtocolPort = "ap-port"
)

func OnSuccess(mgn *ServerClusterManager, m *Member) {
	m.accessFailCnt = 0
	m.Status = Health
	mgn.UpdateMember(m)
}

func OnFail(mgn *ServerClusterManager, m *Member, err error) {
	m.Status = Impeach
	if strings.ContainsAny(err.Error(), "Connection refused") {
		m.Status = Down
	} else {
		m.accessFailCnt++
		if m.accessFailCnt >= sys.GetEnvHolder().ClusterCfg.LookupCfg.MaxProbeFailCnt {
			m.Status = Down
		}
	}
	mgn.UpdateMember(m)
}

func MultiParse(arr ...string) []*Member {
	mList := make([]*Member, len(arr))
	for _, s := range arr {
		mList = append(mList, SingParse(s))
	}
	return mList
}

func SingParse(s string) *Member {
	ss := strings.Split(s, ":")
	ip := ss[0]
	port, err := strconv.ParseInt(ss[1], 10, 32)
	if err != nil {
		panic(err)
	}

	extensionPort := make(map[ProtocolPort]int32)

	extensionPort[ServerPort] = int32(port)
	extensionPort[DiscoveryPort] = int32(port + 100)
	extensionPort[ConfigPort] = int32(port + 200)
	extensionPort[CPPort] = int32(port + 300)
	extensionPort[APPort] = int32(port + 400)

	return &Member{
		Ip:             ip,
		ExtensionPorts: extensionPort,
		MetaData:       make(map[string]string),
	}
}

func KRandomMember(k int, members []*Member, filter func(m *Member) bool) []*Member {

	totalSize := len(members)
	set := utils.NewSet()

	for i := 0; i < 3*totalSize && set.Size() <= k; i++ {
		idx := rand.Intn(totalSize)

		m := members[idx]
		if filter(m) {
			set.Add(m)
		}
	}

	ms := make([]*Member, set.Size())

	index := 0

	set.Range(func(value interface{}) {
		ms[index] = value.(*Member)
		index++
	})

	return ms
}
