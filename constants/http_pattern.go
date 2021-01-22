// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package constants

// basic

const (
	BasePattern = "/conf/v1"
)

// server

const (
	ClusterBasePattern        = BasePattern + "/cluster"
	MemberInfoReportPattern   = ClusterBasePattern + "/member/info/report"
	MemberListPattern         = ClusterBasePattern + "/members"
	MemberLookupNowPattern    = ClusterBasePattern + "/lookup"
	MemberLookupSwitchPattern = ClusterBasePattern + "/lookup/switch"
)

// config

// discovery
