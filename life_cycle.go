package main

import "context"

type LifeCycle interface {

	Init(ctx context.Context)

	Shutdown()

}
