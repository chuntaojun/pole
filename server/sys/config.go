// Copyright (c) 2020, pole-group. All rights reserved.
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

func InitConf() {
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
	BaseDir        string   `yaml:"baseDir"`
	StandaloneMode bool     `yaml:"standaloneMode"`
	ConfPath       string   `yaml:"confPath"`
	DataPath       string   `yaml:"dataPath"`
	IsEmbedded     bool     `yaml:"isEmbedded"`
	DriverType     string   `yaml:"driverType"`
	OpenSSL        bool     `yaml:"openSsl"`
	DbCfg          DBConfig `yaml:"dbCfg"`

	// Cluster config
	ClusterCfg ClusterConfig `yaml:"clusterConfig"`

	// Open port information

	// User-level aware ports
	HttpPort      int64 `yaml:"httpPort"`
	DiscoveryPort int64 `yaml:"discoveryPort"`
	ConfigPort    int64 `yaml:"configPort"`
	// The internal port
	DistroPort int64 `yaml:"distroPort"`
	RaftPort   int64 `yaml:"raftPort"`
}

type ClusterConfig struct {
	LookupCfg MemberLookupConfig `yaml:"lookupCfg"`
}

type MemberLookupConfig struct {
	MaxProbeFailCnt  int32                  `yaml:"maxProbeFail"`
	MemberLookupType string                 `yaml:"lookupType"`
	AddressLookupCfg AddressLookupConfig    `yaml:"addressLookup"`
	K8sLookupCfg     KubernetesLookupConfig `yaml:"k8sLookup"`
}

type KubernetesLookupConfig struct {
	Namespace string `yaml:"namespace"`
}

type AddressLookupConfig struct {
	ServerAddr string `yaml:"addressServer"`
	ServerPort int32 `yaml:"serverPort"`
	ServerPath string `yaml:"serverPath"`
}

// linux cgroup 参数配置
type CGroupConfig struct {
}

type DBConfig struct {
	DbHost   string `yaml:"dbHost"`
	DbPort   int32  `yaml:"dbPort"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
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
