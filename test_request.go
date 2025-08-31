package routeit

import (
	"fmt"
	"strings"
	"testing"
)

// The [TestRequestOptions] allow you to specify certain traits the request
// will have under test. You do not need to use any of the fields, though it
// can be helpful to set certain state on the request, such as a specific
// header, if the unit under test relies on the state to perform an action.
type TestRequestOptions struct {
	// The raw body of the request. This must be marshalled by the user of
	// [TestRequestOptions] and the corresponding Content-Type header should be
	// included in [TestRequestOptions.Headers] if the handler or middleware
	// uses safe body loads. Additionally, the Content-Length header should
	// also be provided.
	Body []byte
	// The headers of the request, if any. These should be specific as key,
	// value pairs in order. For example:
	// 	[]string{"Authorization", "Bearer foo", "Content-Type": "text/plain"}
	Headers []string
	// The Ip address of the client. This may be useful for security middleware
	Ip string
	// Specific path parameters that the URI should contain and extract.
	// Currently this needs to be explicitly defined since unit tests do not go
	// through the entire handling flow, meaning requests are not routed
	// properly so path parameters are not extracted.
	PathParams map[string]string
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
func NewTestRequest(t testing.TB, path string, m HttpMethod, opts TestRequestOptions) *TestRequest {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	uri, err := parseUri(path)
	if err != nil {
		panic(fmt.Errorf(`failed to construct uri from path %#q: %+v`, path, err))
	}

	if opts.PathParams != nil {
		uri.pathParams = opts.PathParams
	}

	headers := &RequestHeaders{constructTestHeaders(opts.Headers...)}
	req := &Request{
		ctx:     t.Context(),
		mthd:    m,
		uri:     *uri,
		headers: headers,
		body:    opts.Body,
		ip:      opts.Ip,
		accept:  parseAcceptHeader(headers),
	}

	if host, hasHost := headers.First("Host"); hasHost {
		req.host = host
	} else {
		req.host = "localhost:8080"
	}

	if ct, hasCt := headers.First("Content-Type"); hasCt {
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
		headers.Append(h[i], h[i+1])
		i += 2
	}
	return headers
}
