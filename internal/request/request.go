package request

import (
	"errors"
	"fmt"
	"github.com/valivishy/httpfromtcp/internal/headers"
	"io"
	"net/http"
	"strings"
)

type state int

const (
	initialized state = iota
	done
	requestStateParsingHeaders
)

type Request struct {
	RequestLine  Line
	requestState state
	Headers      headers.Headers
}

type Line struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const bufferSize = 8
const crlf = "\r\n"

var httpMethods = []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace}

func FromReader(reader io.Reader) (*Request, error) {
	buffer := make([]byte, bufferSize, bufferSize*2)
	readBytes := 0
	request := Request{requestState: initialized, Headers: make(headers.Headers)}

	for request.requestState != done {
		if readBytes >= cap(buffer)/2 {
			temp := make([]byte, cap(buffer)*2)
			copy(temp, buffer[:readBytes])
			buffer = temp
		}

		n, err := reader.Read(buffer[readBytes : readBytes+bufferSize])
		if err != nil && err == io.EOF {
			request.requestState = done
			break
		}

		readBytes += n
		parsed, err := request.parse(buffer)
		if err != nil {
			return nil, err
		}
		if parsed == 0 {
			continue
		}

		var temp []byte
		if len(buffer) > parsed {
			temp = make([]byte, len(buffer)-parsed)
			readBytes -= parsed
		} else {
			temp = make([]byte, len(buffer))
			readBytes = parsed
		}
		copy(temp, buffer[parsed:])
		buffer = temp
	}

	if request.requestState == done && request.RequestLine == (Line{}) {
		return nil, errors.New("error: request line not found")
	}
	return &request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	if r.requestState == initialized {
		line, bytesRead, err := parseRequestLine(string(data))
		if err != nil {
			return -1, err
		}
		if bytesRead == 0 {
			return 0, nil
		}
		r.RequestLine = line
		r.requestState = requestStateParsingHeaders
		return bytesRead, nil
	}

	if r.requestState == requestStateParsingHeaders {
		n, d, err := r.Headers.Parse(data)
		if err != nil {
			return -1, err
		}

		if d {
			r.requestState = done
			return n, nil
		}

		return n, nil
	}

	if r.requestState == done {
		return -1, errors.New("error: trying to read data in a done state")
	}

	return -1, errors.New("error: unknown state")
}

func parseRequestLine(line string) (Line, int, error) {
	if !strings.Contains(line, crlf) {
		return Line{}, 0, nil
	}

	split := strings.Split(line, crlf)
	if len(split) < 1 {
		return Line{}, -1, invalidRequestLine(line)
	}

	requestLine := strings.TrimSpace(split[0])
	if requestLine == "" {
		return Line{}, -1, invalidRequestLine(line)
	}

	lineComponents := strings.Split(requestLine, " ")
	if len(lineComponents) != 3 {
		return Line{}, -1, invalidRequestLine(line)
	}

	method, err := getMethod(lineComponents[0])
	if err != nil {
		return Line{}, -1, err
	}

	target, err := getTarget(lineComponents[1])
	if err != nil {
		return Line{}, -1, err
	}

	httpVersion, err := getHttpVersion(lineComponents[2])
	if err != nil {
		return Line{}, -1, err
	}

	return Line{
			HttpVersion:   httpVersion,
			RequestTarget: target,
			Method:        method,
		},
		len([]byte(split[0])) + len(crlf),
		nil
}

func invalidRequestLine(line string) error {
	return fmt.Errorf("invalid request line: %s", line)
}

func getMethod(component string) (string, error) {
	if strings.ToUpper(component) != component {
		return "", invalidRequestLine(component)
	}

	for _, method := range httpMethods {
		if method == component {
			return method, nil
		}
	}

	return "", invalidRequestLine(component)
}

func getTarget(component string) (string, error) {
	if len(component) < 1 {
		return "", invalidRequestLine(component)
	}

	if component == "/" {
		return component, nil
	}

	paths := strings.Split(component[1:], "/")
	if len(paths) < 1 {
		return "", invalidRequestLine(component)
	}

	for _, path := range paths {
		if len(path) < 1 {
			return "", invalidRequestLine(component)
		}
	}

	return component, nil
}

func getHttpVersion(component string) (string, error) {
	if component != "HTTP/1.1" {
		return "", invalidRequestLine(component)
	}

	return strings.Split(component, "/")[1], nil
}
