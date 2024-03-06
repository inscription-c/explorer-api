package api_code

import "encoding/json"

type ApiCode int

type Response struct {
	Code    ApiCode `json:"code"`
	Message string  `json:"message"`
}

func (r *Response) Error() string {
	data, _ := json.Marshal(r)
	return string(data)
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
