package main

import (
	"testing"

	"github.com/sktylr/routeit"
)

var client = routeit.NewTestClient(GetServer())

func TestHello(t *testing.T) {
	t.Run("unauthorised", func(t *testing.T) {
		tests := []struct {
			name string
			fn   func() *routeit.TestResponse
		}{
			{
				"GET no header",
				func() *routeit.TestResponse { return client.Get("/hello") },
			},
			{
				"GET invalid auth header",
				func() *routeit.TestResponse { return client.Get("/hello", "Authorization", "Bearer 123") },
			},
			{
				// The authorisation is the first point of failure (as opposed
				// to the requested method, which is the second point). For
				// improved security, the first point of failure is returned to
				// the client, ensuring that as little as possible is revealed
				// to the client.
				"POST no header",
				func() *routeit.TestResponse { return client.PostText("/hello", "hello!") },
			},
			{
				"POST invalid auth header",
				func() *routeit.TestResponse {
					return client.PostText("/hello", "hello!", "Authorization", "Bearer 123")
				},
			},
			{
				"GET not found",
				func() *routeit.TestResponse { return client.Get("/goodbye", "Authorization", "Bearer 123") },
			},
			{
				"OPTIONS",
				func() *routeit.TestResponse { return client.Options("/hello") },
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				res := tc.fn()
				res.AssertStatusCode(t, routeit.StatusUnauthorized)
				res.RefuteHeaderPresent(t, "Allow")
			})
		}
	})

	t.Run("authorised", func(t *testing.T) {
		tests := []struct {
			name       string
			fn         func() *routeit.TestResponse
			wantStatus routeit.HttpStatus
			wantAllow  string
		}{
			{
				name:       "GET",
				fn:         func() *routeit.TestResponse { return client.Get("/hello", "Authorization", "LET ME IN") },
				wantStatus: routeit.StatusOK,
			},
			{
				name:       "GET not found",
				fn:         func() *routeit.TestResponse { return client.Get("/goodbye", "Authorization", "LET ME IN") },
				wantStatus: routeit.StatusNotFound,
			},
			{
				name:       "POST",
				fn:         func() *routeit.TestResponse { return client.PostText("/hello", "hello!", "Authorization", "LET ME IN") },
				wantStatus: routeit.StatusMethodNotAllowed,
				wantAllow:  "GET, HEAD, OPTIONS",
			},
			{
				name:       "OPTIONS",
				fn:         func() *routeit.TestResponse { return client.Options("/hello", "Authorization", "LET ME IN") },
				wantStatus: routeit.StatusNoContent,
				wantAllow:  "GET, HEAD, OPTIONS",
			},
			{
				name:       "OPTIONS not found",
				fn:         func() *routeit.TestResponse { return client.Options("/goodbye", "Authorization", "LET ME IN") },
				wantStatus: routeit.StatusNotFound,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				res := tc.fn()
				res.AssertStatusCode(t, tc.wantStatus)
				if tc.wantAllow == "" {
					res.RefuteHeaderPresent(t, "Allow")
				} else {
					res.AssertHeaderMatches(t, "Allow", tc.wantAllow)
				}
			})
		}
	})
}
