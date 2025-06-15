package routeit

import (
	"errors"
	"fmt"
	"testing"
)

// TODO: some of the tests in here are dependent on having a request body in a GET request. Once future request methods are supported these should be changed and a request body should be ignored for GET requests

func TestRequestFromRawInvalidProtocolLine(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantStatus HttpStatus
	}{
		{
			"path contains spaces",
			"GET /hello bad path HTTP/1.1\r\nHost: localhost\r\n\r\n",
			StatusBadRequest,
		},
		{
			"path missing",
			"GET HTTP/1.1\r\nHost: localhost\r\n\r\n",
			StatusBadRequest,
		},
		{
			"unsupported method",
			"POST / HTTP/1.1\r\nHost: localhost\r\n\r\n",
			StatusNotImplemented,
		},
		{
			"unsupported http version",
			"GET / HTTP/2.0\r\nHost: localhost\r\n\r\n",
			StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bts := []byte(tc.input)
			_, err := requestFromRaw(bts)
			verifyHttpError(t, err, tc.wantStatus)
		})
	}
}

func TestRequestFromRawNoCarriageReturn(t *testing.T) {
	bts := []byte("GET / HTTP/1.1\nHost: localhost\n\nbody")
	_, err := requestFromRaw(bts)
	verifyHttpError(t, err, StatusBadRequest)
}

func TestRequestFromRawNoHeaders(t *testing.T) {
	in := []byte("GET / HTTP/1.1\r\n\r\nthe body\r\n")

	_, err := requestFromRaw(in)
	verifyHttpError(t, err, StatusBadRequest)
}

func TestRequestFromRawOneHeader(t *testing.T) {
	in := []byte("GET / HTTP/1.1\r\nHost: localhost\r\n\r\nthe body")

	req, err := requestFromRaw(in)
	if err != nil {
		t.Errorf("requestFromRaw one header unexpected error %s", err)
	}
	// Unparsed since there is no Content-Type header
	expectBody(t, "one header", req.body, "")
	expectUrl(t, "one header", req.url, "/")
	if len(req.headers) != 1 {
		t.Errorf(`requestFromRaw one header headers = %q, wanted {"Host": "localhost"}`, req.headers)
	}
	expectHeader(t, "one header", "Host", req.headers, "localhost")
	expectMethod(t, "one header", req.mthd, GET)
}

func TestRequestFromRawMultipleHeaders(t *testing.T) {
	in := []byte("GET / HTTP/1.1\r\nContent-Length: 8\r\nContent-Type: text/plain\r\nHost: localhost\r\n\r\nthe body")
	wantCl := "8"
	wantCt := "text/plain"
	wantHost := "localhost"
	wantHdrs := map[string]string{
		"Content-Length": wantCl,
		"Content-Type":   wantCt,
		"Host":           wantHost,
	}

	req, err := requestFromRaw(in)
	if err != nil {
		t.Errorf("requestFromRaw multiple headers unexpected error %s", err)
	}
	expectBody(t, "multiple headers", req.body, "the body")
	expectUrl(t, "multiple headers", req.url, "/")
	if len(req.headers) != len(wantHdrs) {
		t.Errorf(`requestFromRaw multiple headers headers = %q, wanted %#q`, req.headers, wantHdrs)
	}
	expectHeader(t, "multiple headers", "Content-Type", req.headers, wantCt)
	expectHeader(t, "multiple headers", "Content-Length", req.headers, wantCl)
	expectHeader(t, "multiple headers", "Host", req.headers, wantHost)
	expectMethod(t, "multiple headers", req.mthd, GET)
}

func TestRequestFromRawOnlyConsumesContentLength(t *testing.T) {
	in := []byte("GET / HTTP/1.1\r\nHost: localhost\r\nContent-Length: 3\r\n\r\nthis is a long body!")
	want := "thi"

	req, err := requestFromRaw(in)
	if err != nil {
		t.Errorf("requestFromRaw only consumes content length unexpected error %s", err)
	}
	expectBody(t, "only consumes content length", req.body, want)
}

func TestRequestFromRawRequiresHeaders(t *testing.T) {
	in := []byte("GET / HTTP/1.1\r\n\r\n")

	req, err := requestFromRaw(in)
	if req != nil {
		t.Error("requestFromRaw requires headers expected nil request")
	}
	verifyHttpError(t, err, StatusBadRequest)
}

