package routeit

import (
	"strings"
)

// The [TestRequestOptions] allow you to specify certain traits the request
// will have under test. You do not need to use any of the fields, though it
// can be helpful to set certain state on the request, such as a specific
// header, if the unit under test relies on the state to perform an action.
type TestRequestOptions struct {
	Body    []byte
	Headers []string
	Ip      string
}

// The [TestRequest] object can be used when unit testing specific components
// of a routeit application, such as middleware. It should be constructed
// using [NewTestRequest] and provides ways to verify behaviour that happens to
// a request, for example if a context value was added properly.
type TestRequest struct {
	req *Request
}

type testRequest struct {
	path    string
	headers headers
	method  HttpMethod
	body    []byte
}

// This will create a new test request object that can be used in tests, for
// example when unit testing middleware.
func NewTestRequest(t testable, path string, m HttpMethod, opts TestRequestOptions) *TestRequest {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	headers := constructTestHeaders(opts.Headers...)
	req := &Request{
		ctx:     t.Context(),
		mthd:    m,
		uri:     uri{edgePath: path},
		headers: headers,
		body:    opts.Body,
		ip:      opts.Ip,
		accept:  parseAcceptHeader(headers),
	}

	if host, hasHost := headers.Get("Host"); hasHost {
		req.host = host
	} else {
		req.host = "localhost:8080"
	}

	if ct, hasCt := headers.Get("Content-Type"); hasCt {
		req.ct = parseContentType(ct)
	}

	if req.ip == "" {
		req.ip = "127.0.0.1"
	}

	if req.userAgent == "" {
		req.userAgent = "routeit-test"
	}

	return &TestRequest{req: req}
}

func (tr *TestRequest) NewContextValue(key string, val any) {
	tr.req.NewContextValue(key, val)
}

func (tr *TestRequest) ContextValue(key string) (any, bool) {
	return tr.req.ContextValue(key)
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
