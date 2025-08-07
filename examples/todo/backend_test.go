package main

import (
	"database/sql"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sktylr/routeit"
	"github.com/sktylr/routeit/examples/todo/auth"
	"github.com/sktylr/routeit/examples/todo/db"
	"github.com/sktylr/routeit/examples/todo/handlers"
)

func TestAuthRegister(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db.WithIntegrationTestConnection(t, func(d *sql.DB) {
			client := routeit.NewTestClient(GetBackendServer(d))
			reqBody := handlers.RegisterUserRequest{
				Name:            "Alice",
				Email:           "alice@example.com",
				Password:        "Secret123",
				ConfirmPassword: "Secret123",
			}

			res := client.PostJson("/auth/register", reqBody)

			res.AssertStatusCode(t, routeit.StatusCreated)
			var resp handlers.RegisterUserResponse
			res.BodyToJson(t, &resp)
			if resp.AccessToken == "" || resp.RefreshToken == "" {
				t.Errorf("expected tokens, got %+v", resp)
			}
		})
	})

	t.Run("missing fields returns 422", func(t *testing.T) {
		db.WithIntegrationTestConnection(t, func(d *sql.DB) {
			client := routeit.NewTestClient(GetBackendServer(d))
			reqBody := handlers.RegisterUserRequest{
				Email:           "bob@example.com",
				Password:        "Secret123",
				ConfirmPassword: "Secret123",
			}

			res := client.PostJson("/auth/register", reqBody)
			res.AssertStatusCode(t, routeit.StatusUnprocessableContent)
		})
	})

	t.Run("password mismatch returns 400", func(t *testing.T) {
		db.WithIntegrationTestConnection(t, func(d *sql.DB) {
			client := routeit.NewTestClient(GetBackendServer(d))
			reqBody := handlers.RegisterUserRequest{
				Name:            "Bob",
				Email:           "bob@example.com",
				Password:        "abc123",
				ConfirmPassword: "123abc",
			}

			res := client.PostJson("/auth/register", reqBody)
			res.AssertStatusCode(t, routeit.StatusBadRequest)
			res.AssertBodyContainsString(t, "Password does not match")
		})
	})

	t.Run("invalid email format returns 400", func(t *testing.T) {
		db.WithIntegrationTestConnection(t, func(d *sql.DB) {
			client := routeit.NewTestClient(GetBackendServer(d))
			reqBody := handlers.RegisterUserRequest{
				Name:            "Daisy",
				Email:           "invalid-email",
				Password:        "Secret123",
				ConfirmPassword: "Secret123",
			}

			res := client.PostJson("/auth/register", reqBody)
			res.AssertStatusCode(t, routeit.StatusBadRequest)
			res.AssertBodyContainsString(t, "Invalid email address format")
		})
	})

	t.Run("duplicate email returns 400", func(t *testing.T) {
		db.WithIntegrationTestConnection(t, func(d *sql.DB) {
			client := routeit.NewTestClient(GetBackendServer(d))
			reqBody := handlers.RegisterUserRequest{
				Name:            "Charlie",
				Email:           "charlie@example.com",
				Password:        "Password123",
				ConfirmPassword: "Password123",
			}
			res := client.PostJson("/auth/register", reqBody)
			res.AssertStatusCode(t, routeit.StatusCreated)

			res2 := client.PostJson("/auth/register", reqBody)
			res2.AssertStatusCode(t, routeit.StatusBadRequest)
		})
	})
}

func TestAuthLogin(t *testing.T) {
	registerUser := func(t *testing.T, client routeit.TestClient, email, password string) {
		reqBody := handlers.RegisterUserRequest{
			Name:            "Eve",
			Email:           email,
			Password:        password,
			ConfirmPassword: password,
		}
		res := client.PostJson("/auth/register", reqBody)
		res.AssertStatusCode(t, routeit.StatusCreated)
	}

	t.Run("success", func(t *testing.T) {
		db.WithIntegrationTestConnection(t, func(d *sql.DB) {
			client := routeit.NewTestClient(GetBackendServer(d))
			email := "eve@example.com"
			password := "StrongPass1!"
			registerUser(t, client, email, password)

			loginReq := handlers.LoginRequest{
				Email:    email,
				Password: password,
			}
			res := client.PostJson("/auth/login", loginReq)
			res.AssertStatusCode(t, routeit.StatusCreated)

			var resp handlers.LoginResponse
			res.BodyToJson(t, &resp)

			if resp.AccessToken == "" || resp.RefreshToken == "" {
				t.Errorf("expected tokens, got %+v", resp)
			}
		})
	})

	t.Run("missing fields returns 422", func(t *testing.T) {
		db.WithIntegrationTestConnection(t, func(d *sql.DB) {
			client := routeit.NewTestClient(GetBackendServer(d))
			loginReq := handlers.LoginRequest{
				Email:    "",
				Password: "",
			}
			res := client.PostJson("/auth/login", loginReq)
			res.AssertStatusCode(t, routeit.StatusUnprocessableContent)
		})
	})

	t.Run("user not found returns 404", func(t *testing.T) {
		db.WithIntegrationTestConnection(t, func(d *sql.DB) {
			client := routeit.NewTestClient(GetBackendServer(d))
			loginReq := handlers.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "doesnotmatter",
			}
			res := client.PostJson("/auth/login", loginReq)
			res.AssertStatusCode(t, routeit.StatusNotFound)
		})
	})

	t.Run("wrong password returns 400", func(t *testing.T) {
		db.WithIntegrationTestConnection(t, func(d *sql.DB) {
			client := routeit.NewTestClient(GetBackendServer(d))
			email := "frank@example.com"
			password := "correctpass"
			registerUser(t, client, email, password)

			loginReq := handlers.LoginRequest{
				Email:    email,
				Password: "wrongpass",
			}
			res := client.PostJson("/auth/login", loginReq)
			res.AssertStatusCode(t, routeit.StatusBadRequest)
		})
	})
}

