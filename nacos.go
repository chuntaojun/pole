package main

import (
	"github.com/gin-gonic/gin"
	"nacos-go/naming"
)

type Nacos struct {
	router	*gin.Engine
}

func Init() *Nacos {
	n := new(Nacos)
	return n
}

func (n *Nacos) Start() {
	n.initWeb()
}

func (n *Nacos) initWeb()  {
	n.router = gin.Default()
}

func (n *Nacos) initNaming() {
}