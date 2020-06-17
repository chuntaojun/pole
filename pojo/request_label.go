package pojo

const (
	InstanceRegister = iota
	InstanceDeregister
	InstanceUpdate
	InstanceHeartBeat

	ClusterCreate
	ClusterUpdate
	ClusterDelete

	ServiceCreate
	ServiceUpdate
	ServiceDelete

	InstanceList
	ServiceList

	ServiceChange
)

const (
	ConfigPublish = iota
	ConfigDelete
	ConfigWatch
)

const (
	DistroSync = iota
)
