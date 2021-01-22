// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sys

import "github.com/pole-group/pole/utils"

var (
	RaftKVLogger utils.Logger = utils.NewTestLogger(utils.LoggerCfg{
		Level:      utils.LogLevelDebug,
		Path:       "",
		MaxSize:    0,
		MaxBackups: 0,
		MaxAge:     0,
		Compress:   false,
	})

	CoreLogger utils.Logger = utils.NewTestLogger(utils.LoggerCfg{
		Level:      utils.LogLevelDebug,
		Path:       "",
		MaxSize:    0,
		MaxBackups: 0,
		MaxAge:     0,
		Compress:   false,
	})

	LookupLogger utils.Logger = utils.NewTestLogger(utils.LoggerCfg{
		Level:      utils.LogLevelDebug,
		Path:       "",
		MaxSize:    0,
		MaxBackups: 0,
		MaxAge:     0,
		Compress:   false,
	})
)


