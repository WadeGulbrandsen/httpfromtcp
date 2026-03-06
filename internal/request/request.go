package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
	state       state
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type state int

const (
	initialized state = iota + 1
	done
)

const crlf = "\r\n"

const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := &Request{state: initialized}
	buffer := make([]byte, bufferSize)
	readToIdx := 0
	for {
		if req.state == done {
			break
		}
		if readToIdx >= len(buffer) {
			new_buffer := make([]byte, len(buffer)*2)
			copy(new_buffer, buffer)
			buffer = new_buffer
		}
		r, err := reader.Read(buffer[readToIdx:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				req.state = done
				break
			}
			return nil, err
		}
		readToIdx += r
		parsed, err := req.parse(buffer[:readToIdx])
		if err != nil {
			return nil, err
		}
		copy(buffer, buffer[parsed:])
		readToIdx -= parsed
	}
	return req, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.state {
	case initialized:
		requestLine, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.state = done
		return n, nil
	case done:
		return 0, errors.New("Error: Trying to read data in a done state")
	default:
		return 0, errors.New("Error: Unknown state")
	}
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}
	first_line := string(data[:idx])
	requestLine, err := requestLineFromString(first_line)
	if err != nil {
		return nil, 0, err
	}
	return requestLine, idx + 2, nil
}

func requestLineFromString(s string) (*RequestLine, error) {
	parts := strings.Split(s, " ")
	if len(parts) != 3 {
		return nil, errors.New("Malformed request")
	}

	method := parts[0]

	for _, char := range method {
		if char < 'A' || char > 'Z' {
			return nil, fmt.Errorf("Invalid method: %v", method)
		}
	}

	target := parts[1]

	httpParts := strings.Split(parts[2], "/")
	if len(httpParts) != 2 {
		return nil, errors.New("Malformed request")
	}

	httpName := httpParts[0]
	httpVer := httpParts[1]

	if httpName != "HTTP" {
		return nil, fmt.Errorf("Unsupported HTTP name: %v", httpName)
	}
	if httpVer != "1.1" {
		return nil, fmt.Errorf("Unsupported HTTP version: %v", httpVer)
	}

	return &RequestLine{HttpVersion: httpVer, RequestTarget: target, Method: method}, nil
}
