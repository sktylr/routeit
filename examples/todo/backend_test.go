package main

import (
	"database/sql"
	"testing"

	"github.com/sktylr/routeit"
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
