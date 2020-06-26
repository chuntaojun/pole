// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sys

import "path/filepath"

const (
	ServerPort = "ServerPort"
	DistroPort = "DistroPort"
	RaftPort   = "RaftPort"
)

type Config struct {
	baseDir        string
	standaloneMode bool
	confPath       string
	dataPath       string
}

func (c *Config) IsStandaloneMode() bool {
	return c.standaloneMode
}

func (c *Config) GetBaseDir() string {
	return c.baseDir
}

func (c *Config) GetConfPath() string {
	if c.confPath == "" {
		c.confPath = filepath.Join(c.baseDir, "conf")
	}
	return c.confPath
}

func (c *Config) GetDataPath() string {
	if c.dataPath == "" {
		c.dataPath = filepath.Join(c.baseDir, "data")
	}
	return c.dataPath
}

func (c *Config) GetClusterConf() string {
	return filepath.Join(c.GetConfPath(), "cluster.conf")
}
