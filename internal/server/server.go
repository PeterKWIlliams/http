package server

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/PeterKWIlliams/http/internal/request"
	"github.com/PeterKWIlliams/http/internal/response"
)

type Server struct {
	Addr     string
	listener net.Listener
	isClosed atomic.Bool
	Handler  Handler
}

type Handler func(w *response.Writer, req *request.Request)

func Serve(port int, handler Handler) (*Server, error) {
	listenAddr := ":" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, err
	}
	server := &Server{
		listener: listener,
		Addr:     listenAddr,
		Handler:  handler,
	}

	go func() {
		server.listen()
	}()
	return server, nil
}

func (s *Server) Close() error {
	s.isClosed.Store(true)
	err := s.listener.Close()
	if err != nil {
		return fmt.Errorf("failed to close listener: %w", err)
	}
	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if s.isClosed.Load() {
			if err != nil && !errors.Is(err, net.ErrClosed) {
				log.Printf("Non-ErrClosed error during shutdown %v", err)
			}
			return
		}
		if err != nil {
			log.Printf("Error accepting connection:%v", err)
			continue
		}

		go s.handle(conn)

	}
}

func WriteError(w *response.Writer, statusCode response.StatusCode, message string) error {
	body := []byte(message)
	contentLength := len(body)
	headers := response.GetDefaultHeaders(contentLength)

	err := w.WriteStatusLine(statusCode)
	if err != nil {
		return fmt.Errorf("error writing error: %w", err)
	}
	err = w.WriteHeaders(headers)
	if err != nil {
		return fmt.Errorf("error writing error : %w", err)
	}
	_, err = w.WriteBody(body)
	if err != nil {
		return fmt.Errorf("error writing error: %w", err)
	}
	return nil
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	req, err := request.RequestFromReader(conn)
	resWriter := &response.Writer{
		Writer: conn,
	}
	if err != nil {
		err = WriteError(resWriter, response.BadRequest, "could not process request")
		if err != nil {
			log.Printf("error %s", err)
		}
		return
	}
	s.Handler(resWriter, req)
}
