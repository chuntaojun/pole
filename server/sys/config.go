// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sys

import (
	"io/ioutil"
	"path/filepath"
	"sync/atomic"

	"gopkg.in/yaml.v2"
)

const (
	APIVersion                = "/api/v1"
	EnvMemberMaxAccessFailCnt = "conf.cluster.member.max-access-fail"
)

var propertiesHolder atomic.Value

func init() {
	propertiesHolder = atomic.Value{}
	s, err := ioutil.ReadFile("./conf/pole.yaml")
	if err != nil {
		panic(err)
	}

	p := &Properties{}
	if err := yaml.Unmarshal(s, p); err != nil {
		panic(err)
	}
	propertiesHolder.Store(p)
}

func GetEnvHolder() *Properties {
	return propertiesHolder.Load().(*Properties)
}

type Properties struct {
	BaseDir        string `yaml:"base_dir"`
	StandaloneMode bool   `yaml:"standalone_mode"`
	ConfPath       string `yaml:"conf_path"`
	DataPath       string `yaml:"data_path"`
	IsEmbedded     bool   `yaml:"is_embedded"`
	DriverType     string `yaml:"driver_type"`
	OpenSSL        bool   `yaml:"open_ssl"`

	// Cluster config
	ClusterCfg ClusterConfig `yaml:"cluster_config"`

	// Open port information

	// User-level aware ports
	HttpPort      int64 `yaml:"http_port"`
	DiscoveryPort int64 `yaml:"discovery_port"`
	ConfigPort    int64 `yaml:"config_port"`
	// The internal port
	DistroPort int64 `yaml:"distro_port"`
	RaftPort   int64 `yaml:"raft_port"`
}

type ClusterConfig struct {
	LookupCfg MemberLookupConfig `yaml:"lookup_cfg"`
}

type MemberLookupConfig struct {
	MaxProbeFailCnt  int32                  `yaml:"max_probe_fail"`
	MemberLookupType string                 `yaml:"lookup_type"`
	AddressLookupCfg AddressLookupConfig    `yaml:"address_lookup"`
	K8sLookupCfg     KubernetesLookupConfig `yaml:"k8s_lookup"`
}

type KubernetesLookupConfig struct {
}

type AddressLookupConfig struct {
	ServerAddr string `yaml:"address_server"`
	ServerPort uint64 `yaml:"server_port"`
	ServerPath string `yaml:"server_path"`
}

// linux cgroup 参数配置
type CGroupConfig struct {

}

func (c *Properties) IsStandaloneMode() bool {
	return c.StandaloneMode
}

func (c *Properties) GetBaseDir() string {
	return c.BaseDir
}

func (c *Properties) GetConfPath() string {
	if c.ConfPath == "" {
		c.ConfPath = filepath.Join(c.BaseDir, "conf")
	}
	return c.ConfPath
}

func (c *Properties) GetDataPath() string {
	if c.DataPath == "" {
		c.DataPath = filepath.Join(c.BaseDir, "data")
	}
	return c.DataPath
}

func (c *Properties) GetClusterConfPath() string {
	return filepath.Join(c.GetConfPath(), "cluster.conf")
}
