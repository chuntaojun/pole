package plugin

import (
	"github.com/Conf-Group/pole/common"
)

type Plugin interface {
	Name() string

	Init(ctx *common.ContextPole)

	Run()

	Destroy()
}

type TransportPlugin interface {
	Plugin
}

type StoragePlugin interface {
	Plugin
}
