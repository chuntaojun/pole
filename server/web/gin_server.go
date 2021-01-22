// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
	"github.com/gin-gonic/gin"

	"github.com/pole-group/pole/common"
)

type WebServer struct {
	server *gin.Engine
}

func (ws *WebServer) Init(ctx *common.ContextPole)  {

}

func (ws *WebServer) Shutdown() {

}
