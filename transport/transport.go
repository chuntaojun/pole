package transport

import "github.com/Conf-Group/pole/pojo"

type TransportServer interface {

	OnHandler(req *pojo.ServerRequest) *pojo.ServerResponse

}
