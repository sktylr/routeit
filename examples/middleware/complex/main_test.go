package main

import (
	"testing"

	"github.com/sktylr/routeit"
)

func TestServer(t *testing.T) {
	tests := []struct {
		name       string
		route      string
		headers    []string
		wantStatus routeit.HttpStatus
		wantBody   string
	}{
		{
			name:       "/scoped unauthorised",
			route:      "/scoped",
			wantStatus: routeit.StatusUnauthorized,
		},
		{
			name:       "/scoped Authorization header present, malformed",
			route:      "/scoped",
			headers:    []string{"Authorization", "Bearer"},
			wantStatus: routeit.StatusUnauthorized,
		},
		{
			name:       "/scoped authorised, missing scopes",
			route:      "/scoped",
			headers:    []string{"Authorization", "Bearer user_123"},
			wantStatus: routeit.StatusForbidden,
		},
		{
			name:       "/scoped authorised, correct scopes",
			route:      "/scoped",
			headers:    []string{"Authorization", "Bearer superuser_123"},
			wantStatus: routeit.StatusOK,
			wantBody:   `You are authenticated and have the correct scopes: "fooscope:write", "fooscope:read" and "barscope".`,
		},
		{
			name:       "/scopeless unauthorised",
			route:      "/scopeless",
			wantStatus: routeit.StatusUnauthorized,
		},
		{
			name:       "/scopeless authorised, missing scopes",
			route:      "/scopeless",
			headers:    []string{"Authorization", "Bearer user_123"},
			wantStatus: routeit.StatusOK,
			wantBody:   "You are authenticated and have the following scopes: [barscope]",
		},
		{
			name:       "/scopeless authorised, with all scopes",
			route:      "/scopeless",
			headers:    []string{"Authorization", "Bearer superuser_123"},
			wantStatus: routeit.StatusOK,
			wantBody:   "You are authenticated and have the following scopes: [fooscope:write fooscope:read barscope]",
		},
		{
			name:       "/no-auth without Authorization header",
			route:      "/no-auth",
			wantStatus: routeit.StatusOK,
			wantBody:   "You do not need to be authenticated to reach this endpoint!",
		},
		{
			name:       "/no-auth with Authorization header",
			route:      "/no-auth",
			headers:    []string{"Authorization", "Bearer superuser_123"},
			wantStatus: routeit.StatusOK,
			wantBody:   "You do not need to be authenticated to reach this endpoint!",
		},
		{
			name:       "/hello unauthorised",
			route:      "/hello",
			wantStatus: routeit.StatusUnauthorized,
		},
		{
			name:       "/hello authorised, missing barscope",
			route:      "/hello",
			headers:    []string{"Authorization", "Bearer unknown_user"},
			wantStatus: routeit.StatusForbidden,
		},
		{
			name:       "/hello authorised, has barscope",
			route:      "/hello",
			headers:    []string{"Authorization", "Bearer user_456"},
			wantStatus: routeit.StatusOK,
			wantBody:   `You need to be authenticated and have "barscopes" to reach this endpoint. You have [barscope] scopes`,
		},
		{
			name:       "/not-found without auth",
			route:      "/not-found",
			wantStatus: routeit.StatusUnauthorized,
		},
		{
			name:       "/not-found with invalid auth",
			route:      "/not-found",
			headers:    []string{"Authorization", "Bearer"},
			wantStatus: routeit.StatusUnauthorized,
		},
		{
			name:       "/not-found with valid auth but missing scopes",
			route:      "/not-found",
			headers:    []string{"Authorization", "Bearer unknown_user"},
			wantStatus: routeit.StatusForbidden,
		},
		{
			name:       "/not-found with full access",
			route:      "/not-found",
			headers:    []string{"Authorization", "Bearer superuser_999"},
			wantStatus: routeit.StatusNotFound,
		},
	}
	client := routeit.NewTestClient(GetServer())

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("GET", func(t *testing.T) {
				res := client.Get(tc.route, tc.headers...)

				res.AssertStatusCode(t, tc.wantStatus)
				if tc.wantBody != "" {
					res.AssertBodyMatchesString(t, tc.wantBody)
				}
			})

			t.Run("HEAD", func(t *testing.T) {
				res := client.Head(tc.route, tc.headers...)

				res.AssertStatusCode(t, tc.wantStatus)
				res.AssertBodyEmpty(t)
			})
		})
	}
}
