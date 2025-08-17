package routeit

import (
	"fmt"
	"reflect"
	"slices"
	"testing"
)

var defaultRequestSize = 10 * KiB

func TestRequestFromRaw(t *testing.T) {
	expectBody := func(t *testing.T, got []byte, want string) {
		t.Helper()
		if string(got) != want {
			t.Errorf(`body = %q, wanted %#q`, got, want)
		}
	}
	expectUrl := func(t *testing.T, got *Request, want string) {
		t.Helper()
		if got.Path() != want {
			t.Errorf(`Path() = %q, wanted %#q`, got.Path(), want)
		}
	}
	expectHeader := func(t *testing.T, key string, hdrs *RequestHeaders, want string) {
		t.Helper()
		got, exists := hdrs.All(key)
		if !exists {
			t.Errorf(`expected header %#q to exist`, key)
		}
		if !slices.Contains(got, want) {
			t.Errorf(`headers[%q] = %+v, wanted %#q`, key, got, want)
		}
	}
	expectMethod := func(t *testing.T, got, want HttpMethod) {
		t.Helper()
		if got != want {
			t.Errorf(`mthd = %q, wanted %#q`, got, want)
		}
	}

	t.Run("invalid", func(t *testing.T) {
		tests := []struct {
			name       string
			maxSize    RequestSize
			input      string
			wantStatus HttpStatus
		}{
			{
				name:       "path contains spaces",
				input:      "GET /hello bad path HTTP/1.1\r\nHost: localhost\r\n\r\n",
				wantStatus: StatusBadRequest,
			},
			{
				name:       "path missing",
				input:      "GET HTTP/1.1\r\nHost: localhost\r\n\r\n",
				wantStatus: StatusBadRequest,
			},
			{
				name:       "unsupported method",
				input:      "FOO / HTTP/1.1\r\nHost: localhost\r\n\r\n",
				wantStatus: StatusNotImplemented,
			},
			{
				name:       "unsupported http version",
				input:      "GET / HTTP/2.0\r\nHost: localhost\r\n\r\n",
				wantStatus: StatusHttpVersionNotSupported,
			},
			{
				name:       "missing method",
				input:      "/ HTTP/1.1\r\nHost: localhost\r\n\r\n",
				wantStatus: StatusBadRequest,
			},
			{
				name:       "missing protocol",
				input:      "GET /\r\nHost: localhost\r\n\r\n",
				wantStatus: StatusBadRequest,
			},
			{
				name:       "no carriage return",
				input:      "POST / HTTP/1.1\nHost: localhost\n\nbody",
				wantStatus: StatusBadRequest,
			},
			{
				name:       "malformed query string",
				input:      "GET /endpoint?q=foo?bar HTTP/1.1\r\nHost: localhost\r\n\r\n",
				wantStatus: StatusBadRequest,
			},
			{
				name:       "* path and not OPTIONS",
				input:      "GET * HTTP/1.1\r\nHost: localhost\r\n\r\n",
				wantStatus: StatusBadRequest,
			},
			{
				name:       "incomplete escaping",
				input:      "GET /foo% HTTP/1.1\r\nHost: localhost\r\n\r\n",
				wantStatus: StatusBadRequest,
			},
			{
				name:       "invalid escaping",
				input:      "GET /foo%ZZbar HTTP/1.1\r\nHost: localhost\r\n\r\n",
				wantStatus: StatusBadRequest,
			},
			{
				name:       "mixed escaping",
				input:      "GET /foo%\"bar HTTP/1.1\r\nHost: localhost\r\n\r\n",
				wantStatus: StatusBadRequest,
			},
			{
				name:       "body but no content type",
				input:      "POST / HTTP/1.1\r\nContent-Length: 6\r\n\r\nHello!",
				wantStatus: StatusBadRequest,
			},
			{
				name:       "unparsable Content-Type but Content-Length",
				input:      "POST / HTTP/1.1\r\nContent-Length: 6\r\nContent-Type: text-plain\r\n\r\nHello!",
				wantStatus: StatusBadRequest,
			},
			{
				name:       "max size exceeded by content-length",
				input:      "POST / HTTP/1.1\r\nContent-Length: 100\r\nContent-Type: text/plain\r\n\r\nunread",
				maxSize:    99 * Byte,
				wantStatus: StatusContentTooLarge,
			},
			{
				name:       "no blank line between body",
				input:      "POST / HTTP/1.1\r\nContent-Length: 6\r\nContent-Type: text/plain\r\nHello!",
				wantStatus: StatusBadRequest,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				bts := []byte(tc.input)
				max := tc.maxSize
				if max == 0 {
					max = defaultRequestSize
				}

				_, err := requestFromRaw(bts, max, t.Context())

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

		req, err := requestFromRaw(bts, defaultRequestSize, t.Context())
		if err != nil {
			t.Errorf(`Error() = %v, expected nil`, err)
		}
		expectUrl(t, req, "*")
		expectMethod(t, req.mthd, OPTIONS)
	})

	t.Run("TRACE", func(t *testing.T) {
		t.Run("stores entire request on body", func(t *testing.T) {
			bts := []byte("TRACE /hello HTTP/1.1\r\nHost: localhost:8080\r\n\r\n")

			req, err := requestFromRaw(bts, defaultRequestSize, t.Context())
			if err != nil {
				t.Errorf(`Error() = %v, expected nil`, err)
			}
			expectUrl(t, req, "/hello")
			expectMethod(t, req.mthd, TRACE)
			expectBody(t, req.body, string(bts))
		})

		t.Run("stores entire request on body even if request has body", func(t *testing.T) {
			bts := []byte("TRACE /hello HTTP/1.1\r\nHost: localhost:8080\r\nContent-Length: 4\r\nContent-Type: text/plain\r\n\r\nbody")

			req, err := requestFromRaw(bts, defaultRequestSize, t.Context())
			if err != nil {
				t.Errorf(`Error() = %v, expected nil`, err)
			}
			expectUrl(t, req, "/hello")
			expectMethod(t, req.mthd, TRACE)
			expectBody(t, req.body, string(bts))
		})
	})

	t.Run("prefixes leading slash", func(t *testing.T) {
		bts := []byte("GET hello HTTP/1.1\r\nHost: localhost\r\n\r\n")

		req, err := requestFromRaw(bts, defaultRequestSize, t.Context())
		if err != nil {
			t.Errorf("did not expect error = %#v", err)
		}
		expectUrl(t, req, "/hello")
	})

	t.Run("headers", func(t *testing.T) {
		t.Run("one header", func(t *testing.T) {
			in := []byte("POST / HTTP/1.1\r\nHost: localhost\r\n\r\nthe body")

			req, err := requestFromRaw(in, defaultRequestSize, t.Context())
			if err != nil {
				t.Errorf("requestFromRaw one header unexpected error %s", err)
			}
			// Unparsed since there is no Content-Length header
			expectBody(t, req.body, "")
			expectUrl(t, req, "/")
			if len(req.headers.headers) != 1 {
				t.Errorf(`requestFromRaw one header headers = %q, wanted {"Host": "localhost"}`, req.headers.headers)
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

			req, err := requestFromRaw(in, defaultRequestSize, t.Context())
			if err != nil {
				t.Errorf("requestFromRaw multiple headers unexpected error %s", err)
			}
			expectBody(t, req.body, "the body")
			expectUrl(t, req, "/")
			if len(req.headers.headers) != len(wantHdrs) {
				t.Errorf(`requestFromRaw multiple headers headers = %q, wanted %#q`, req.headers.headers, wantHdrs)
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
			in := []byte("POST / HTTP/1.1\r\nHost: localhost\r\nContent-Length: 3\r\nContent-Type: text/plain\r\n\r\nthis is a long body!")
			want := "thi"

			req, err := requestFromRaw(in, defaultRequestSize, t.Context())
			if err != nil {
				t.Errorf("requestFromRaw only consumes content length unexpected error %s", err)
			}
			expectBody(t, req.body, want)
		})

		t.Run("does not consume body unless required by method", func(t *testing.T) {
			methods := []string{"GET", "HEAD", "OPTIONS"}

			for _, m := range methods {
				t.Run(m, func(t *testing.T) {
					in := fmt.Appendf(nil, "%s / HTTP/1.1\r\nHost: localhost\r\nContent-Length: 4\r\nContent-Type: text/plain\r\n\r\nbody", m)

					req, err := requestFromRaw(in, defaultRequestSize, t.Context())
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
			in := fmt.Appendf(nil, "POST / HTTP/1.1\r\nHost: localhost\r\nContent-Length: %d\r\nContent-Type: application/json\r\n\r\n%s", len(bodyBytes), body)

			req, err := requestFromRaw(in, defaultRequestSize, t.Context())
			if err != nil {
				t.Errorf("requestFromRaw complex bodies unexpected error %s", err)
			}
			expectBody(t, req.body, body)
			expectHeader(t, "Host", req.headers, "localhost")
			expectHeader(t, "Content-Length", req.headers, fmt.Sprintf("%d", len(bodyBytes)))
			expectMethod(t, req.mthd, POST)
			expectUrl(t, req, "/")
		})

		t.Run("body contains CRLF characters", func(t *testing.T) {
			body := "Hello\r\nThis is\r\nA broken up\r\nbody, which is legal"
			bodyBytes := []byte(body)
			in := fmt.Appendf(nil, "POST / HTTP/1.1\r\nContent-Length: %d\r\nContent-Type: text/plain\r\n\r\n%s", len(bodyBytes), body)

			req, err := requestFromRaw(in, defaultRequestSize, t.Context())
			if err != nil {
				t.Errorf("unexpected error %s", err)
			}
			expectBody(t, req.body, body)
			expectHeader(t, "Content-Length", req.headers, fmt.Sprintf("%d", len(bodyBytes)))
			expectMethod(t, req.mthd, POST)
			expectUrl(t, req, "/")
		})
	})

	t.Run("parses valid query string", func(t *testing.T) {
		tests := []struct {
			in   string
			want map[string][]string
		}{
			{
				"?q1=hello&q2=nice",
				map[string][]string{
					"q1": {"hello"},
					"q2": {"nice"},
				},
			},
			{
				"?q1=",
				map[string][]string{"q1": {""}},
			},
			{
				"?q1=hello%5D",
				map[string][]string{"q1": {"hello]"}},
			},
			{
				"?q1=foo&q2=bar&q1=baz",
				map[string][]string{"q1": {"foo", "baz"}, "q2": {"bar"}},
			},
		}

		for _, tc := range tests {
			t.Run(tc.in, func(t *testing.T) {
				in := fmt.Appendf(nil, "GET /endpoint%s HTTP/1.1\r\nHost: localhost\r\n\r\n", tc.in)

				req, err := requestFromRaw(in, defaultRequestSize, t.Context())
				if err != nil {
					t.Errorf("unexpected error %s", err)
				}
				expectBody(t, req.body, "")
				expectUrl(t, req, "/endpoint")
				expectMethod(t, req.mthd, GET)
				if len(req.uri.queryParams.q) != len(tc.want) {
					t.Errorf(`string query = %q, wanted %#q`, req.uri.queryParams, tc.want)
				}
				for key, want := range tc.want {
					actual, exists := req.uri.queryParams.q[key]
					if !exists {
						t.Errorf(`expected %#q to exist`, key)
					}
					if !reflect.DeepEqual(actual, want) {
						t.Errorf(`QueryParam(%#q) = %v, wanted %v`, key, actual, want)
					}
				}
			})
		}
	})

	t.Run("url", func(t *testing.T) {
		tests := []struct {
			in          string
			wantRaw     string
			wantDecoded string
		}{
			{
				in:          "/simple",
				wantRaw:     "/simple",
				wantDecoded: "/simple",
			},
			{
				in:          "missing/leading/slash",
				wantRaw:     "/missing/leading/slash",
				wantDecoded: "/missing/leading/slash",
			},
			{
				in:          `/foo%20bar`,
				wantRaw:     `/foo%20bar`,
				wantDecoded: "/foo bar",
			},
		}

		for _, tc := range tests {
			t.Run(tc.in, func(t *testing.T) {
				in := fmt.Appendf(nil, "GET %s HTTP/1.1\r\nHost: localhost\r\n\r\n", tc.in)

				req, err := requestFromRaw(in, defaultRequestSize, t.Context())
				if err != nil {
					t.Errorf("unexpected error %v", err)
				}
				expectBody(t, req.body, "")
				expectUrl(t, req, tc.wantDecoded)
				expectMethod(t, req.mthd, GET)
				if req.uri.rawPath != tc.wantRaw {
					t.Errorf(`rawPath = %#q, wanted %#q`, req.uri.rawPath, tc.wantRaw)
				}
			})
		}
	})
}

func TestAcceptsContentType(t *testing.T) {
	tests := []struct {
		name   string
		accept []ContentType
		in     ContentType
		want   bool
	}{
		{
			name:   "empty accept",
			accept: []ContentType{},
			in:     CTApplicationJson,
			want:   false,
		},
		{
			name:   "empty input, accept = */*",
			accept: []ContentType{CTAcceptAll},
			in:     ContentType{},
			want:   true,
		},
		{
			name:   "empty input, accept = application/*",
			accept: []ContentType{{part: "application", subtype: "*"}},
			in:     ContentType{},
			want:   false,
		},
		{
			name:   "exact match, single accept list",
			accept: []ContentType{CTTextHtml},
			in:     CTTextHtml,
			want:   true,
		},
		{
			name:   "exact match, multiple accept list",
			accept: []ContentType{CTApplicationGraphQL, CTTextHtml},
			in:     CTTextHtml,
			want:   true,
		},
		{
			name:   "match by subtype, part = *",
			accept: []ContentType{{part: "*", subtype: "javascript"}},
			in:     CTTextJavaScript,
			want:   true,
		},
		{
			name:   "match by part, subtype = *",
			accept: []ContentType{{part: "multipart", subtype: "*"}},
			in:     CTMultipartByteranges,
			want:   true,
		},
		{
			name:   "accept = */*",
			accept: []ContentType{CTAcceptAll},
			in:     CTTextCsv,
			want:   true,
		},
		{
			name:   "accept contains */*",
			accept: []ContentType{CTTextHtml, CTTextCss, CTApplicationFormUrlEncoded, CTAcceptAll},
			in:     CTTextCsv,
			want:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := &Request{accept: tc.accept}

			got := req.AcceptsContentType(tc.in)

			if got != tc.want {
				t.Errorf(`AcceptsContentType(%#q) = %t, wanted %t`, tc.in.string(), got, tc.want)
			}
		})
	}
}

func TestContextValueAs(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(req *Request)
		key      any
		expected string
		ok       bool
	}{
		{
			name:  "missing key returns zero and false",
			setup: func(req *Request) {},
			key:   "userId",
		},
		{
			name: "existing key correct type",
			setup: func(req *Request) {
				req.NewContextValue("userId", "abc123")
			},
			key:      "userId",
			expected: "abc123",
			ok:       true,
		},
		{
			name: "existing key wrong type",
			setup: func(req *Request) {
				req.NewContextValue("count", 42)
			},
			key: "count",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := &Request{ctx: t.Context()}
			tc.setup(req)

			got, ok := ContextValueAs[string](req, tc.key)
			if ok != tc.ok {
				t.Errorf("expected ok=%v, got %v", tc.ok, ok)
			}
			if got != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}
