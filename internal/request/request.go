package request

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/PeterKWIlliams/http/internal/headers"
)

const bufferSize = 8

type requestState int

const (
	initialized requestState = iota
	parsingHeaders
	parsingBody
	done
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	state       requestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := &Request{
		state:   initialized,
		Headers: headers.NewHeaders(),
	}
	buffer := make([]byte, bufferSize)
	var readToIndex int

	for {
		if request.state == done {
			break
		}
		if readToIndex == len(buffer) {
			newBuffer := make([]byte, len(buffer)*2)
			copy(newBuffer, buffer)
			buffer = newBuffer
		}
		r, err := reader.Read(buffer[readToIndex:])
		if err == io.EOF {
			request.state = done
			break
		}
		if err != nil {
			return nil, fmt.Errorf("Error getting request from reader: %s", err)
		}
		readToIndex += r

		p, err := request.parse(buffer[:readToIndex])
		if err != nil {
			return nil, fmt.Errorf("Error parsing request: %s", err)
		}

		copy(buffer, buffer[p:readToIndex])
		readToIndex -= p
	}

	return request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	var totalBytesParsed int
	for r.state != done && r.state != parsingBody {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, fmt.Errorf("Could not parse request component: %s", err)
		}
		if n == 0 {
			break
		}
		totalBytesParsed += n
		if len(data) == totalBytesParsed {
			break
		}
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case initialized:
		rl, consumedBytes, err := parseRequestLine(data)
		if err != nil {
			return 0, fmt.Errorf("Could not parse request line: %s", err)
		}
		if consumedBytes == 0 {
			return 0, nil
		}
		r.RequestLine = *rl
		r.state = parsingHeaders
		return consumedBytes, nil
	case parsingHeaders:
		n, finished, err := r.Headers.Parse(data)
		if err != nil {
			return 0, fmt.Errorf("Error parsing header %s", err)
		}
		if n == 0 {
			return 0, nil
		}
		if finished {
			r.state = done
			return n, nil
		}
		return n, nil
	case done:
		return 0, errors.New("Parsing done")
	default:
		return 0, errors.New("unknown state")
	}
}

var supportedMethods = map[string]struct{}{
	"GET":    {},
	"POST":   {},
	"DELETE": {},
	"HEAD":   {},
	"PUT":    {},
}

func parseRequestLine(rl []byte) (*RequestLine, int, error) {
	rlString := string(rl)

	if !strings.Contains(rlString, "\r\n") {
		return nil, 0, nil
	}

	index := strings.Index(rlString, "\r\n")
	consumedBytes := index + 2
	rlString = rlString[:index]

	requestLineParts := strings.Fields(rlString)
	if len(requestLineParts) != 3 {
		return nil, consumedBytes, fmt.Errorf("Invalid number of parts in request line: %s", rlString)
	}

	method := requestLineParts[0]
	if _, exists := supportedMethods[method]; !exists {
		return nil, consumedBytes, fmt.Errorf("Invalid HTTP method: %s", method)
	}

	httpVersion := requestLineParts[2]
	if httpVersion != "HTTP/1.1" {
		return nil, consumedBytes, fmt.Errorf("Unsupported HTTP version: %s", httpVersion)
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: requestLineParts[1],
		HttpVersion:   "1.1",
	}, consumedBytes, nil
}

// func (r *Request) parse(data []byte) (int, error) { if r.state == done {
// 		return 0, errors.New("parsing is complete")
// 	}
// 	if r.state != initialized {
// 		return 0, errors.New("unknown state")
// 	}
//
// 	rl, consumedBytes, err := parseRequestLine(data)
// 	if err != nil {
// 		return 0, fmt.Errorf("Could not parse request line: %s", err)
// 	}
// 	if consumedBytes == 0 {
// 		return 0, nil
// 	}
// 	r.state = parsingHeaders
//
// 	r.state = done
// 	r.RequestLine = *rl
// 	return consumedBytes, nil
// }
