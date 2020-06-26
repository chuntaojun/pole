// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/rsocket/rsocket-go/rx/mono"
	"github.com/stretchr/testify/assert"
)

func Test_MonoCreateHasError(t *testing.T) {
	wait := sync.WaitGroup{}
	wait.Add(1)
	reference := atomic.Value{}
	m := mono.Create(func(ctx context.Context, sink mono.Sink) {
		panic(fmt.Errorf("test mono inner panic error"))
	}).DoOnError(func(e error) {
		fmt.Printf("has error : %s\n", e)
		reference.Store(e)
		wait.Done()
	})
	ctx := context.Background()
	m.Subscribe(ctx)
	wait.Wait()
	assert.NotNil(t, reference.Load(), "must not nil")
}
