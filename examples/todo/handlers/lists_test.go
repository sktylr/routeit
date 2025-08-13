package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sktylr/routeit"
	"github.com/sktylr/routeit/examples/todo/dao"
	"github.com/sktylr/routeit/examples/todo/db"
)

func TestListsMultiHandler_Post(t *testing.T) {
	tests := []struct {
		name           string
		bodyFn         func(t *testing.T) []byte
		addUser        bool
		mockSetup      func(sqlmock.Sqlmock)
		expectedStatus routeit.HttpStatus
		assertBody     func(t *testing.T, res *routeit.TestResponse)
	}{
		{
			name: "valid create list returns list response",
			bodyFn: func(t *testing.T) []byte {
				req := CreateListRequest{
					Name:        "Groceries",
					Description: "Weekly shopping list",
				}
				bytes, err := json.Marshal(req)
				if err != nil {
					t.Fatalf("failed to marshal request: %v", err)
				}
				return bytes
			},
			addUser: true,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO lists (id, created, updated, user_id, name, description) VALUES (?, ?, ?, ?, ?, ?)`)).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "user-123", "Groceries", "Weekly shopping list").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: routeit.StatusCreated,
			assertBody: func(t *testing.T, res *routeit.TestResponse) {
				var body CreateListResponse
				res.BodyToJson(t, &body)
				if body.Name != "Groceries" || body.Description != "Weekly shopping list" {
					t.Errorf("unexpected response: %+v", body)
				}
			},
		},
		{
			name: "missing user ID returns unauthorized",
			bodyFn: func(t *testing.T) []byte {
				req := CreateListRequest{
					Name:        "No User",
					Description: "No auth header",
				}
				bytes, err := json.Marshal(req)
				if err != nil {
					t.Fatalf("failed to marshal request: %v", err)
				}
				return bytes
			},
			expectedStatus: routeit.StatusUnauthorized,
		},
		{
			name: "invalid JSON body returns 400",
			bodyFn: func(t *testing.T) []byte {
				return []byte("{invalid-json}")
			},
			addUser:        true,
			expectedStatus: routeit.StatusBadRequest,
		},
		{
			name: "empty name returns 400",
			bodyFn: func(t *testing.T) []byte {
				req := CreateListRequest{
					Name:        "",
					Description: "Some description",
				}
				bytes, err := json.Marshal(req)
				if err != nil {
					t.Fatalf("failed to marshal request: %v", err)
				}
				return bytes
			},
			addUser:        true,
			expectedStatus: routeit.StatusBadRequest,
		},
		{
			name: "db error returns 503",
			bodyFn: func(t *testing.T) []byte {
				req := CreateListRequest{
					Name:        "ErrList",
					Description: "DB will fail",
				}
				bytes, err := json.Marshal(req)
				if err != nil {
					t.Fatalf("failed to marshal request: %v", err)
				}
				return bytes
			},
			addUser: true,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO lists (id, created, updated, user_id, name, description) VALUES (?, ?, ?, ?, ?, ?)`)).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "user-123", "ErrList", "DB will fail").
					WillReturnError(errors.New("db insert failed"))
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
				repo := db.NewTodoListRepository(dbConn)
				handler := ListsMultiHandler(repo)
				req := routeit.NewTestRequest(t, "/lists", routeit.POST, routeit.TestRequestOptions{
					Body:    tc.bodyFn(t),
					Headers: []string{"Content-Type", "application/json"},
				})
				if tc.addUser {
					req.NewContextValue("user", &dao.User{Meta: dao.Meta{Id: "user-123"}})
				}

				res, err := routeit.TestHandler(handler, req)

				wantErr := tc.expectedStatus.Is4xx() || tc.expectedStatus.Is5xx()
				if err != nil {
					if !wantErr {
						t.Fatalf("unexpected error: %v", err)
					}
					httpErr, ok := err.(*routeit.HttpError)
					if !ok {
						t.Fatalf("expected HttpError, got %T", err)
					}
					if httpErr.Status() != tc.expectedStatus {
						t.Errorf("status = %v, want %v", httpErr.Status(), tc.expectedStatus)
					}
				} else {
					res.AssertStatusCode(t, tc.expectedStatus)
					tc.assertBody(t, res)
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Errorf("unmet SQL mock expectations: %v", err)
				}
			})
		})
	}
}
