package main

import (
	"github.com/Conf-Group/pole/common"
)

type LifeCycle interface {

	Init(ctx *common.ContextPole)

	Shutdown()

}
