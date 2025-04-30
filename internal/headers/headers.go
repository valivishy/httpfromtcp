package headers

import (
	"errors"
	"regexp"
	"strings"
)

const (
	crlf          = "\r\n"
	invalidHeader = "invalid header"
)

type Headers map[string]string

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	if len(data) < 1 {
		return 0, false, errors.New("no data provided")
	}

	potentialTarget := string(data)
	crlfIndex := strings.Index(potentialTarget, crlf)
	if crlfIndex == -1 {
		return 0, false, nil
	} else if crlfIndex == 0 {
		return 0, true, nil
	}

	potentialTarget = strings.TrimSpace(potentialTarget[:crlfIndex+2])

	colonIndex := strings.Index(potentialTarget, ":")
	if colonIndex == -1 {
		return 0, false, errors.New(invalidHeader)
	}

	// If the colon is the first symbol, it's invalid
	headerName := strings.TrimSpace(potentialTarget[:colonIndex])
	if headerName == "" {
		return 0, false, errors.New(invalidHeader)
	}

	// If the colon is not right after the header name, it's invalid
	if len(headerName) != colonIndex {
		return 0, false, errors.New(invalidHeader)
	}

	if err = validateHeaderName(headerName); err != nil {
		return 0, false, err
	}

	h[strings.ToLower(headerName)] = strings.TrimSpace(potentialTarget[colonIndex+1:])

	return len([]byte(string(data)[:crlfIndex+2])), false, nil
}

func validateHeaderName(name string) error {
	if name == "" {
		return errors.New(invalidHeader)
	}

	validHeaderName := regexp.MustCompile(`^[A-Za-z0-9!#$%&'*+\-.\^_` + "`" + `{|}~]+$`)
	if !validHeaderName.MatchString(name) {
		return errors.New(invalidHeader)
	}

	return nil
}
