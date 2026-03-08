package response

import (
	"fmt"
)

type StatusCode int

const (
	StatusCodeSuccess             StatusCode = 200
	StatusCodeBadRequest          StatusCode = 400
	StatusCodeInternalServerError StatusCode = 500
)

func getStatusLine(statusCode StatusCode) []byte {
	reason := ""
	switch statusCode {
	case StatusCodeSuccess:
		reason = "OK"
	case StatusCodeBadRequest:
		reason = "Bad Request"
	case StatusCodeInternalServerError:
		reason = "Internal Server Error"
	}
	return []byte(fmt.Sprintf("HTTP/1.1 %v %s\r\n", statusCode, reason))
}
