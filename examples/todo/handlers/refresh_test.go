package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sktylr/routeit"
	"github.com/sktylr/routeit/examples/todo/auth"
	"github.com/sktylr/routeit/examples/todo/db"
)

func TestRefreshTokenHandler(t *testing.T) {
	db.WithUnitTestConnection(t, func(conn *sql.DB, mock sqlmock.Sqlmock) {
		repo := db.NewUsersRepository(conn)
		handler := RefreshTokenHandler(repo)

		validUserID := "3d43d55e-7c89-4695-99e9-4cf50b09ea4d"

		validTokens, err := auth.GenerateTokens(validUserID)
		if err != nil {
			t.Fatalf("could not generate valid refresh token: %v", err)
		}

		expiredToken := func() string {
			claims := jwt.RegisteredClaims{
				Subject:   validUserID,
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			}
			token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
			signed, err := token.SignedString([]byte("super-secret-key"))
			if err != nil {
				t.Fatalf("could not create expired token: %v", err)
			}
			return signed
		}()

		tests := []struct {
			name            string
			refreshToken    string
			mockQuery       func()
			wantErr         error
			expectSuccess   bool
			expectTokenData bool
		}{
			{
				name:      "missing token",
				mockQuery: func() {},
				wantErr:   routeit.ErrUnprocessableContent(),
			},
			{
				name:         "malformed token",
				refreshToken: "invalid.token.here",
				mockQuery:    func() {},
				wantErr:      routeit.ErrUnauthorized(),
			},
			{
				name:         "expired token",
				refreshToken: expiredToken,
				mockQuery:    func() {},
				wantErr:      routeit.ErrUnauthorized(),
			},
			{
				name:         "user not found",
				refreshToken: validTokens.RefreshToken,
				mockQuery: func() {
					mock.ExpectQuery("SELECT id, name, email, password, created, updated FROM users WHERE id = ?").
						WithArgs(validUserID).
						WillReturnRows(sqlmock.NewRows([]string{}))
				},
				wantErr: routeit.ErrUnauthorized(),
			},
			{
				name:         "db error on lookup",
				refreshToken: validTokens.RefreshToken,
				mockQuery: func() {
					mock.ExpectQuery("SELECT id, name, email, password, created, updated FROM users WHERE id = ?").
						WithArgs(validUserID).
						WillReturnError(errors.New("db failure"))
				},
				wantErr: db.ErrDatabaseIssue,
			},
			{
				name:         "using access token instead of refresh",
				refreshToken: validTokens.AccessToken,
				mockQuery:    func() {},
				wantErr:      routeit.ErrUnauthorized(),
			},
			{
				name:         "success",
				refreshToken: validTokens.RefreshToken,
				mockQuery: func() {
					mock.ExpectQuery("SELECT id, name, email, password, created, updated FROM users WHERE id = ?").
						WithArgs(validUserID).
						WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "created", "updated"}).
							AddRow(validUserID, "Alice", "alice@example.com", "somehashedpassword", time.Now(), time.Now()))
				},
				expectSuccess:   true,
				expectTokenData: true,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				tc.mockQuery()
				body, _ := json.Marshal(map[string]string{
					"refresh_token": tc.refreshToken,
				})
				req := routeit.NewTestRequest(t, "/auth/refresh", routeit.POST, routeit.TestRequestOptions{
					Body: body,
					Headers: []string{
						"Content-Type", "application/json",
						"Content-Length", fmt.Sprint(len(body)),
					},
				})

				resp, err := routeit.TestHandler(handler, req)

				if err != nil && tc.expectSuccess {
					t.Fatalf("unexpected error: %v", err)
				}
				if !tc.expectSuccess {
					if err == nil {
						t.Error("no error when expected failure")
					}
					if !errors.Is(err, tc.wantErr) {
						t.Fatalf(`err = %v, wanted %v`, err, tc.wantErr)
					}
					return
				}
				resp.AssertStatusCode(t, routeit.StatusCreated)
				if tc.expectTokenData {
					var data RefreshTokenResponse
					resp.BodyToJson(t, &data)
					if data.AccessToken == "" || data.RefreshToken == "" {
						t.Errorf("expected tokens in response, got: %+v", data)
					}
				}
			})
		}
	})
}
