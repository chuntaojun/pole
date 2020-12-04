// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package client

type Config struct {
	ServerAddr    string
	Endpoint      string
	Token         string
	NamespaceId   string
	DiscoveryAddr []string
	ConfigAddr    []string
}
