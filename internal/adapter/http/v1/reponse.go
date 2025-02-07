package v1

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type ErrorBody struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

var (
	InternalError = ErrorBody{Code: 500, Msg: "Internal Error"}
	InvalidParams = ErrorBody{Code: 400, Msg: "Invalid Params"}
	Unauthorized  = ErrorBody{Code: 401, Msg: "Unauthorized"}
	NotFound      = ErrorBody{Code: 404, Msg: "Not Found"}
)

func NewResponse(code int, msg string, data interface{}) *Response {
	return &Response{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}

func Success(data interface{}) *Response {
	return NewResponse(0, "success", data)
}

func Fail(e ErrorBody) *Response {
	return NewResponse(e.Code, e.Msg, nil)
}
