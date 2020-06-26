// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pojo

type RestResult struct {
	Code int         `json:"code"`
	Body interface{} `json:"body,omitempty"`
	Msg  string      `json:"msg"`
}
