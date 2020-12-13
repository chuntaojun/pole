package config

type ConfigQuery struct {
	Namespace string
	Group     string
	FileName  string
}

type ConfigStorageOperator interface {
	CreateConfig(cfg *ConfigFile)

	ModifyConfig(cfg *ConfigFile)

	DeleteConfig(cfg *ConfigFile)

	BatchCreateConfig(cfgs []*ConfigFile)

	BatchModifyConfig(cfgs []*ConfigFile)

	BatchRemoveConfig(cfgs []*ConfigFile)

	FindOneConfig(query ConfigQuery) *ConfigFile

	FindConfigs(query ConfigQuery) *[]ConfigFile
}

type ConfigBetaStorageOperator interface {
	CreateBetaConfig(cfg *ConfigBetaFile)

	ModifyBetaConfig(cfg *ConfigBetaFile)

	DeleteBetaConfig(cfg *ConfigBetaFile)

	FindOneConfig(query ConfigQuery) *ConfigBetaFile
}
