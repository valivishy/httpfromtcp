package server

import (
	"fmt"
	"net"
)

type Server struct {
	listener net.Listener
	open     bool
}

func Serve(port int) (*Server, error) {
	server := &Server{}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server.listener = listener
	server.open = true

	go server.listen()

	return server, nil
}

func (s *Server) Close() {
	err := s.listener.Close()
	if err != nil {
		panic(err)
	}

	s.listener = nil
	s.open = false
}

func (s *Server) listen() bool {
	for {
		if !s.open {
			return false
		}

		accept, err := s.listener.Accept()
		if err != nil {
			panic(err)
		}

		go s.handle(accept)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			panic(err)
		}
	}(conn)
	if _, err := conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello World!")); err != nil {
		panic(err)
	}
}
