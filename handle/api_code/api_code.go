package api_code

type ApiCode int

type Response struct {
	Code    ApiCode `json:"code"`
	Message string  `json:"message"`
}

func NewResponse(code ApiCode, message string) *Response {
	return &Response{
		Code:    code,
		Message: message,
	}
}

const (
	InternalServerErr ApiCode = 500
	InvalidParams     ApiCode = 10000
)
