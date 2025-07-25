package main

import (
	"testing"

	"github.com/sktylr/routeit"
)

func TestPreflight(t *testing.T) {
	tests := []struct {
		name             string
		origin           string
		method           string
		path             string
		headers          []string
		wantStatus       routeit.HttpStatus
		wantAllowOrigin  bool
		wantAllowMethod  bool
		wantAllowHeaders string
		wantVary         string
		wantMaxAge       bool
	}{
		{
			name:            "valid preflight to /create with POST (default allowed method)",
			origin:          "http://localhost:3000",
			method:          "POST",
			path:            "/create",
			wantStatus:      routeit.StatusNoContent,
			wantAllowOrigin: true,
			wantAllowMethod: true,
			wantVary:        "Origin, Access-Control-Request-Method",
			wantMaxAge:      true,
		},
		{
			name:            "valid preflight to /simple with GET",
			origin:          "http://localhost:3000",
			method:          "GET",
			path:            "/simple",
			wantStatus:      routeit.StatusNoContent,
			wantAllowOrigin: true,
			wantAllowMethod: true,
			wantVary:        "Origin, Access-Control-Request-Method",
			wantMaxAge:      true,
		},
		{
			name:            "invalid origin (missing port)",
			origin:          "http://localhost",
			method:          "POST",
			path:            "/create",
			wantStatus:      routeit.StatusForbidden,
			wantAllowOrigin: false,
			wantVary:        "Origin",
		},
		{
			name:            "invalid origin (wrong host)",
			origin:          "http://example.com",
			method:          "POST",
			path:            "/create",
			wantStatus:      routeit.StatusForbidden,
			wantAllowOrigin: false,
			wantVary:        "Origin",
		},
		{
			name:            "disallowed method (PATCH) despite route being defined",
			origin:          "http://localhost:3000",
			method:          "PATCH",
			path:            "/update",
			wantStatus:      routeit.StatusMethodNotAllowed,
			wantAllowOrigin: false,
			wantVary:        "Origin, Access-Control-Request-Method",
		},
		{
			name:            "CORS allowed method (PUT), but route does not support it",
			origin:          "http://localhost:3000",
			method:          "PUT",
			path:            "/simple",
			wantStatus:      routeit.StatusMethodNotAllowed,
			wantAllowOrigin: false,
			wantVary:        "Origin, Access-Control-Request-Method",
		},
		{
			name:            "valid CORS method (DELETE), but route does not support it",
			origin:          "http://localhost:3000",
			method:          "DELETE",
			path:            "/create",
			wantStatus:      routeit.StatusMethodNotAllowed,
			wantAllowOrigin: false,
			wantVary:        "Origin, Access-Control-Request-Method",
		},
		{
			name:             "valid preflight with custom header",
			origin:           "http://localhost:1234",
			method:           "POST",
			path:             "/simple",
			headers:          []string{"Access-Control-Request-Headers", "X-Custom-Header"},
			wantStatus:       routeit.StatusNoContent,
			wantAllowOrigin:  true,
			wantAllowMethod:  true,
			wantAllowHeaders: "X-Custom-Header",
			wantVary:         "Origin, Access-Control-Request-Method, Access-Control-Request-Headers",
			wantMaxAge:       true,
		},
		{
			name:             "valid preflight with multiple headers, all allowed",
			origin:           "http://localhost:1234",
			method:           "POST",
			path:             "/simple",
			headers:          []string{"Access-Control-Request-Headers", "Content-Type, X-Custom-Header"},
			wantStatus:       routeit.StatusNoContent,
			wantAllowOrigin:  true,
			wantAllowMethod:  true,
			wantAllowHeaders: "Content-Type, X-Custom-Header",
			wantVary:         "Origin, Access-Control-Request-Method, Access-Control-Request-Headers",
			wantMaxAge:       true,
		},
		{
			name:            "invalid preflight with disallowed custom header",
			origin:          "http://localhost:1234",
			method:          "POST",
			path:            "/simple",
			headers:         []string{"Access-Control-Request-Headers", "X-Disallowed-Header"},
			wantStatus:      routeit.StatusNoContent,
			wantAllowOrigin: true,
			wantAllowMethod: true,
			wantVary:        "Origin, Access-Control-Request-Method",
			wantMaxAge:      true,
		},
		{
			name:            "valid preflight with multiple headers, some disallowed",
			origin:          "http://localhost:1234",
			method:          "POST",
			path:            "/simple",
			headers:         []string{"Access-Control-Request-Headers", "X-Bad, X-Custom-Header, X-Disallowed-Header"},
			wantStatus:      routeit.StatusNoContent,
			wantAllowOrigin: true,
			wantAllowMethod: true,
			wantVary:        "Origin, Access-Control-Request-Method",
			wantMaxAge:      true,
		},
		{
			name:            "valid preflight with multiple headers, all disallowed",
			origin:          "http://localhost:1234",
			method:          "POST",
			path:            "/simple",
			headers:         []string{"Access-Control-Request-Headers", "X-Bad, X-Disallowed-Header"},
			wantStatus:      routeit.StatusNoContent,
			wantAllowOrigin: true,
			wantAllowMethod: true,
			wantVary:        "Origin, Access-Control-Request-Method",
			wantMaxAge:      true,
		},
		{
			name:            "valid preflight to /simple with HEAD (default allowed)",
			origin:          "http://localhost:9999",
			method:          "HEAD",
			path:            "/simple",
			wantStatus:      routeit.StatusNoContent,
			wantAllowOrigin: true,
			wantAllowMethod: true,
			wantVary:        "Origin, Access-Control-Request-Method",
			wantMaxAge:      true,
		},
		{
			name:            "preflight with method not supported by route (TRACE)",
			origin:          "http://localhost:9999",
			method:          "TRACE",
			path:            "/create",
			wantStatus:      routeit.StatusMethodNotAllowed,
			wantAllowOrigin: false,
			wantVary:        "Origin, Access-Control-Request-Method",
		},
		{
			name:            "valid preflight to /remove with DELETE",
			origin:          "http://localhost:3000",
			method:          "DELETE",
			path:            "/remove",
			wantStatus:      routeit.StatusNoContent,
			wantAllowOrigin: true,
			wantAllowMethod: true,
			wantVary:        "Origin, Access-Control-Request-Method",
			wantMaxAge:      true,
		},
		{
			name:            "invalid preflight to /remove with unsupported method (PATCH)",
			origin:          "http://localhost:3000",
			method:          "PATCH",
			path:            "/remove",
			wantStatus:      routeit.StatusMethodNotAllowed,
			wantAllowOrigin: false,
			wantAllowMethod: false,
			wantVary:        "Origin, Access-Control-Request-Method",
		},
		{
			name:            "invalid preflight to /remove with unsupported method (FOOBAR)",
			origin:          "http://localhost:3000",
			method:          "FOOBAR",
			path:            "/remove",
			wantStatus:      routeit.StatusMethodNotAllowed,
			wantAllowOrigin: false,
			wantAllowMethod: false,
			wantVary:        "Origin, Access-Control-Request-Method",
		},
		{
			name:            "invalid origin for /remove (wrong domain)",
			origin:          "http://evil.com",
			method:          "DELETE",
			path:            "/remove",
			wantStatus:      routeit.StatusForbidden,
			wantAllowOrigin: false,
			wantVary:        "Origin",
		},
		{
			name:             "valid preflight to /remove with DELETE and custom header",
			origin:           "http://localhost:3000",
			method:           "DELETE",
			path:             "/remove",
			headers:          []string{"Access-Control-Request-Headers", "X-Custom-Header"},
			wantStatus:       routeit.StatusNoContent,
			wantAllowOrigin:  true,
			wantAllowMethod:  true,
			wantAllowHeaders: "X-Custom-Header",
			wantVary:         "Origin, Access-Control-Request-Method, Access-Control-Request-Headers",
			wantMaxAge:       true,
		},
	}
	client := routeit.NewTestClient(GetServer())

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			baseHeaders := []string{
				"Origin", tc.origin,
				"Access-Control-Request-Method", tc.method,
			}
			h := append(baseHeaders, tc.headers...)

			res := client.Options(tc.path, h...)

			res.AssertStatusCode(t, tc.wantStatus)
			if tc.wantAllowOrigin {
				res.AssertHeaderMatches(t, "Access-Control-Allow-Origin", tc.origin)
				res.AssertBodyEmpty(t)
			} else {
				res.RefuteHeaderPresent(t, "Access-Control-Allow-Origin")
			}
			if tc.wantAllowMethod {
				res.AssertHeaderMatches(t, "Access-Control-Allow-Methods", tc.method)
			} else {
				res.RefuteHeaderPresent(t, "Access-Control-Allow-Methods")
			}
			if tc.wantAllowHeaders != "" {
				res.AssertHeaderMatches(t, "Access-Control-Allow-Headers", tc.wantAllowHeaders)
			} else {
				res.RefuteHeaderPresent(t, "Access-Control-Allow-Headers")
			}
			res.AssertHeaderMatches(t, "Vary", tc.wantVary)
			if tc.wantMaxAge {
				res.AssertHeaderMatches(t, "Access-Control-Max-Age", "15")
			} else {
				res.RefuteHeaderPresent(t, "Access-Control-Max-Age")
			}
		})
	}
}

