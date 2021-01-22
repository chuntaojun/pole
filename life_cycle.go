package main

import (
	"github.com/pole-group/pole/common"
)

type LifeCycle interface {

	Init(ctx *common.ContextPole)

	Shutdown()

}
