package transport

import (
	"github.com/Conf-Group/pole/common"
	"github.com/Conf-Group/pole/pojo"
)

type ServerHandler func(cxt *common.ContextPole, req *pojo.ServerRequest) *pojo.RestResult

type TransportServer interface {
	RegisterRequestHandler(path string, handler ServerHandler)

	RegisterStreamRequestHandler(path string, handler ServerHandler)
}

func ParseErrorToResult(err *common.PoleError) *pojo.RestResult {
	return &pojo.RestResult{
		Code: int32(err.ErrCode()),
		Msg:  err.Error(),
	}
}
