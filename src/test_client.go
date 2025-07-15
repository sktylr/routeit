package routeit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type TestClient struct {
	s *Server
}

type testRequest struct {
	path    string
	headers headers
	method  HttpMethod
	body    []byte
}

// Instantiates a test client that can be used to perform end-to-end tests on
// the server.
func NewTestClient(s *Server) TestClient {
	return TestClient{s}
}

// Makes a GET request against the specific path. Should not include a trailing
// slash but may optionally omit a leading slash.
func (tc TestClient) Get(path string) *TestResponse {
	req := testRequest{
		path:    path,
		method:  GET,
		headers: headers{},
	}
	return tc.makeRequest(req)
}

// Makes a HEAD request against the specific path. Should not include a trailing
// slash but may optionally omit the leading slash
func (tc TestClient) Head(path string) *TestResponse {
	req := testRequest{
		path:    path,
		method:  HEAD,
		headers: headers{},
	}
	return tc.makeRequest(req)
}

func (tc TestClient) PostJson(path string, body any) *TestResponse {
	bodyJson, err := json.Marshal(body)
	if err != nil {
		// We panic here since this is inside a test and expected to be correct
		panic(err)
	}
	headers := headers{
		"Content-Type":   CTApplicationJson.string(),
		"Content-Length": fmt.Sprintf("%d", len(bodyJson)),
	}
	req := testRequest{
		path:    path,
		method:  POST,
		headers: headers,
		body:    bodyJson,
	}
	return tc.makeRequest(req)
}

func (tc TestClient) PostText(path string, text string) *TestResponse {
	raw := []byte(text)
	headers := headers{
		"Content-Type":   CTTextPlain.string(),
		"Content-Length": fmt.Sprintf("%d", len(raw)),
	}
	req := testRequest{
		path:    path,
		method:  POST,
		headers: headers,
		body:    raw,
	}
	return tc.makeRequest(req)
}

func (tc TestClient) makeRequest(req testRequest) *TestResponse {
	if !strings.HasPrefix(req.path, "/") {
		req.path = "/" + req.path
	}

	if _, found := req.headers["Host"]; !found {
		req.headers.set("Host", "routeit")
	}
	if _, found := req.headers["User-Agent"]; !found {
		req.headers.set("User-Agent", "test-client")
	}

	var rb bytes.Buffer
	rb.WriteString(fmt.Sprintf("%s %s HTTP/1.1\r\n", req.method.name, req.path))
	for k, v := range req.headers {
		rb.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	rb.Write(req.body)

	rw := tc.s.handleNewRequest(rb.Bytes())
	return &TestResponse{rw}
}
