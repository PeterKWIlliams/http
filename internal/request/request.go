package request

import (
	"errors"
	"fmt"
	"io"
	"strconv"
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
	RequestLine   RequestLine
	Headers       headers.Headers
	Body          []byte
	State         requestState
	Contentlength int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := &Request{
		State:   initialized,
		Headers: headers.NewHeaders(),
	}
	buffer := make([]byte, bufferSize)
	var readToIndex int

	for request.State != done {
		if readToIndex == len(buffer) {
			newBuffer := make([]byte, len(buffer)*2)
			copy(newBuffer, buffer)
			buffer = newBuffer
		}
		r, err := reader.Read(buffer[readToIndex:])
		if err == io.EOF {
			if request.validBody() {
				return nil, errors.New("invalid body size")
			}

			request.State = done
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error getting request from reader: %s", err)
		}
		readToIndex += r

		p, err := request.parse(buffer[:readToIndex])
		if err != nil {
			return nil, fmt.Errorf("error parsing request: %s", err)
		}

		copy(buffer, buffer[p:readToIndex])
		readToIndex -= p
	}

	return request, nil
}

func (r *Request) validBody() bool {
	return r.BodyLength() != int64(r.Contentlength)
}

func (r *Request) parse(data []byte) (int, error) {
	var totalBytesParsed int
	for r.State != done {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, fmt.Errorf("could not parse request component: %s", err)
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
	switch r.State {
	case initialized:
		rl, consumedBytes, err := parseRequestLine(data)
		if err != nil {
			return 0, fmt.Errorf("could not parse request line: %s", err)
		}
		if consumedBytes == 0 {
			return 0, nil
		}
		r.RequestLine = *rl
		r.State = parsingHeaders
		return consumedBytes, nil
	case parsingHeaders:
		n, finished, err := r.Headers.Parse(data)
		if err != nil {
			return 0, fmt.Errorf("error parsing header %s", err)
		}
		if n == 0 {
			return 0, nil
		}
		if finished {
			r.State = parsingBody
			c, err := r.Headers.Get("Content-Length")
			if err != nil {
				r.State = done
				return n, nil
			}
			contentLength, err := strconv.ParseInt(c, 10, 64)
			if err != nil {
				r.State = done
				return 0, errors.New("invalid content-length: NaN")
			}

			r.Contentlength = int(contentLength)
			r.State = parsingBody
		}
		return n, nil
	case parsingBody:
		bytesParsed := len(data)
		r.Body = append(r.Body, data...)

		if r.BodyLength() > int64(r.Contentlength) {
			return 0, errors.New("error parsing body: invalid body size")
		}

		if r.BodyLength() == int64(r.Contentlength) {
			r.State = done
			return bytesParsed, nil
		}
		return bytesParsed, nil
	case done:
		return 0, errors.New("parsing done")
	default:
		return 0, errors.New("unknown state")
	}
}

func (r *Request) BodyLength() int64 {
	return int64(len(r.Body))
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
		return nil, consumedBytes, fmt.Errorf("invalid number of parts in request line: %s", rlString)
	}

	method := requestLineParts[0]
	if _, exists := supportedMethods[method]; !exists {
		return nil, consumedBytes, fmt.Errorf("invalid HTTP method: %s", method)
	}

	httpVersion := requestLineParts[2]
	if httpVersion != "HTTP/1.1" {
		return nil, consumedBytes, fmt.Errorf("unsupported HTTP version: %s", httpVersion)
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: requestLineParts[1],
		HttpVersion:   "1.1",
	}, consumedBytes, nil
}
