package request

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Request struct {
	RequestLine Line
}

type Line struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

var httpMethods = []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace}

func FromReader(reader io.Reader) (*Request, error) {
	all, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	line, err := parseRequestLine(string(all))
	if err != nil {
		return nil, err
	}

	return &Request{line}, nil
}

func parseRequestLine(line string) (Line, error) {
	split := strings.Split(line, "\r\n")
	if len(split) < 1 {
		return Line{}, invalidRequestLine(line)
	}

	requestLine := strings.TrimSpace(split[0])
	if requestLine == "" {
		return Line{}, invalidRequestLine(line)
	}

	lineComponents := strings.Split(requestLine, " ")
	if len(lineComponents) != 3 {
		return Line{}, invalidRequestLine(line)
	}

	method, err := getMethod(lineComponents[0])
	if err != nil {
		return Line{}, err
	}

	target, err := getTarget(lineComponents[1])
	if err != nil {
		return Line{}, err
	}

	httpVersion, err := getHttpVersion(lineComponents[2])
	if err != nil {
		return Line{}, err
	}

	return Line{
		HttpVersion:   httpVersion,
		RequestTarget: target,
		Method:        method,
	}, nil
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
