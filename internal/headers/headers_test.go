package headers

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseSingleHeader(t *testing.T) {
	headers := Headers{}
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)
}

func TestParseSingleHeaderWithExtraWhitespace(t *testing.T) {
	headers := Headers{}
	data := []byte("Host:     localhost:42069  \r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, 29, n)
	assert.False(t, done)
}

func TestParseMultipleHeadersWithExisting(t *testing.T) {
	headers := Headers{"Content-Type": "application/json"}
	source := "Host: localhost:42069\r\nAccept: */*\r\n\r\n"
	data := []byte(source)
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, 23, n)

	source = source[n:]
	data = []byte(source)
	n, done, err = headers.Parse(data)
	assert.Equal(t, "*/*", headers["Accept"])
	assert.Equal(t, 13, n)
	assert.False(t, done)

	source = source[n:]
	data = []byte(source)
	n, done, err = headers.Parse(data)
	assert.Equal(t, 0, n)
	assert.True(t, done)

	assert.Equal(t, "application/json", headers["Content-Type"])
}

func TestRequestLineParseFailing(t *testing.T) {
	// Test: Invalid spacing header
	headers := Headers{}
	data := []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

func TestRequestLineWithEqualsParsing(t *testing.T) {
	// Test: Unusual but valid header format
	headers := Headers{}
	data := []byte("Host=localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "42069", headers["Host=localhost"])
	assert.Equal(t, 22, n)
	assert.False(t, done)
}