func TestAuthRefresh(t *testing.T) {
	registerAndLogin := func(t *testing.T, client routeit.TestClient, email, password string) handlers.LoginResponse {
		registerReq := handlers.RegisterUserRequest{
			Name:            "Test User",
			Email:           email,
			Password:        password,
			ConfirmPassword: password,
		}
		regRes := client.PostJson("/auth/register", registerReq)
		regRes.AssertStatusCode(t, routeit.StatusCreated)

		loginReq := handlers.LoginRequest{
			Email:    email,
			Password: password,
		}
		loginRes := client.PostJson("/auth/login", loginReq)
		loginRes.AssertStatusCode(t, routeit.StatusCreated)

		var tokens handlers.LoginResponse
		loginRes.BodyToJson(t, &tokens)
		return tokens
	}

	t.Run("valid refresh token returns new tokens", func(t *testing.T) {
		db.WithIntegrationTestConnection(t, func(d *sql.DB) {
			client := routeit.NewTestClient(GetBackendServer(d))
			tokens := registerAndLogin(t, client, "refresh@example.com", "StrongPass123!")

			refreshReq := handlers.RefreshTokenRequest{
				RefreshToken: tokens.RefreshToken,
			}
			res := client.PostJson("/auth/refresh", refreshReq)
			res.AssertStatusCode(t, routeit.StatusCreated)

			var newTokens handlers.RefreshTokenResponse
			res.BodyToJson(t, &newTokens)

			if newTokens.AccessToken == "" || newTokens.RefreshToken == "" {
				t.Errorf("expected new tokens, got %+v", newTokens)
			}
		})
	})

	t.Run("missing token returns 422", func(t *testing.T) {
		db.WithIntegrationTestConnection(t, func(d *sql.DB) {
			client := routeit.NewTestClient(GetBackendServer(d))

			refreshReq := handlers.RefreshTokenRequest{}
			res := client.PostJson("/auth/refresh", refreshReq)

			res.AssertStatusCode(t, routeit.StatusUnprocessableContent)
		})
	})

	t.Run("invalid token returns 401", func(t *testing.T) {
		db.WithIntegrationTestConnection(t, func(d *sql.DB) {
			client := routeit.NewTestClient(GetBackendServer(d))

			refreshReq := handlers.RefreshTokenRequest{
				RefreshToken: "not-a-real-token",
			}
			res := client.PostJson("/auth/refresh", refreshReq)

			res.AssertStatusCode(t, routeit.StatusUnauthorized)
		})
	})

	t.Run("expired token returns 401", func(t *testing.T) {
		db.WithIntegrationTestConnection(t, func(d *sql.DB) {
			client := routeit.NewTestClient(GetBackendServer(d))
			email := "expired@example.com"
			password := "SecurePass!"
			registerAndLogin(t, client, email, password)

			expired := time.Now().Add(-1 * time.Hour)
			claims := auth.Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(expired),
					Issuer:    "todo-sample-app",
					IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
					Subject:   "fake-user-id",
				},
				Type: "refresh",
			}
			token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
			expiredToken, err := token.SignedString([]byte("super-secret-key"))
			if err != nil {
				t.Fatalf("failed to sign expired token: %v", err)
			}

			refreshReq := handlers.RefreshTokenRequest{
				RefreshToken: expiredToken,
			}
			res := client.PostJson("/auth/refresh", refreshReq)
			res.AssertStatusCode(t, routeit.StatusUnauthorized)
		})
	})

	t.Run("refresh token for nonexistent user returns 401", func(t *testing.T) {
		db.WithIntegrationTestConnection(t, func(d *sql.DB) {
			client := routeit.NewTestClient(GetBackendServer(d))

			claims := auth.Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
					Issuer:    "todo-sample-app",
					IssuedAt:  jwt.NewNumericDate(time.Now()),
					Subject:   "nonexistent-user-id",
				},
				Type: "refresh",
			}
			token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
			fakeToken, err := token.SignedString([]byte("super-secret-key"))
			if err != nil {
				t.Fatalf("failed to sign token: %v", err)
			}

			refreshReq := handlers.RefreshTokenRequest{
				RefreshToken: fakeToken,
			}
			res := client.PostJson("/auth/refresh", refreshReq)
			res.AssertStatusCode(t, routeit.StatusUnauthorized)
		})
	})
}
