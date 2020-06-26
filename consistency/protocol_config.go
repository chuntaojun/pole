// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package consistency

type ProtocolConfig struct {
	Parameters map[string]string
}

type RaftConfig struct {
	ProtocolConfig
}

type DistroConfig struct {
	ProtocolConfig
}
