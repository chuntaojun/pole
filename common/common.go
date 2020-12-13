// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import "sync"

const (
	PoleContextKey = "pole_context"
)

type PoleContext struct {
	lock     sync.RWMutex
	metadata map[string]interface{}
}
