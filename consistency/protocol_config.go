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
