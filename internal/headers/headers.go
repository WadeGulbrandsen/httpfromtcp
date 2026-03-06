package headers

import (
	"bytes"
	"fmt"
	"slices"
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
	if !validTokens(fieldName) {
		return 0, false, fmt.Errorf("Invalid header token in field name: %v", fieldName)
	}
	fieldValue = strings.TrimSpace(fieldValue)
	h[fieldName] = fieldValue
	return idx + 2, false, nil
}

var tokenChars = []rune("!#$%&'*+-.^_`|~")

func validTokens(s string) bool {
	for _, c := range s {
		if !isTokenChar(c) {
			return false
		}
	}
	return true
}

func isTokenChar(c rune) bool {
	if c >= 'A' && c <= 'Z' ||
		c >= 'a' && c <= 'z' ||
		c >= '0' && c <= '9' {
		return true
	}
	return slices.Contains(tokenChars, c)
}
