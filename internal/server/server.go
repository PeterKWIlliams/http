package server

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync/atomic"
)

type Server struct {
	Addr     string
	listener net.Listener
	isClosed atomic.Bool
}

type Handler struct{}

func Serve(port int) (*Server, error) {
	listenAddr := ":" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, err
	}
	server := &Server{
		listener: listener,
		Addr:     listenAddr,
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

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	response := "HTTP/1.1 200 OK\r\n" +
		"Content-Length: 12\r\n" +
		"Content-Type: text/plain\r\n" +
		"\r\n" +
		"Hello World!"

	n, err := conn.Write([]byte(response))
	if err != nil {
		log.Printf("Error writing response to %s: %v", conn.RemoteAddr(), err)
	}

	fmt.Printf("Wrote %d bytes to client\n", n)
}
