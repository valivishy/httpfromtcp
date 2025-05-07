package server

import (
	"bytes"
	"fmt"
	"github.com/valivishy/httpfromtcp/internal/request"
	"github.com/valivishy/httpfromtcp/internal/response"
	"net"
	"time"
)

type Server struct {
	listener net.Listener
	handler  Handler
	open     bool
}

func Serve(port int, handler Handler) (*Server, error) {
	server := &Server{handler: handler}

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

	if err := conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond)); err != nil {
		fmt.Printf("warning: failed to set read deadline: %v\n", err)
		return
	}
	parsedRequest, err := request.FromReader(conn)
	fmt.Printf("Parsed request: %#v\n", parsedRequest)
	if err != nil {
		fmt.Printf("warning: failed to parse request: %v\n", err)
		return
	}

	buffer := bytes.Buffer{}
	handlerError := s.handler(&buffer, parsedRequest)
	if handlerError != nil {
		if err = WriteHandlerError(conn, *handlerError); err != nil {
			panic(err)
		}
		return
	}

	if err = response.WriteStatusLine(conn, response.OK); err != nil {
		panic(err)
	}

	if err = response.WriteHeaders(conn, response.GetDefaultHeaders(buffer.Len())); err != nil {
		panic(err)
	}

	_, err = conn.Write(buffer.Bytes())
	if err != nil {
		fmt.Printf("warning: failed to write to connection: %v\n", err)
		return
	}
}
