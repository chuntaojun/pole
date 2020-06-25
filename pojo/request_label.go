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
