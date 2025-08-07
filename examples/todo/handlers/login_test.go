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
		name           string
		request        LoginRequest
		mockSetup      func(mock sqlmock.Sqlmock)
		expectedStatus routeit.HttpStatus
		assertBody     func(t *testing.T, res *routeit.TestResponse)
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
			expectedStatus: routeit.StatusUnprocessableContent,
		},
		{
			name: "missing password returns 422",
			request: LoginRequest{
				Email:    "user@example.com",
				Password: "",
			},
			expectedStatus: routeit.StatusUnprocessableContent,
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
			expectedStatus: routeit.StatusNotFound,
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
			expectedStatus: routeit.StatusBadRequest,
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
			expectedStatus: routeit.StatusServiceUnavailable,
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
				wantErr := tc.expectedStatus.Is4xx() || tc.expectedStatus.Is5xx()

				res, err := routeit.TestHandler(handler, req)

				if err != nil {
					if !wantErr {
						t.Fatalf("unexpected error returned: %v", err)
					}
					httpErr, ok := err.(*routeit.HttpError)
					if !ok {
						t.Fatalf("expected HttpError, got %T", err)
					}
					if httpErr.Status() != tc.expectedStatus {
						t.Errorf(`status = %+v, wanted %+v`, httpErr.Status(), tc.expectedStatus)
					}
				} else {
					res.AssertStatusCode(t, tc.expectedStatus)
					tc.assertBody(t, res)
				}
			})
		})
	}
}
