//  Copyright (c) 2020, pole-group. All rights reserved.
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.

package plugin

import (
	"context"
)

type Plugin interface {
	Name() string

	Init(ctx context.Context) error

	Run()

	Destroy()
}

type TransportPlugin interface {
	Plugin
}

type StoragePlugin interface {
	Plugin
}
