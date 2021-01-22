// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package consistency

type ProtocolConfig interface {
	Get(key string) string

	GetOrDefault(key, defaultVal string) string
}

type RaftConfig struct {
	Parameters map[string]string `yaml:"parameters"`
}

func (raftCfg *RaftConfig) Get(key string) string {
	return raftCfg.Parameters[key]
}

func (raftCfg *RaftConfig) GetOrDefault(key, defaultVal string) string {
	if val, exist := raftCfg.Parameters[key]; exist {
		return val
	}
	return defaultVal
}

type DistroConfig struct {
	Parameters map[string]string `yaml:"parameters"`
}

func (distroCfg *DistroConfig) Get(key string) string {
	return distroCfg.Parameters[key]
}

func (distroCfg *DistroConfig) GetOrDefault(key, defaultVal string) string {
	if val, exist := distroCfg.Parameters[key]; exist {
		return val
	}
	return defaultVal
}