package response

import (
	"fmt"
	"github.com/valivishy/httpfromtcp/internal/headers"
	"io"
	"strconv"
)

type StatusCode int

const (
	OK                  StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var message string
	switch statusCode {
	case OK:
		message = "HTTP/1.1 200 OK\r\n"
	case BadRequest:
		message = "HTTP/1.1 400 Bad Request\r\n"
	case InternalServerError:
		message = "HTTP/1.1 500 Internal Server Error\r\n"
	}

	if _, err := w.Write([]byte(message)); err != nil {
		return err
	}

	return nil
}

func GetDefaultHeaders(contentLength int) headers.Headers {

	header := headers.Headers{}
	header["Content-Type"] = "text/plain"
	header["Connection"] = "close"
	header["Content-Length"] = strconv.Itoa(contentLength)

	return header
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for name, value := range headers {
		if _, err := fmt.Fprintf(w, "%s: %s\r\n", name, value); err != nil {
			return err
		}
	}
	if _, err := w.Write([]byte("\r\n")); err != nil {
		return err
	}

	return nil
}
