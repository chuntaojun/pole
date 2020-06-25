// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pojo

const (
	InstanceRegisterReq = iota
	InstanceDeregisterReq
	InstanceUpdateReq
	InstanceHeartBeatReq

	ClusterCreateReq
	ClusterUpdateReq
	ClusterDeleteReq

	ServiceCreateReq
	ServiceUpdateReq
	ServiceDeleteReq

	InstanceListReq
	ServiceListReq

	ServiceChangeReq
)

const (
	ConfigPublishReq = iota
	ConfigDeleteReq
	ConfigWatchReq
)

const (
	DistroSyncReq = iota
)
