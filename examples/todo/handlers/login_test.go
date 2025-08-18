package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sktylr/routeit"
	"github.com/sktylr/routeit/examples/todo/auth"
	"github.com/sktylr/routeit/examples/todo/db"
)

func TestLoginHandler(t *testing.T) {
	tests := []struct {
		name       string
		request    LoginRequest
		mockSetup  func(mock sqlmock.Sqlmock)
		wantErr    error
		assertBody func(t *testing.T, res *routeit.TestResponse)
	}{
		{
			name: "valid login returns tokens",
			request: LoginRequest{
				Email:    "valid@example.com",
				Password: "CorrectPassword",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				hashedPw, _ := auth.HashPassword("CorrectPassword")
				mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, email, password, created, updated FROM users WHERE email = ?")).
					WithArgs("valid@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "created", "updated"}).
						AddRow("user-123", "Valid User", "valid@example.com", hashedPw, time.Now(), time.Now()))
			},
			assertBody: func(t *testing.T, res *routeit.TestResponse) {
				var body LoginResponse
				res.BodyToJson(t, &body)
				if body.AccessToken == "" || body.RefreshToken == "" {
					t.Errorf("expected tokens, got: %+v", body)
				}
			},
		},
		{
			name: "missing email returns 422",
			request: LoginRequest{
				Email:    "",
				Password: "pass",
			},
			wantErr: routeit.ErrUnprocessableContent(),
		},
		{
			name: "missing password returns 422",
			request: LoginRequest{
				Email:    "user@example.com",
				Password: "",
			},
			wantErr: routeit.ErrUnprocessableContent(),
		},
		{
			name: "user not found returns 404",
			request: LoginRequest{
				Email:    "ghost@example.com",
				Password: "irrelevant",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, email, password, created, updated FROM users WHERE email = ?")).
					WithArgs("ghost@example.com").
					WillReturnRows(sqlmock.NewRows([]string{}))
			},
			wantErr: routeit.ErrNotFound(),
		},
		{
			name: "password mismatch returns 400",
			request: LoginRequest{
				Email:    "bob@example.com",
				Password: "wrongpassword",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				hashed, _ := auth.HashPassword("correctpassword")
				mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, email, password, created, updated FROM users WHERE email = ?")).
					WithArgs("bob@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "created", "updated"}).
						AddRow("bob-123", "Bob", "bob@example.com", hashed, time.Now(), time.Now()))
			},
			wantErr: routeit.ErrBadRequest(),
		},
		{
			name: "internal db error is propagated",
			request: LoginRequest{
				Email:    "fail@example.com",
				Password: "pass",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, email, password, created, updated FROM users WHERE email = ?")).
					WithArgs("fail@example.com").
					WillReturnError(errors.New("db connection lost"))
			},
			wantErr: db.ErrDatabaseIssue,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db.WithUnitTestConnection(t, func(dbConn *sql.DB, mock sqlmock.Sqlmock) {
				if tc.mockSetup != nil {
					tc.mockSetup(mock)
				}
				repo := db.NewUsersRepository(dbConn)
				handler := LoginHandler(repo)
				bodyBytes, err := json.Marshal(tc.request)
				if err != nil {
					t.Fatalf("failed to marshal request: %v", err)
				}
				req := routeit.NewTestRequest(t, "/auth/login", routeit.POST, routeit.TestRequestOptions{
					Body: bodyBytes,
					Headers: []string{
						"Content-Type", "application/json",
						"Content-Length", fmt.Sprintf("%d", len(bodyBytes)),
					},
				})

				res, err := routeit.TestHandler(handler, req)

				if err != nil {
					if tc.wantErr == nil {
						t.Fatalf("unexpected error returned: %v", err)
					}
					if !errors.Is(err, tc.wantErr) {
						t.Fatalf(`error = %v, wanted %v`, err, tc.wantErr)
					}
				} else {
					res.AssertStatusCode(t, routeit.StatusCreated)
					tc.assertBody(t, res)
				}
			})
		})
	}
}
