package headers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaderParse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:32020\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:32020", headers["Host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Invalid spacing header (spaces between field name and colon)
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid header with space inside the field name
	headers = NewHeaders()
	data = []byte("Ho st: localhost:32020\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Finished parsing (empty line signals end of headers)
	headers = NewHeaders()
	data = []byte("\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.True(t, done)

	// Test: Multiple valid headers
	headers = NewHeaders()
	data = []byte("Host: localhost:32020\r\nUser-Agent: Go-http-client/1.1\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.False(t, done)
	assert.Equal(t, "localhost:32020", headers["Host"])

	data = data[n:]
	n2, done, err := headers.Parse(data)
	require.NoError(t, err)
	assert.False(t, done)
	assert.Equal(t, "Go-http-client/1.1", headers["User-Agent"])

	data = data[n2:]
	n3, done, err := headers.Parse(data)
	require.NoError(t, err)
	assert.True(t, done)
	assert.Equal(t, 0, n3)

	// Test: Header with extra leading and trailing whitespace
	headers = NewHeaders()
	data = []byte("    Connection:   keep-alive    \r\n\r\n")
	n, done, err = headers.Parse(data)
	fmt.Printf("This is the error: %s", err)
	require.NoError(t, err)
	assert.Equal(t, "keep-alive", headers["Connection"])
	assert.Equal(t, 34, n)
	assert.False(t, done)

	// Test: Header with no value (should be valid)
	headers = NewHeaders()
	data = []byte("X-Custom-Header: \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "", headers["X-Custom-Header"])
	assert.Equal(t, 19, n)
	assert.False(t, done)

	// Test: Header without a colon (invalid)
	headers = NewHeaders()
	data = []byte("InvalidHeader\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Header with only whitespace (invalid)
	headers = NewHeaders()
	data = []byte("       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}
