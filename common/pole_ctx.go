package common

import (
	"context"
	"time"
)

const (
	ModuleLabel = "Module-Label"
	RequestID   = "Request-ID"
)

type ContextPole struct {
	parent *ContextPole
	ctx    context.Context
	Values map[interface{}]interface{}
}

func (c *ContextPole) Deadline() (deadline time.Time, ok bool) {
	return c.ctx.Deadline()
}

func (c *ContextPole) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *ContextPole) Err() error {
	return c.ctx.Err()
}

func (c *ContextPole) Value(key interface{}) interface{} {
	if v, exist := c.Values[key]; exist {
		return v
	}
	if v := c.ctx.Value(key); v != nil {
		return v
	}
	if c.parent != nil {
		return c.parent.Value(key)
	}
	return nil
}

func (c *ContextPole) NewSubCtx() *ContextPole {
	ctx, _ := context.WithCancel(c.ctx)
	subCtx := &ContextPole{
		parent: c,
		ctx:    ctx,
		Values: make(map[interface{}]interface{}),
	}
	return subCtx
}
