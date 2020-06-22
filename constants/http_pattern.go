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
