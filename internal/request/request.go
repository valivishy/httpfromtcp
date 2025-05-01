package request

import (
	"errors"
	"fmt"
	"github.com/valivishy/httpfromtcp/internal/headers"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type state int

const (
	initialized state = iota
	done
	requestStateParsingHeaders
	requestStateParsingBody
)

type Request struct {
	RequestLine  Line
	Headers      headers.Headers
	Body         []byte
	requestState state
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
		buffer = resizeBuffer(readBytes, buffer)

		var n int
		var err error
		if request.requestState == requestStateParsingBody {
			buffer, readBytes, err = readBody(reader, buffer, readBytes, err)
			if err != nil {
				return nil, err
			}
		} else {
			n, err = reader.Read(buffer[readBytes : readBytes+bufferSize])
			if err != nil && err == io.EOF {
				request.requestState = done
				break
			}
			readBytes += n
		}

		parsed, err := request.parse(buffer)
		if err != nil {
			return nil, err
		}
		if parsed == 0 {
			continue
		}

		readBytes, buffer = rebuildBuffer(buffer, parsed, readBytes)
	}

	if request.requestState == done && request.RequestLine == (Line{}) {
		return nil, errors.New("error: request line not found")
	}
	return &request, nil
}

func readBody(reader io.Reader, buffer []byte, readBytes int, err error) ([]byte, int, error) {
	for {
		buffer = resizeBuffer(readBytes, buffer)

		n, err := reader.Read(buffer[readBytes : readBytes+bufferSize])
		if err != nil && err == io.EOF {
			break
		}
		readBytes += n
	}
	buffer = buffer[:readBytes]

	return buffer, readBytes, err
}

func rebuildBuffer(buffer []byte, parsed int, readBytes int) (int, []byte) {
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

	return readBytes, buffer
}

func resizeBuffer(readBytes int, buffer []byte) []byte {
	if readBytes >= cap(buffer)/2 {
		temp := make([]byte, (cap(buffer)+bufferSize)*2)
		copy(temp, buffer[:readBytes])
		buffer = temp
	}
	return buffer
}

func (r *Request) parse(data []byte) (int, error) {
	if r.requestState == initialized {
		return r.parseLine(data)
	}

	if r.requestState == requestStateParsingHeaders {
		return r.parseHeaders(data)
	}

	if r.requestState == requestStateParsingBody {
		return r.parseBody(data)
	}

	if r.requestState == done {
		return -1, errors.New("error: trying to read data in a done state")
	}

	return -1, errors.New("error: unknown state")
}

func (r *Request) parseBody(data []byte) (int, error) {
	contentLengthString, ok := r.Headers.Get("Content-Length")
	if !ok {
		r.requestState = done
		return 0, nil
	}

	contentLength, err := strconv.Atoi(contentLengthString)
	if err != nil {
		return -1, err
	}

	// Remove the CRLF
	data = data[2:]

	temp := make([]byte, len(r.Body)+len(data))
	copy(temp, r.Body)
	copy(temp[len(r.Body):], data)
	r.Body = temp

	if len(r.Body) != contentLength {
		return -1, errors.New("error: body length exceeds content length")
	}

	r.requestState = done

	return len(data), nil
}

func (r *Request) parseHeaders(data []byte) (int, error) {
	n, d, err := r.Headers.Parse(data)
	if err != nil {
		return -1, err
	}

	if d {
		r.requestState = requestStateParsingBody
		return n, nil
	}

	return n, nil
}

func (r *Request) parseLine(data []byte) (int, error) {
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
