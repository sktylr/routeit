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
// slash but may optionally omit a leading slash. Can include an arbitrary
// number of headers, specified after the path. Keys and values of headers
// should be individual arguments.
func (tc TestClient) Get(path string, h ...string) *TestResponse {
	req := testRequest{
		path:    path,
		method:  GET,
		headers: tc.constructHeaders(h...),
	}
	return tc.makeRequest(req)
}

// Makes a HEAD request against the specific path. Should not include a trailing
// slash but may optionally omit the leading slash. Can include an arbitrary
// number of headers, specified after the path. Keys and values of headers
// should be individual arguments.
func (tc TestClient) Head(path string, h ...string) *TestResponse {
	req := testRequest{
		path:    path,
		method:  HEAD,
		headers: tc.constructHeaders(h...),
	}
	return tc.makeRequest(req)
}

// Makes a POST request against the specified path, using the second argument
// as the request body, which is converted to Json. Panics if the Json
// conversion fails. Can include an arbitrary number of headers, specified
// after the request body. Keys and values of headers should be individual
// arguments.
func (tc TestClient) PostJson(path string, body any, h ...string) *TestResponse {
	bodyJson, err := json.Marshal(body)
	if err != nil {
		// We panic here since this is inside a test and expected to be correct
		panic(err)
	}
	headers := tc.constructHeaders(h...)
	headers.Set("Content-Type", CTApplicationJson.string())
	headers.Set("Content-Length", fmt.Sprintf("%d", len(bodyJson)))
	req := testRequest{
		path:    path,
		method:  POST,
		headers: headers,
		body:    bodyJson,
	}
	return tc.makeRequest(req)
}

// Makes a POST request against the specified path, using a text request body.
// Can include an arbitrary number of headers, specific after the request body.
// Keys and values of headers should be individual arguments.
func (tc TestClient) PostText(path string, text string, h ...string) *TestResponse {
	raw := []byte(text)
	headers := tc.constructHeaders(h...)
	headers.Set("Content-Type", CTTextPlain.string())
	headers.Set("Content-Length", fmt.Sprintf("%d", len(raw)))
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
		req.headers.Set("Host", "routeit")
	}
	if _, found := req.headers["User-Agent"]; !found {
		req.headers.Set("User-Agent", "test-client")
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

func (tc TestClient) constructHeaders(h ...string) headers {
	i := 0
	total := len(h)
	headers := headers{}
	for i < total-1 {
		headers.Set(h[i], h[i+1])
		i++
	}
	return headers
}
