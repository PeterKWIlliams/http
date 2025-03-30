package request

import (
	"fmt"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req, err := io.ReadAll(reader)
	if err != nil {
		return nil, nil
	}

	requestStr := strings.Split(string(req), "\r\n")
	requestLineStr := requestStr[0]

	requestLine, err := parseRequestLine(requestLineStr)
	if err != nil {
		return nil, fmt.Errorf("Error Getting request from reader: %s", err)
	}

	return &Request{
		RequestLine: *requestLine,
	}, nil
}

var supportedMethods = map[string]struct{}{
	"GET":    {},
	"POST":   {},
	"DELETE": {},
	"HEAD":   {},
	"PUT":    {},
}

func parseRequestLine(rl string) (*RequestLine, error) {
	requestLineParts := strings.Fields(rl)
	if len(requestLineParts) != 3 {
		return nil, fmt.Errorf("invalid number of parts in request line %s", rl)
	}

	method := requestLineParts[0]
	if _, exists := supportedMethods[method]; !exists {
		return nil, fmt.Errorf("invalid HTTP method %s", method)
	}
	requestTarget := requestLineParts[1]

	httpVersion := requestLineParts[2]
	if httpVersion != "HTTP/1.1" {
		return nil, fmt.Errorf("Unsupported HTTP version :%s", httpVersion)
	}
	versionNo := strings.Split(httpVersion, "/")[1]

	return &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   versionNo,
	}, nil
}
