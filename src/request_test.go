package routeit

import (
	"fmt"
	"testing"
)

func TestRequestFromRaw(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
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
				"PATCH / HTTP/1.1\r\nHost: localhost\r\n\r\n",
				StatusNotImplemented,
			},
			{
				"unsupported http version",
				"GET / HTTP/2.0\r\nHost: localhost\r\n\r\n",
				StatusHttpVersionNotSupported,
			},
			{
				"missing method",
				"/ HTTP/1.1\r\nHost: localhost\r\n\r\n",
				StatusBadRequest,
			},
			{
				"missing protocol",
				"GET /\r\nHost: localhost\r\n\r\n",
				StatusBadRequest,
			},
			{
				"no carriage return",
				"POST / HTTP/1.1\nHost: localhost\n\nbody",
				StatusBadRequest,
			},
			{
				"no headers and body",
				"POST / HTTP/1.1\r\n\r\nthe body\r\n",
				StatusBadRequest,
			},
			{
				"no headers or body",
				"GET / HTTP/1.1\r\n\r\n",
				StatusBadRequest,
			},
			{
				"without Host header",
				"GET / HTTP/1.1\r\nAccept: */*\r\n\r\n",
				StatusBadRequest,
			},
			{
				"malformed query string",
				"GET /endpoint?q=foo?bar HTTP/1.1\r\nHost: localhost\r\n\r\n",
				StatusBadRequest,
			},
			{
				"* path and not OPTIONS",
				"GET * HTTP/1.1\r\nHost: localhost\r\n\r\n",
				StatusBadRequest,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				bts := []byte(tc.input)
				_, err := requestFromRaw(bts)
				if err == nil {
					t.Fatal("expected error to be present")
				}
				if err.status != tc.wantStatus {
					t.Errorf("httpError status got [status=%d, msg=%s], wanted [status=%d, msg=%s]", err.status.code, err.status.msg, tc.wantStatus.code, tc.wantStatus.msg)
				}
			})
		}
	})

	t.Run("allows OPTIONS * requests", func(t *testing.T) {
		bts := []byte("OPTIONS * HTTP/1.1\r\nHost: localhost\r\n\r\n")

		req, err := requestFromRaw(bts)
		if err != nil {
			t.Errorf(`Error() = %v, expected nil`, err)
		}
		if req.Path() != "*" {
			t.Errorf(`Path() = %#q, wanted "*"`, req.Path())
		}
		if req.Method() != OPTIONS {
			t.Errorf(`Method() = %#q, wanted OPTIONS`, req.Method().name)
		}
	})

	t.Run("prefixes leading slash", func(t *testing.T) {
		// TODO: do we need this? is it spec compliant to do this?
		bts := []byte("GET hello HTTP/1.1\r\nHost: localhost\r\n\r\n")

		req, err := requestFromRaw(bts)
		if err != nil {
			t.Errorf("did not expect error = %#v", err)
		}
		expectUrl(t, req, "/hello")
	})

	t.Run("headers", func(t *testing.T) {
		t.Run("one header", func(t *testing.T) {
			in := []byte("POST / HTTP/1.1\r\nHost: localhost\r\n\r\nthe body")

			req, err := requestFromRaw(in)
			if err != nil {
				t.Errorf("requestFromRaw one header unexpected error %s", err)
			}
			// Unparsed since there is no Content-Length header
			expectBody(t, req.body, "")
			expectUrl(t, req, "/")
			if len(req.headers) != 1 {
				t.Errorf(`requestFromRaw one header headers = %q, wanted {"Host": "localhost"}`, req.headers)
			}
			expectHeader(t, "Host", req.headers, "localhost")
			expectMethod(t, req.mthd, POST)
		})

		t.Run("multiple headers", func(t *testing.T) {
			in := []byte("POST / HTTP/1.1\r\nContent-Length: 8\r\nContent-Type: text/plain\r\nHost: localhost\r\n\r\nthe body")
			wantCl := "8"
			wantCtRaw := "text/plain"
			wantHost := "localhost"
			wantHdrs := map[string]string{
				"Content-Length": wantCl,
				"Content-Type":   wantCtRaw,
				"Host":           wantHost,
			}

			req, err := requestFromRaw(in)
			if err != nil {
				t.Errorf("requestFromRaw multiple headers unexpected error %s", err)
			}
			expectBody(t, req.body, "the body")
			expectUrl(t, req, "/")
			if len(req.headers) != len(wantHdrs) {
				t.Errorf(`requestFromRaw multiple headers headers = %q, wanted %#q`, req.headers, wantHdrs)
			}
			expectHeader(t, "Content-Type", req.headers, wantCtRaw)
			expectHeader(t, "Content-Length", req.headers, wantCl)
			expectHeader(t, "Host", req.headers, wantHost)
			expectMethod(t, req.mthd, POST)
			if req.ContentType() != CTTextPlain {
				t.Errorf(`req.ContentType() = %#q, wanted text/plain`, req.ContentType().string())
			}
		})
	})

	t.Run("body", func(t *testing.T) {
		t.Run("only consumes content length", func(t *testing.T) {
			in := []byte("POST / HTTP/1.1\r\nHost: localhost\r\nContent-Length: 3\r\n\r\nthis is a long body!")
			want := "thi"

			req, err := requestFromRaw(in)
			if err != nil {
				t.Errorf("requestFromRaw only consumes content length unexpected error %s", err)
			}
			expectBody(t, req.body, want)
		})

		t.Run("does not consume body unless required by method", func(t *testing.T) {
			methods := []string{"GET", "HEAD", "OPTIONS"}

			for _, m := range methods {
				t.Run(m, func(t *testing.T) {
					in := fmt.Appendf(nil, "%s / HTTP/1.1\r\nHost: localhost\r\nContent-Length: 4\r\n\r\nbody", m)

					req, err := requestFromRaw(in)
					if err != nil {
						t.Errorf(`requestFromRaw does not consume body unexpected error %v`, err)
					}
					expectBody(t, req.body, "")
				})
			}
		})

		t.Run("allows complex bodies", func(t *testing.T) {
			body := `
{
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
			in := fmt.Appendf(nil, "POST / HTTP/1.1\r\nHost: localhost\r\nContent-Length: %d\r\n\r\n%s", len(bodyBytes), body)

			req, err := requestFromRaw(in)
			if err != nil {
				t.Errorf("requestFromRaw complex bodies unexpected error %s", err)
			}
			expectBody(t, req.body, body)
			expectHeader(t, "Host", req.headers, "localhost")
			expectHeader(t, "Content-Length", req.headers, fmt.Sprintf("%d", len(bodyBytes)))
			expectMethod(t, req.mthd, POST)
			expectUrl(t, req, "/")
		})
	})

	t.Run("query string", func(t *testing.T) {
		t.Run("parses valid", func(t *testing.T) {
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
			expectBody(t, req.body, "")
			expectUrl(t, req, "/endpoint")
			expectMethod(t, req.mthd, GET)
			if len(req.uri.queryParams) != len(wantQuery) {
				t.Errorf(`requestFromRaw parses query string query = %q, wanted %#q`, req.uri.queryParams, wantQuery)
			}
			q1, exists := req.QueryParam("q1")
			if !exists {
				t.Errorf(`requestFromRaw parses query string expected header "q1" to exist`)
			}
			if q1 != wantq1 {
				t.Errorf(`requestFromRaw parses query string headers["q1"] = %q, wanted %#q`, q1, wantq1)
			}
			q2, exists := req.QueryParam("q2")
			if !exists {
				t.Errorf(`requestFromRaw parses query string expected header "q2" to exist`)
			}
			if q2 != wantq2 {
				t.Errorf(`requestFromRaw parses query string headers["q2"] = %q, wanted %#q`, q2, wantq2)
			}
		})
	})
}

func expectBody(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf(`requestFromRaw body = %q, wanted %#q`, got, want)
	}
}

func expectUrl(t *testing.T, got *Request, want string) {
	t.Helper()
	if got.Path() != want {
		t.Errorf(`requestFromRaw Path() = %q, wanted %#q`, got.Path(), want)
	}
}

func expectHeader(t *testing.T, key string, hdrs headers, want string) {
	t.Helper()
	got, exists := hdrs.Get(key)
	if !exists {
		t.Errorf(`requestFromRaw expected header %#q to exist`, key)
	}
	if got != want {
		t.Errorf(`requestFromRaw headers[%q] = %q, wanted %#q`, key, got, want)
	}
}

func expectMethod(t *testing.T, got, want HttpMethod) {
	t.Helper()
	if got != want {
		t.Errorf(`requestFromRaw mthd = %q, wanted %#q`, got, want)
	}
}
