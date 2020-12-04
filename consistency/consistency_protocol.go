// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package consistency

import "github.com/Conf-Group/pole/server/cluster"

type Protocol interface {
	// 初始化
	Init(cfg ProtocolConfig, memberMgn *cluster.ServerClusterManager)

	// 读数据请求
	Read(req *Read)

	// 写数据请求
	Write(req *Write)

	// 添加请求处理者
	AddProcessor(processors ...RequestProcessor) error

	// 发生节点变更
	OnMemberChange(newMembers []*cluster.Member)

	// 关闭一致性协议
	Shutdown()
}

type RequestProcessor interface {
	// 处理读请求
	OnRead(req *Read) Response

	// 处理写请求
	OnWrite(req *Write) Response

	// 处理错误
	OnError(err error)

	// 分组详细
	Group() (string,uint64)
}

type RequestProcessor4AP interface {
	RequestProcessor
}

type RequestProcessor4CP interface {
	RequestProcessor
}
