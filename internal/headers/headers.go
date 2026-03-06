package headers

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

const crlf = "\r\n"

type Headers map[string]string

func NewHeaders() Headers {
	return Headers{}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, false, nil
	}
	if idx == 0 {
		return 2, true, nil
	}
	fieldLine := string(data[:idx])
	fieldName, fieldValue, ok := strings.Cut(fieldLine, ":")
	if !ok || strings.TrimRightFunc(fieldName, unicode.IsSpace) != fieldName {
		return 0, false, fmt.Errorf("Invalid header format: %v", fieldLine)
	}
	fieldName = strings.ToLower(strings.TrimSpace(fieldName))
	fieldValue = strings.TrimSpace(fieldValue)
	h[fieldName] = fieldValue
	return idx + 2, false, nil
}
