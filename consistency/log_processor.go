// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package consistency

type LogProcessor interface {
	OnRequest()

	OnApply()

	OnError()

	Group() string
}

type LogProcessor4AP interface {
	LogProcessor
}

type LogProcessor4CP interface {
	LogProcessor
}
