// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"github.com/pole-group/pole/pojo"
	"github.com/pole-group/pole/utils"
)

type ConfigMetadata struct {
	Namespace string
	Group     string
	FileName  string
	IsEncrypt bool
	Solt      string
}

type ConfigFile struct {
	ConfigMetadata
	FType          pojo.FileType
	Version        int64
	Desc           string
	LastModifyUser string
	Content        string
}

type ConfigTmpFile struct {
	ConfigFile
}

type ConfigBetaFile struct {
	ConfigFile
	BetaClients []string
}

type ConfigHistoryFile struct {
	Operation int8
}

type ConfigChangeEvent struct {
	Namespace string
	Group     string
	FileName  string
	Content   string
	Version   int64
	IsEncrypt bool
	Solt      string
	FType     pojo.FileType
}

func (ce *ConfigChangeEvent) Name() string {
	return "ConfigChangeEvent"
}

func (ce *ConfigChangeEvent) Sequence() int64 {
	return utils.GetCurrentTimeMs()
}

type ConfigBetaChangeEvent struct {
	InnerEvent *ConfigChangeEvent
	clientIds  []string
}

func (ce *ConfigBetaChangeEvent) Name() string {
	return "ConfigBetaChangeEvent"
}

func (ce *ConfigBetaChangeEvent) Sequence() int64 {
	return utils.GetCurrentTimeMs()
}
