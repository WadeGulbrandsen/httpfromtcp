package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/WadeGulbrandsen/httpfromtcp/internal/headers"
)

type Request struct {
	RequestLine    RequestLine
	Headers        headers.Headers
	Body           []byte
	state          requestState
	bodyLengthRead int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type requestState int

const (
	requestStateInitialized requestState = iota + 1
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

const crlf = "\r\n"
const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := &Request{
		state:   requestStateInitialized,
		Headers: headers.NewHeaders(),
		Body:    make([]byte, 0),
	}
	buffer := make([]byte, bufferSize)
	readToIdx := 0
	for req.state != requestStateDone {
		if readToIdx >= len(buffer) {
			new_buffer := make([]byte, len(buffer)*2)
			copy(new_buffer, buffer)
			buffer = new_buffer
		}

		r, err := reader.Read(buffer[readToIdx:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.state != requestStateDone {
					return nil, fmt.Errorf("incomplete request, in state: %d, read n bytes on EOF: %d", req.state, r)
				}
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
	totalBytesParsed := 0
	for r.state != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		totalBytesParsed += n
		if n == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateInitialized:
		requestLine, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.state = requestStateParsingHeaders
		return n, nil
	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.state = requestStateParsingBody
		}
		return n, nil
	case requestStateParsingBody:
		contentLengthStr, ok := r.Headers.Get("Content-Length")
		if !ok {
			r.state = requestStateDone
			return len(data), nil
		}
		contentLength, err := strconv.Atoi(contentLengthStr)
		if err != nil {
			return 0, fmt.Errorf("Malformed Content-Length: %v", err)
		}
		r.Body = append(r.Body, data...)
		r.bodyLengthRead += len(data)
		if r.bodyLengthRead > contentLength {
			return 0, fmt.Errorf("Content-Length %v exceeded. Body read so far: %v", contentLength, r.bodyLengthRead)
		}
		if r.bodyLengthRead == contentLength {
			r.state = requestStateDone
		}
		return len(data), nil
	case requestStateDone:
		return 0, errors.New("Error: Trying to read data in a done state")
	default:
		return 0, fmt.Errorf("Unknown state")
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
		return nil, fmt.Errorf("Poorly formatted request-line: %s", s)
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
		return nil, fmt.Errorf("malformed start-line: %s", s)
	}

	httpName := httpParts[0]
	if httpName != "HTTP" {
		return nil, fmt.Errorf("Unsupported HTTP name: %v", httpName)
	}
	httpVer := httpParts[1]
	if httpVer != "1.1" {
		return nil, fmt.Errorf("Unsupported HTTP version: %v", httpVer)
	}

	return &RequestLine{HttpVersion: httpVer, RequestTarget: target, Method: method}, nil
}
