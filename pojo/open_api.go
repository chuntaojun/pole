// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pojo

type HttpResult struct {
	Code int32
	Msg  string
	Body interface{}
}
