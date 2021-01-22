// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type Conf struct {
}

func Init() *Conf {
	n := new(Conf)
	return n
}

func (n *Conf) Start() {
}

func (n *Conf) initDiscovery() {
}