func TestRequestFromRawRequiresHostHeader(t *testing.T) {
	in := []byte("GET / HTTP/1.1\r\nAccept: */*\r\n\r\n")

	req, err := requestFromRaw(in)
	if req != nil {
		t.Error("requestFromRaw requires Host headers expected nil request")
	}
	verifyHttpError(t, err, StatusBadRequest)
}

func TestRequestFromRawParsesQueryString(t *testing.T) {
	in := []byte("GET /endpoint?q1=hello&q2=nice HTTP/1.1\r\nHost: localhost\r\n\r\n")
	wantq1 := "hello"
	wantq2 := "nice"
	wantQuery := map[string]string{
		"q1": wantq1,
		"q2": wantq2,
	}

	req, err := requestFromRaw(in)
	if err != nil {
		t.Errorf("requestFromRaw parses query string unexpected error %s", err)
	}
	expectBody(t, "parses query string", req.body, "")
	expectUrl(t, "parses query string", req.url, "/endpoint")
	expectMethod(t, "parses query string", req.mthd, GET)
	if len(req.queries) != len(wantQuery) {
		t.Errorf(`requestFromRaw parses query string query = %q, wanted %#q`, req.queries, wantQuery)
	}
	q1, exists := req.queries["q1"]
	if !exists {
		t.Errorf(`requestFromRaw parses query string expected header "q1" to exist`)
	}
	if q1 != wantq1 {
		t.Errorf(`requestFromRaw parses query string headers["q1"] = %q, wanted %#q`, q1, wantq1)
	}
	q2, exists := req.queries["q2"]
	if !exists {
		t.Errorf(`requestFromRaw parses query string expected header "q2" to exist`)
	}
	if q2 != wantq2 {
		t.Errorf(`requestFromRaw parses query string headers["q2"] = %q, wanted %#q`, q2, wantq2)
	}
}

func TestRequestFromRawRejectsMalformedQueryStrings(t *testing.T) {
	in := []byte("GET /endpoint?q=foo?bar HTTP/1.1\r\nHost: localhost\r\n\r\n")

	_, err := requestFromRaw(in)
	verifyHttpError(t, err, StatusBadRequest)
}

func TestRequestFromRawAllowsComplexBodies(t *testing.T) {
	body := `{
  "key": "value",
  "list": [
    {
      "key1": "value",
      "key2": {
        "name": "John"
      }
    },
    {
      "key1": "value2"
    }
  ]
}`
	bodyBytes := []byte(body)
	in := []byte(fmt.Sprintf("GET / HTTP/1.1\r\nHost: localhost\r\nContent-Length: %d\r\n\r\n%s", len(bodyBytes), body))

	req, err := requestFromRaw(in)
	if err != nil {
		t.Errorf("requestFromRaw complex bodies unexpected error %s", err)
	}
	expectBody(t, "complex bodies", req.body, body)
	expectHeader(t, "complex bodies", "Host", req.headers, "localhost")
	expectHeader(t, "complex bodies", "Content-Length", req.headers, fmt.Sprintf("%d", len(bodyBytes)))
	expectMethod(t, "complex bodies", req.mthd, GET)
	expectUrl(t, "complex bodies", req.url, "/")
}

func expectBody(t *testing.T, msg string, got string, want string) {
	if got != want {
		t.Errorf(`requestFromRaw %s body = %q, wanted %#q`, msg, got, want)
	}
}

func expectUrl(t *testing.T, msg string, got string, want string) {
	if got != want {
		t.Errorf(`requestFromRaw %s url = %q, wanted %#q`, msg, got, want)
	}
}

func expectHeader(t *testing.T, msg string, key string, hdrs headers, want string) {
	got, exists := hdrs[key]
	if !exists {
		t.Errorf(`requestFromRaw %s expected header %#q to exist`, msg, key)
	}
	if got != want {
		t.Errorf(`requestFromRaw %s headers[%q] = %q, wanted %#q`, msg, key, got, want)
	}
}

func expectMethod(t *testing.T, msg string, got HttpMethod, want HttpMethod) {
	if got != want {
		t.Errorf(`requestFromRaw %s mthd = %q, wanted %#q`, msg, got, want)
	}
}

func verifyHttpError(t *testing.T, err error, want HttpStatus) {
	if err == nil {
		t.Error("expected error to be present")
	}
	var httpErr *httpError
	if !errors.As(err, &httpErr) {
		t.Errorf("expected httpError, got %T", err)
	}
	if httpErr.Status != want {
		t.Errorf("httpError status got [status=%d, msg=%s], wanted [status=%d, msg=%s]", httpErr.Status.code, httpErr.Status.msg, want.code, want.msg)
	}
}
