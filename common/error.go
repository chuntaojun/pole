// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

type ErrorCode int32

type PoleError struct {
	ErrMsg	string
	Code ErrorCode
}

func (pe *PoleError) Error() string  {
	return pe.ErrMsg
}

func (pe *PoleError) ErrCode() ErrorCode  {
	return pe.Code
}
