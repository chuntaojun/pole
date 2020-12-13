package config

import "github.com/Conf-Group/pole/pojo"

type ConfigMetadata struct {
	Namespace string
	Group     string
	FileName  string
	IsEncrypt bool
	Solt      string
}

type ConfigFile struct {
	ConfigMetadata
	Type           pojo.FileType
	Version        int64
	Desc           string
	LastModifyUser string
	Content        string
}

type ConfigTmpFile struct {
	ConfigFile
}

type ConfigBetaFile struct {
	ConfigFile
	BetaClients []string
}

type ConfigHistoryFile struct {
	Operation int8
}
