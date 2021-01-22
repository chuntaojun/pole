// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"context"
	"time"
)

const (
	ModuleLabel = "Module-Label"
	RequestID   = "Request-ID"
)

var (
	EmptyContext = NewCtxPole()
)

type ContextPole struct {
	parent *ContextPole
	cancel context.CancelFunc
	ctx    context.Context
	Values map[interface{}]interface{}
}

func NewCtxPole() *ContextPole {
	ctx, cancel := context.WithCancel(context.Background())
	return &ContextPole{
		parent: nil,
		ctx:    ctx,
		cancel: cancel,
		Values: make(map[interface{}]interface{}),
	}
}

func (c *ContextPole) Write(key, value interface{}) {
	c.Values[key] = value
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

func (c *ContextPole) Cancel() {
	if c.cancel != nil {
		c.cancel()
	}
}

func (c *ContextPole) NewSubCtx() *ContextPole {
	ctx, cancel := context.WithCancel(c.ctx)
	subCtx := &ContextPole{
		parent: c,
		cancel: cancel,
		ctx:    ctx,
		Values: make(map[interface{}]interface{}),
	}
	return subCtx
}
