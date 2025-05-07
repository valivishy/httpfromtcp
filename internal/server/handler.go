package server

import (
	"github.com/valivishy/httpfromtcp/internal/request"
	"github.com/valivishy/httpfromtcp/internal/response"
	"io"
)

type HandlerError struct {
	StatusCode int
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func WriteHandlerError(w io.Writer, handlerError HandlerError) error {
	var statusCode response.StatusCode
	switch handlerError.StatusCode {
	case 400:
		statusCode = response.BadRequest
	case 500:
		statusCode = response.InternalServerError
	default:
		statusCode = response.BadRequest
	}

	if err := response.WriteStatusLine(w, statusCode); err != nil {
		return err
	}

	headers := response.GetDefaultHeaders(len(handlerError.Message))
	if err := response.WriteHeaders(w, headers); err != nil {
		return err
	}

	if _, err := w.Write([]byte(handlerError.Message)); err != nil {
		return err
	}

	return nil
}

func HandlerFunc(w io.Writer, req *request.Request) *HandlerError {
	if req.RequestLine.RequestTarget == "/yourproblem" {
		return &HandlerError{
			StatusCode: 400,
			Message:    "Your problem is not my problem",
		}
	}

	if req.RequestLine.RequestTarget == "/myproblem" {
		return &HandlerError{
			StatusCode: 500,
			Message:    "Woopsie, my bad\n",
		}
	}

	if _, err := w.Write([]byte("All good, frfr\n")); err != nil {
		panic(err)
	}

	return nil
}
