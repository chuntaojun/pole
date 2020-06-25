// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cluster

import (
	"fmt"
	"nacos-go/sys"
	"nacos-go/utils"
)

type MemberLookup interface {

	Start() error

	Observer(observer func(newMembers []*Member))

	Shutdown()

	Name() string

}

const (
	memberLookupKey = "conf.core.cluster.member.lookup"

	standaloneMemberLookup = "standalone"

	fileMemberLookup = "file"

	addressServerMemberLookup = "address-server"
)

func CreateMemberLookup(config *sys.Config, observer func(newMembers []*Member)) (MemberLookup, error) {
	var lookup MemberLookup
	if config.IsStandaloneMode() {
		lookup = &StandaloneMemberLookup{}
	} else {
		val := utils.GetStringFromEnv(memberLookupKey)
		switch val {
		case fileMemberLookup:
			lookup = &FileMemberLookup{}
		case addressServerMemberLookup:
			lookup = &AddressServerMemberLookup{}
		default:
			panic(fmt.Errorf("this member-lookup [%s] unrealized", val))
		}
	}

	lookup.Observer(observer)
	err := lookup.Start()
	return lookup, err
}

func SwitchMemberLookupAndCloseOld(name string, config *sys.Config, oldLookup MemberLookup, observer func(newMembers []*Member)) (MemberLookup, error) {

	if config.IsStandaloneMode() {
		return nil, fmt.Errorf("only support in cluster mode")
	}

	oldLookup.Shutdown()
	var newLookup MemberLookup

	switch name {
	case fileMemberLookup:
		newLookup = &FileMemberLookup{}
	case addressServerMemberLookup:
		newLookup = &AddressServerMemberLookup{}
	default:
		panic(fmt.Errorf("this member-lookup [%s] unrealized", name))
	}

	newLookup.Observer(observer)
	err := newLookup.Start()
	return newLookup, err
}

type StandaloneMemberLookup struct {

}

func (s *StandaloneMemberLookup) Start() error {
	return nil
}

func  (s *StandaloneMemberLookup) Observer(observer func(newMembers []*Member))  {

}

func (s *StandaloneMemberLookup) Shutdown() {
}

func (s *StandaloneMemberLookup) Name() string {
	return standaloneMemberLookup
}

type FileMemberLookup struct {

}


func (s *FileMemberLookup) Start() error {
	return nil
}

func  (s *FileMemberLookup) Observer(observer func(newMembers []*Member))  {

}

func (s *FileMemberLookup) Shutdown() {
}

func (s *FileMemberLookup) Name() string {
	return fileMemberLookup
}

type AddressServerMemberLookup struct {

}


func (s *AddressServerMemberLookup) Start() error {
	return nil
}

func  (s *AddressServerMemberLookup) Observer(observer func(newMembers []*Member))  {

}


func (s *AddressServerMemberLookup) Shutdown() {
}

func (s *AddressServerMemberLookup) Name() string {
	return addressServerMemberLookup
}

