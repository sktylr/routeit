package routeit

import (
	"bytes"
	"fmt"
	"strings"
)

type TestClient struct {
	s *Server
}

// Instantiates a test client that can be used to perform end-to-end tests on
// the server.
func NewTestClient(s *Server) TestClient {
	return TestClient{s}
}

// Makes a GET request against the specific path. Should not include a trailing
// slash but may optionally omit a leading slash.
func (tc TestClient) Get(path string) *TestResponse {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	var rb bytes.Buffer
	rb.WriteString(fmt.Sprintf("GET %s HTTP/1.1\r\n", path))
	rb.WriteString("Host: routeit\r\nUser-Agent: test-client\n\r\n")
	rb.WriteString("\r\n")
	rw := tc.s.handleNewRequest(rb.Bytes())
	return &TestResponse{rw}
}

// Makes a HEAD request against the specific path. Should not include a trailing
// slash but may optionally omit the leading slash
func (tc TestClient) Head(path string) *TestResponse {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	var rb bytes.Buffer
	rb.WriteString(fmt.Sprintf("HEAD %s HTTP/1.1\r\n", path))
	rb.WriteString("Host: routeit\r\nUser-Agent: test-client\n\r\n")
	rb.WriteString("\r\n")
	rw := tc.s.handleNewRequest(rb.Bytes())
	return &TestResponse{rw}
}
