package naming

import "github.com/gin-gonic/gin"

type Naming struct {
	
}

func NewNaming(r *gin.Engine) *Naming {
	naming := Naming{}
	naming.initWebHandler()
	naming.initGrpcHandler()
	return &naming
}

func (n *Naming) initWebHandler() {

}

func (n *Naming) initGrpcHandler() {

}