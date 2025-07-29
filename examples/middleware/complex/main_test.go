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

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		headers       []string
		wantProceeded bool
		wantErr       bool
		wantUserId    string
		wantScopes    []string
	}{
		{
			name:          "allows requests to /no-auth without auth headers",
			path:          "/no-auth",
			wantProceeded: true,
		},
		{
			name:          "allows requests to /no-auth with invalid auth headers",
			path:          "/no-auth",
			headers:       []string{"Authorization", "foobar"},
			wantProceeded: true,
		},
		{
			name:          "allows requests to /no-auth with valid auth headers",
			path:          "/no-auth",
			headers:       []string{"Authorization", "Bearer superuser_123"},
			wantProceeded: true,
		},
		{
			name:    "blocks request with no Authorization header",
			path:    "/scoped",
			wantErr: true,
		},
		{
			name:    "blocks request with non-Bearer Authorization header",
			path:    "/scoped",
			headers: []string{"Authorization", "Token abc123"},
			wantErr: true,
		},
		{
			name:          "sets userId and scopes for user with user_ prefix",
			path:          "/scoped",
			headers:       []string{"Authorization", "Bearer user_123"},
			wantProceeded: true,
			wantUserId:    "user_123",
			wantScopes:    []string{"barscope"},
		},
		{
			name:          "sets userId and scopes for user with superuser_ prefix",
			path:          "/scoped",
			headers:       []string{"Authorization", "Bearer superuser_admin"},
			wantProceeded: true,
			wantUserId:    "superuser_admin",
			wantScopes:    []string{"fooscope:write", "fooscope:read", "barscope"},
		},
		{
			name:          "sets empty scopes for unknown user",
			path:          "/scoped",
			headers:       []string{"Authorization", "Bearer anonymous"},
			wantProceeded: true,
			wantUserId:    "anonymous",
			wantScopes:    []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := routeit.NewTestRequest(t, tc.path, routeit.POST, routeit.TestRequest{Headers: tc.headers})

			_, proceeded, err := routeit.TestMiddleware(AuthMiddleware, req)

			if proceeded != tc.wantProceeded {
				t.Errorf("proceeded = %v, want %v", proceeded, tc.wantProceeded)
			}
			if (err != nil) != tc.wantErr {
				t.Errorf("error = %v, want error? %v", err, tc.wantErr)
			}

			if tc.wantUserId != "" {
				userRaw, has := req.ContextValue("userId")
				if !has {
					t.Fatalf("expected context key 'userId' to be present")
				}
				user, ok := userRaw.(string)
				if !ok {
					t.Fatalf("expected userId to be string, got %T", userRaw)
				}
				if user != tc.wantUserId {
					t.Errorf("userId = %q, want %q", user, tc.wantUserId)
				}
			}

			if tc.wantScopes != nil {
				scopesRaw, has := req.ContextValue("scopes")
				if !has {
					t.Fatalf("expected context key 'scopes' to be present")
				}
				scopes, ok := scopesRaw.([]string)
				if !ok {
					t.Fatalf("expected scopes to be []string, got %T", scopesRaw)
				}
				if len(scopes) != len(tc.wantScopes) {
					t.Errorf("scopes length = %d, want %d", len(scopes), len(tc.wantScopes))
				}
				for i := range tc.wantScopes {
					if scopes[i] != tc.wantScopes[i] {
						t.Errorf("scopes[%d] = %q, want %q", i, scopes[i], tc.wantScopes[i])
					}
				}
			}
		})
	}
}

func TestScopesMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		scopes        any
		wantProceeded bool
		wantErr       bool
	}{
		{
			name:          "allows requests to /no-auth without scopes",
			path:          "/no-auth",
			wantProceeded: true,
		},
		{
			name:    "forbids requests without scopes context",
			path:    "/scoped",
			wantErr: true,
		},
		{
			name:    "forbids requests when scopes is not []string",
			path:    "/scoped",
			scopes:  "not-a-slice",
			wantErr: true,
		},
		{
			name:    "forbids when required scopes are missing for /scoped",
			path:    "/scoped",
			scopes:  []string{"fooscope:read"},
			wantErr: true,
		},
		{
			name:          "allows when all required scopes are present for /scoped",
			path:          "/scoped",
			scopes:        []string{"fooscope:read", "fooscope:write"},
			wantProceeded: true,
		},
		{
			name:          "allows any scopes for /scopeless",
			path:          "/scopeless",
			scopes:        []string{"whatever"},
			wantProceeded: true,
		},
		{
			name:    "forbids when barscope missing for default case",
			path:    "/default",
			scopes:  []string{"fooscope:read"},
			wantErr: true,
		},
		{
			name:          "allows when barscope is present for default case",
			path:          "/default",
			scopes:        []string{"barscope"},
			wantProceeded: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := routeit.NewTestRequest(t, tc.path, routeit.GET, routeit.TestRequest{})

			if tc.scopes != nil {
				req.NewContextValue("scopes", tc.scopes)
			}

			_, proceeded, err := routeit.TestMiddleware(ScopesMiddleware, req)

			if proceeded != tc.wantProceeded {
				t.Errorf("proceeded = %v, want %v", proceeded, tc.wantProceeded)
			}
			if (err != nil) != tc.wantErr {
				t.Errorf("error = %v, want error? %v", err, tc.wantErr)
			}
		})
	}
}
