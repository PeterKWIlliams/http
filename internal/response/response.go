package response

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/PeterKWIlliams/http/internal/headers"
)

type StatusCode int

const (
	OK                  = StatusCode(200)
	BadRequest          = StatusCode(400)
	InternalServerError = StatusCode(500)
)

type writerState int

const (
	writeSL writerState = iota
	writeHD
	writeBOD
)

var statusText = map[StatusCode]string{
	OK:                  "OK",
	BadRequest:          "Bad Request",
	InternalServerError: "Server Error",
}

type Writer struct {
	Writer      io.Writer
	writerState writerState
}

var outOfOrderCall = errors.New("out of order call")

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != writeSL {
		return outOfOrderCall
	}
	reasonPhrase, found := statusText[statusCode]
	if !found {
		reasonPhrase = ""
	}
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase)
	_, err := w.Writer.Write([]byte(statusLine))
	w.writerState = writeHD
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	defaultContentLen := strconv.Itoa(contentLen)
	defaultConnection := "close"
	defaultContentType := "text/plain"

	headers["content-length"] = defaultContentLen
	headers["connection"] = defaultConnection
	headers["content-type"] = defaultContentType

	return headers
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writerState != writeHD {
		return outOfOrderCall
	}
	for fieldName, fieldValue := range headers {
		res := fmt.Sprintf("%s: %s\r\n", fieldName, fieldValue)
		_, err := w.Writer.Write([]byte(res))
		if err != nil {
			return fmt.Errorf("error writing headers: %w", err)
		}
	}
	_, err := w.Writer.Write([]byte("\r\n"))
	if err != nil {
		return fmt.Errorf("error writing final header CRLF: %w", err)
	}
	w.writerState = writeBOD
	return nil
}

func (w *Writer) WriteBody(body []byte) (int, error) {
	if w.writerState != writeBOD {
		return 0, outOfOrderCall
	}
	n, err := w.Writer.Write(body)
	if err != nil {
		return 0, fmt.Errorf("error writing body %w", err)
	}
	w.writerState = writeSL
	return n, nil
}

func (w *Writer) Write(statusCode StatusCode, headers headers.Headers, body []byte) error {
	err := w.WriteStatusLine(statusCode)
	if err != nil {
		return err
	}
	err = w.WriteHeaders(headers)
	if err != nil {
		return err
	}
	_, err = w.WriteBody(body)
	if err != nil {
		return err
	}
	return err
}
