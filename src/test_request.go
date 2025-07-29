package routeit

import (
	"strings"
)

type TestRequest struct {
	Body      []byte
	Headers   []string
	UserAgent string
	Ip        string
}

type testRequest struct {
	path    string
	headers headers
	method  HttpMethod
	body    []byte
}

// This will create a new request object that can be used in tests, for example
// when unit testing middleware.
func NewTestRequest(t testable, path string, m HttpMethod, tr TestRequest) *Request {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	headers := constructTestHeaders(tr.Headers...)
	r := &Request{
		ctx:       t.Context(),
		mthd:      m,
		uri:       uri{edgePath: path},
		headers:   headers,
		body:      tr.Body,
		userAgent: tr.UserAgent,
		ip:        tr.Ip,
		accept:    parseAcceptHeader(headers),
	}

	if host, hasHost := headers.Get("Host"); hasHost {
		r.host = host
	} else {
		r.host = "localhost:8080"
	}

	if ct, hasCt := headers.Get("Content-Type"); hasCt {
		r.ct = parseContentType(ct)
	}

	if r.ip == "" {
		r.ip = "127.0.0.1"
	}

	if r.userAgent == "" {
		r.userAgent = "routeit-test"
	}

	return r
}

func constructTestHeaders(h ...string) headers {
	i := 0
	total := len(h)
	headers := headers{}
	for i < total-1 {
		headers.Set(h[i], h[i+1])
		i += 2
	}
	return headers
}
