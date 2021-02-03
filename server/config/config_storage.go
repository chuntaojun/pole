package config

import (
	"github.com/pole-group/pole/plugin"
	"github.com/pole-group/pole/server/storage"
)

type ConfigQuery struct {
	Namespace string
	Group     string
	FileName  string
}

type ConfigStorageOperator interface {
	plugin.Plugin

	CreateConfig(cfg *ConfigFile)

	ModifyConfig(cfg *ConfigFile)

	DeleteConfig(cfg *ConfigFile)

	BatchCreateConfig(cfgs []*ConfigFile)

	BatchModifyConfig(cfgs []*ConfigFile)

	BatchRemoveConfig(cfgs []*ConfigFile)

	FindOneConfig(query ConfigQuery) *ConfigFile

	FindConfigs(query ConfigQuery) *[]ConfigFile

	CreateBetaConfig(cfg *ConfigBetaFile)

	ModifyBetaConfig(cfg *ConfigBetaFile)

	DeleteBetaConfig(cfg *ConfigBetaFile)

	FindOneBetaConfig(query ConfigQuery) *ConfigBetaFile

	CreateHistoryConfig(cfg *ConfigHistoryFile)

	DeleteHistoryConfig(cfg *ConfigHistoryFile)

	FindOneHistoryConfig(query ConfigQuery) *ConfigHistoryFile
}

type EmbeddedRdsConfigStorage struct {
	rds storage.Rds
}

type ExternalRdsConfigStorage struct {
}
