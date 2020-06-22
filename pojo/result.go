package pojo

type RestResult struct {
	Code int         `json:"code"`
	Body interface{} `json:"body,omitempty"`
	Msg  string      `json:"msg"`
}