func TestActualRequests(t *testing.T) {
	client := routeit.NewTestClient(GetServer())
	tests := []struct {
		name              string
		doRequest         func() *routeit.TestResponse
		wantStatus        routeit.HttpStatus
		wantBody          string
		wantAllow         string
		wantAllowOrigin   string
		wantVary          string
		wantExposeHeaders string
	}{
		{
			name: "GET /simple without Origin",
			doRequest: func() *routeit.TestResponse {
				return client.Get("/simple")
			},
			wantStatus: routeit.StatusOK,
			wantBody:   "Hello from GET simple!",
			wantVary:   "Origin",
		},
		{
			name: "GET /simple with valid Origin",
			doRequest: func() *routeit.TestResponse {
				return client.Get("/simple", "Origin", "http://localhost:3000")
			},
			wantStatus:        routeit.StatusOK,
			wantBody:          "Hello from GET simple!",
			wantAllowOrigin:   "http://localhost:3000",
			wantVary:          "Origin",
			wantExposeHeaders: "X-Response-Header",
		},
		{
			name: "GET /simple with invalid Origin",
			doRequest: func() *routeit.TestResponse {
				return client.Get("/simple", "Origin", "http://evil.com")
			},
			wantStatus:      routeit.StatusForbidden,
			wantAllowOrigin: "",
			wantVary:        "Origin",
		},
		{
			name: "DELETE /create without Origin (disallowed method)",
			doRequest: func() *routeit.TestResponse {
				return client.Delete("/create")
			},
			wantStatus: routeit.StatusMethodNotAllowed,
			wantAllow:  "POST, OPTIONS",
			wantVary:   "Origin",
		},
		{
			name: "DELETE /create with valid Origin (method not supported)",
			doRequest: func() *routeit.TestResponse {
				return client.Delete("/create", "Origin", "http://localhost:3000")
			},
			wantStatus: routeit.StatusMethodNotAllowed,
			wantAllow:  "POST, OPTIONS",
			// This tells the client that the requests can be made from their
			// origin to the server's origin, but not with the requested method
			wantAllowOrigin:   "http://localhost:3000",
			wantVary:          "Origin",
			wantExposeHeaders: "X-Response-Header",
		},
		{
			name: "DELETE /remove with valid Origin (allowed)",
			doRequest: func() *routeit.TestResponse {
				return client.Delete("/remove", "Origin", "http://localhost:3000")
			},
			wantStatus:        routeit.StatusNoContent,
			wantBody:          "Deleted!",
			wantAllowOrigin:   "http://localhost:3000",
			wantVary:          "Origin",
			wantExposeHeaders: "X-Response-Header",
		},
		{
			name: "DELETE /remove with invalid Origin (rejected)",
			doRequest: func() *routeit.TestResponse {
				return client.Delete("/remove", "Origin", "http://attacker.com")
			},
			wantStatus:      routeit.StatusForbidden,
			wantAllowOrigin: "",
			wantVary:        "Origin",
		},
		{
			name: "PATCH /update without Origin",
			doRequest: func() *routeit.TestResponse {
				return client.PatchText("/update", "body")
			},
			wantStatus: routeit.StatusOK,
			wantVary:   "Origin",
			wantBody:   "Hello from PATCH /update!\n",
		},
		{
			name: "PATCH /update with valid Origin",
			doRequest: func() *routeit.TestResponse {
				return client.PatchText("/update", "body", "Origin", "http://localhost:3000")
			},
			wantStatus:        routeit.StatusOK,
			wantAllowOrigin:   "http://localhost:3000",
			wantVary:          "Origin",
			wantBody:          "Hello from PATCH /update!\n",
			wantExposeHeaders: "X-Response-Header",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.doRequest()

			res.AssertStatusCode(t, tc.wantStatus)
			if tc.wantBody != "" {
				res.AssertBodyContainsString(t, tc.wantBody)
			}
			if tc.wantAllow != "" {
				res.AssertHeaderMatches(t, "Allow", tc.wantAllow)
			} else {
				res.RefuteHeaderPresent(t, "Allow")
			}
			if tc.wantAllowOrigin != "" {
				res.AssertHeaderMatches(t, "Access-Control-Allow-Origin", tc.wantAllowOrigin)
			} else {
				res.RefuteHeaderPresent(t, "Access-Control-Allow-Origin")
			}
			if tc.wantVary != "" {
				res.AssertHeaderMatches(t, "Vary", tc.wantVary)
			} else {
				res.RefuteHeaderPresent(t, "Vary")
			}
			if tc.wantExposeHeaders != "" {
				res.AssertHeaderMatches(t, "Access-Control-Expose-Headers", tc.wantExposeHeaders)
			} else {
				res.RefuteHeaderPresent(t, "Access-Control-Expose-Headers")
			}
		})
	}
}
