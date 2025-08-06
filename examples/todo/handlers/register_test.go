package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/sktylr/routeit"
	"github.com/sktylr/routeit/examples/todo/db"
)

func TestRegisterUserHandler(t *testing.T) {
	tests := []struct {
		name           string
		request        RegisterUserRequest
		mockSetup      func(mock sqlmock.Sqlmock)
		expectedStatus routeit.HttpStatus
		assertBody     func(t *testing.T, res *routeit.TestResponse)
		wantErrMsg     string
	}{
		{
			name: "valid request returns tokens",
			request: RegisterUserRequest{
				Name:            "Alice",
				Email:           "alice@example.com",
				Password:        "Secure123",
				ConfirmPassword: "Secure123",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(
					"INSERT INTO users (id, name, email, password, created, updated) VALUES (?, ?, ?, ?, ?, ?)",
				)).WithArgs(
					sqlmock.AnyArg(), "Alice", "alice@example.com",
					sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				).WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: routeit.StatusCreated,
			assertBody: func(t *testing.T, res *routeit.TestResponse) {
				var body RegisterUserResponse
				res.BodyToJson(t, &body)
				if body.AccessToken == "" || body.RefreshToken == "" {
					t.Errorf("expected tokens, got: %+v", body)
				}
			},
		},
		{
			name: "missing fields returns 422",
			request: RegisterUserRequest{
				Email:           "bob@example.com",
				Password:        "Secret!",
				ConfirmPassword: "Secret!",
			},
			mockSetup:  nil,
			wantErrMsg: "422: Unprocessable Content",
		},
		{
			name: "password mismatch returns 400",
			request: RegisterUserRequest{
				Name:            "Bob",
				Email:           "bob@example.com",
				Password:        "abc123",
				ConfirmPassword: "123abc",
			},
			mockSetup:  nil,
			wantErrMsg: "Password does not match",
		},
		{
			name: "duplicate email returns 400",
			request: RegisterUserRequest{
				Name:            "Charlie",
				Email:           "charlie@example.com",
				Password:        "charlie123",
				ConfirmPassword: "charlie123",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(
					"INSERT INTO users (id, name, email, password, created, updated) VALUES (?, ?, ?, ?, ?, ?)",
				)).WithArgs(
					sqlmock.AnyArg(), "Charlie", "charlie@example.com",
					sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				).WillReturnError(&mysql.MySQLError{
					Number:  1062,
					Message: "Duplicate entry",
				})
			},
			wantErrMsg: "400: Bad Request",
		},
		{
			name: "internal DB error returns 503",
			request: RegisterUserRequest{
				Name:            "Dave",
				Email:           "dave@example.com",
				Password:        "davepass",
				ConfirmPassword: "davepass",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(
					"INSERT INTO users (id, name, email, password, created, updated) VALUES (?, ?, ?, ?, ?, ?)",
				)).WithArgs(
					sqlmock.AnyArg(), "Dave", "dave@example.com",
					sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				).WillReturnError(errors.New("DB connection lost"))
			},
			wantErrMsg: "503: Service Unavailable",
		}, {
			name: "invalid email format",
			request: RegisterUserRequest{
				Name:            "Daisy",
				Email:           "not-an-email",
				Password:        "Secret123",
				ConfirmPassword: "Secret123",
			},
			wantErrMsg: "400: Bad Request. Invalid email address format.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db.WithUnitTestConnection(t, func(dbConn *sql.DB, mock sqlmock.Sqlmock) {
				if tc.mockSetup != nil {
					tc.mockSetup(mock)
				}

				repo := db.NewUsersRepository(dbConn)
				handler := RegisterUserHandler(repo)

				bodyBytes, err := json.Marshal(tc.request)
				if err != nil {
					t.Fatalf("failed to marshal request: %v", err)
				}

				req := routeit.NewTestRequest(t, "/auth/register", routeit.POST, routeit.TestRequestOptions{
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
