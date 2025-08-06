package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sktylr/routeit"
	"github.com/sktylr/routeit/examples/todo/auth"
	"github.com/sktylr/routeit/examples/todo/db"
)

func TestLoginHandler(t *testing.T) {
	tests := []struct {
		name           string
		request        LoginRequest
		mockSetup      func(mock sqlmock.Sqlmock)
		expectedStatus routeit.HttpStatus
		assertBody     func(t *testing.T, res *routeit.TestResponse)
		wantErrMsg     string
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
			expectedStatus: routeit.StatusCreated,
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
			wantErrMsg: "422: Unprocessable Content",
		},
		{
			name: "missing password returns 422",
			request: LoginRequest{
				Email:    "user@example.com",
				Password: "",
			},
			wantErrMsg: "422: Unprocessable Content",
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
			wantErrMsg: "404: Not Found",
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
			wantErrMsg: "400: Bad Request",
		},
		{
			name: "internal db error returns 503",
			request: LoginRequest{
				Email:    "fail@example.com",
				Password: "pass",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, email, password, created, updated FROM users WHERE email = ?")).
					WithArgs("fail@example.com").
					WillReturnError(errors.New("db connection lost"))
			},
			wantErrMsg: "503: Service Unavailable",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db.WithTestConnection(t, func(dbConn *sql.DB, mock sqlmock.Sqlmock) {
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
					if tc.wantErrMsg == "" {
						t.Fatalf("unexpected error returned: %v", err)
					}
					httpErr, ok := err.(*routeit.HttpError)
					if !ok {
						t.Fatalf("expected HttpError, got %T", err)
					}
					if !strings.Contains(httpErr.Error(), tc.wantErrMsg) {
						t.Errorf(`Error() = %#q, wanted %#q`, httpErr.Error(), tc.wantErrMsg)
					}
				} else {
					res.AssertStatusCode(t, tc.expectedStatus)
					tc.assertBody(t, res)
				}
			})
		})
	}
}
