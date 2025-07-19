package main

import (
	"testing"

	"github.com/sktylr/routeit"
)

func TestGetEndpoints(t *testing.T) {
	tests := []struct {
		route string
		want  routeit.HttpStatus
	}{
		{
			"/invalid",
			routeit.StatusInternalServerError,
		},
		{
			"/forbidden",
			routeit.StatusForbidden,
		},
	}
	client := routeit.NewTestClient(GetServer())

	for _, tc := range tests {
		t.Run(tc.route, func(t *testing.T) {
			t.Run("GET", func(t *testing.T) {
				res := client.Get(tc.route)
				res.AssertStatusCode(t, tc.want)
			})

			t.Run("HEAD", func(t *testing.T) {
				res := client.Head(tc.route)
				res.AssertStatusCode(t, tc.want)
			})

			t.Run("POST", func(t *testing.T) {
				res := client.PostText(tc.route, "Hello")
				res.AssertStatusCode(t, routeit.StatusMethodNotAllowed)
				res.AssertHeaderMatches(t, "Allow", "GET, HEAD, OPTIONS")
			})
		})
	}
}

func TestLogin(t *testing.T) {
	client := routeit.NewTestClient(GetServer())

	t.Run("POST", func(t *testing.T) {
		t.Run("happy", func(t *testing.T) {
			in := LoginRequest{
				Username: "user@email.com",
				Password: "Password123!",
			}
			want := LoginResponse{
				AccessToken:  "access_123",
				RefreshToken: "refresh_123",
			}

			res := client.PostJson("/login", in)

			res.AssertStatusCode(t, routeit.StatusOK)
			var out LoginResponse
			res.BodyToJson(t, &out)
			if out != want {
				t.Errorf("body = %#q, wanted %#q", out, want)
			}
		})

		t.Run("failures", func(t *testing.T) {
			t.Run("json, bad body", func(t *testing.T) {
				tests := []struct {
					name string
					in   LoginRequest
					want routeit.HttpStatus
				}{
					{
						name: "missing username",
						in:   LoginRequest{Password: "Password123!"},
						want: routeit.StatusUnprocessableContent,
					},
					{
						name: "missing password",
						in:   LoginRequest{Username: "user@email.com"},
						want: routeit.StatusUnprocessableContent,
					},
					{
						name: "incorrect password",
						in:   LoginRequest{Username: "user@email.com", Password: "WrongPassword!"},
						want: routeit.StatusBadRequest,
					},
				}

				for _, tc := range tests {
					t.Run(tc.name, func(t *testing.T) {
						res := client.PostJson("/login", tc.in)
						res.AssertStatusCode(t, tc.want)
					})
				}
			})

			t.Run("text/plain", func(t *testing.T) {
				res := client.PostText("/login", "username=user@email.com,password=Password123!")

				res.AssertStatusCode(t, routeit.StatusUnsupportedMediaType)
				res.AssertHeaderMatches(t, "Accept", "application/json")
			})

			t.Run("method not allowed", func(t *testing.T) {
				t.Run("GET", func(t *testing.T) {
					res := client.Get("/login")
					res.AssertStatusCode(t, routeit.StatusMethodNotAllowed)
					res.AssertHeaderMatches(t, "Allow", "POST, OPTIONS")
				})

				t.Run("HEAD", func(t *testing.T) {
					res := client.Head("/login")
					res.AssertStatusCode(t, routeit.StatusMethodNotAllowed)
					res.AssertHeaderMatches(t, "Allow", "POST, OPTIONS")
				})
			})
		})
	})
}
