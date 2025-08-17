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
	"github.com/sktylr/routeit/examples/todo/dao"
	"github.com/sktylr/routeit/examples/todo/db"
)

func TestListsMultiHandler(t *testing.T) {
	t.Run("GET", func(t *testing.T) {
		tests := []struct {
			name           string
			query          string
			addUser        bool
			mockSetup      func(sqlmock.Sqlmock)
			expectedStatus routeit.HttpStatus
			assertBody     func(t *testing.T, res *routeit.TestResponse)
		}{
			{
				name:    "defaults page=1, page_size=10 when not provided",
				addUser: true,
				mockSetup: func(mock sqlmock.Sqlmock) {
					created := time.Now().Add(-time.Hour)
					updated := created.Add(time.Minute)
					mock.ExpectQuery(regexp.QuoteMeta(`
					SELECT l.id, l.created, l.updated, l.user_id, l.name, l.description,
						COUNT(i.id) AS total_items,
						SUM(CASE WHEN i.status = 'COMPLETED' THEN 1 ELSE 0 END) AS completed_items
					FROM lists l
					LEFT JOIN items i ON l.id = i.list_id
					WHERE l.user_id = ?
					GROUP BY l.id
					ORDER BY l.created DESC
					LIMIT ? OFFSET ?
				`)).WithArgs("user-123", 10, 0).WillReturnRows(sqlmock.NewRows([]string{
						"id", "created", "updated", "l.user_id", "name", "description", "total_items", "completed_items",
					}).AddRow("list-1", created, updated, "user-123", "Groceries", "Weekly shopping", 3, 1))
					mock.ExpectQuery(regexp.QuoteMeta(`
					SELECT id, created, updated, user_id, list_id, name, status
					FROM items
					WHERE list_id IN (?)
					ORDER BY list_id, created ASC
				`)).WithArgs("list-1").WillReturnRows(sqlmock.NewRows([]string{
						"id", "created", "updated", "user_id", "list_id", "name", "status",
					}).AddRow("item-1", created, updated, "user-123", "list-1", "Milk", "COMPLETED").
						AddRow("item-2", created, updated, "user-123", "list-1", "Bread", "PENDING").
						AddRow("item-3", created, updated, "user-123", "list-1", "Bananas", "PENDING"))
				},
				expectedStatus: routeit.StatusOK,
				assertBody: func(t *testing.T, res *routeit.TestResponse) {
					var body ListListsResponse
					res.BodyToJson(t, &body)
					if len(body.Lists) != 1 {
						t.Fatalf("expected 1 list, got %d", len(body.Lists))
					}
					if body.Lists[0].Name != "Groceries" {
						t.Errorf("expected name 'Groceries', got %q", body.Lists[0].Name)
					}
					if len(body.Lists[0].Items) != 3 {
						t.Errorf(`# items = %d, expected 3`, len(body.Lists[0].Items))
					}
				},
			},
			{
				name:           "invalid page returns 400",
				query:          "?page=abc",
				addUser:        true,
				expectedStatus: routeit.StatusBadRequest,
			},
			{
				name:           "invalid page_size returns 400",
				query:          "?page_size=-5",
				addUser:        true,
				expectedStatus: routeit.StatusBadRequest,
			},
			{
				name:           "missing user header returns 401",
				expectedStatus: routeit.StatusUnauthorized,
			},
			{
				name:    "custom page and size applied",
				query:   "?page=2&page_size=5",
				addUser: true,
				mockSetup: func(mock sqlmock.Sqlmock) {
					created := time.Now()
					updated := created
					mock.ExpectQuery(regexp.QuoteMeta(`
					SELECT l.id, l.created, l.updated, l.user_id, l.name, l.description,
						COUNT(i.id) AS total_items,
						SUM(CASE WHEN i.status = 'COMPLETED' THEN 1 ELSE 0 END) AS completed_items
					FROM lists l
					LEFT JOIN items i ON l.id = i.list_id
					WHERE l.user_id = ?
					GROUP BY l.id
					ORDER BY l.created DESC
					LIMIT ? OFFSET ?
				`)).WithArgs("user-123", 5, 5).WillReturnRows(sqlmock.NewRows([]string{
						"id", "created", "updated", "l.user_id", "name", "description", "total_items", "completed_items",
					}).AddRow("list-2", created, updated, "user-123", "Work", "Tasks", 2, 0))
					mock.ExpectQuery(regexp.QuoteMeta(`
					SELECT id, created, updated, user_id, list_id, name, status
					FROM items
					WHERE list_id IN (?)
					ORDER BY list_id, created ASC
				`)).WithArgs("list-2").WillReturnRows(sqlmock.NewRows([]string{
						"id", "created", "updated", "user_id", "list_id", "name", "status",
					}).AddRow("item-1", created, updated, "user-123", "list-1", "Send email", "PENDING").
						AddRow("item-2", created, updated, "user-123", "list-1", "Plan meeting", "PENDING"))
				},
				expectedStatus: routeit.StatusOK,
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
					req := routeit.NewTestRequest(t, "/lists"+tc.query, routeit.GET, routeit.TestRequestOptions{})
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
						if tc.assertBody != nil {
							tc.assertBody(t, res)
						}
					}
					if err := mock.ExpectationsWereMet(); err != nil {
						t.Errorf("unmet SQL mock expectations: %v", err)
					}
				})
			})
		}
	})

	t.Run("POST", func(t *testing.T) {
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
	})
}

func TestListsIndividualHandler(t *testing.T) {
	t.Run("GET", func(t *testing.T) {
		t.Run("uses list from context", func(t *testing.T) {
			handler := ListsIndividualHandler(nil)
			req := routeit.NewTestRequest(t, "/lists/list-123", routeit.GET, routeit.TestRequestOptions{
				PathParams: map[string]string{"list": "list-123"},
			})
			now := time.Now().Truncate(time.Minute)
			req.NewContextValue("list", &dao.TodoList{
				Meta: dao.Meta{
					Id:      "list-123",
					Created: now,
					Updated: now,
				},
				Name:        "foobar",
				Description: "baz",
			})

			res, err := routeit.TestHandler(handler, req)

			if err != nil {
				t.Fatalf(`err = %+v, wanted nil`, err)
			}
			res.AssertStatusCode(t, routeit.StatusOK)

			var body GetListResponse
			res.BodyToJson(t, &body)

			if body.Id != "list-123" {
				t.Errorf(`body.Id = %s, wanted "list-123"`, body.Id)
			}
			if body.Created != now {
				t.Errorf(`body.Created = %+v, wanted %+v`, body.Created, now)
			}
			if body.Updated != now {
				t.Errorf(`body.Updated = %+v, wanted %+v`, body.Updated, now)
			}
			if body.Name != "foobar" {
				t.Errorf(`body.Name = %s, wanted "foobar"`, body.Name)
			}
			if body.Description != "baz" {
				t.Errorf(`body.Description = %s, wanted "baz"`, body.Description)
			}
		})

		t.Run("panics whenever list is not set (unexpected)", func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Error("expected panic")
				}
			}()

			req := routeit.NewTestRequest(t, "/lists/list-123", routeit.GET, routeit.TestRequestOptions{
				PathParams: map[string]string{"list": "list-123"},
			})
			routeit.TestHandler(ListsIndividualHandler(nil), req)
		})
	})

	t.Run("DELETE", func(t *testing.T) {
		tests := []struct {
			name           string
			listId         string
			mockSetup      func(mock sqlmock.Sqlmock)
			expectedStatus routeit.HttpStatus
		}{
			{
				name:   "successfully deletes list",
				listId: "list-123",
				mockSetup: func(mock sqlmock.Sqlmock) {
					mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM lists WHERE id = ?`)).
						WithArgs("list-123").
						WillReturnResult(sqlmock.NewResult(0, 1))
				},
				expectedStatus: routeit.StatusNoContent,
			},
			{
				name:   "delete returns not found if no rows affected",
				listId: "list-123",
				mockSetup: func(mock sqlmock.Sqlmock) {
					mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM lists WHERE id = ?`)).
						WithArgs("list-123").
						WillReturnResult(sqlmock.NewResult(0, 0))
				},
				expectedStatus: routeit.StatusNotFound,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				db.WithUnitTestConnection(t, func(dbConn *sql.DB, mock sqlmock.Sqlmock) {
					tc.mockSetup(mock)
					repo := db.NewTodoListRepository(dbConn)
					handler := ListsIndividualHandler(repo)
					req := routeit.NewTestRequest(t,
						fmt.Sprintf("/lists/%s", tc.listId),
						routeit.DELETE,
						routeit.TestRequestOptions{
							PathParams: map[string]string{"list": tc.listId},
						},
					)

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
					}
					if err := mock.ExpectationsWereMet(); err != nil {
						t.Errorf("unmet SQL mock expectations: %v", err)
					}
				})
			})
		}
	})

	t.Run("PUT", func(t *testing.T) {
		tests := []struct {
			name           string
			listId         string
			addUser        bool
			body           []byte
			mockSetup      func(mock sqlmock.Sqlmock, created, updated time.Time)
			expectedStatus routeit.HttpStatus
			assertBody     func(t *testing.T, res *routeit.TestResponse, created, updated time.Time)
		}{
			{
				name:    "successfully updates list",
				listId:  "list-123",
				addUser: true,
				body:    []byte(`{"name":"Updated Groceries","description":"Bi-weekly shopping"}`),
				mockSetup: func(mock sqlmock.Sqlmock, created, updated time.Time) {
					mock.ExpectQuery(regexp.QuoteMeta(`
						SELECT id, created, updated, user_id, name, description
						FROM lists
						WHERE id = ?
					`)).
						WithArgs("list-123").
						WillReturnRows(sqlmock.NewRows([]string{
							"id", "created", "updated", "user_id", "name", "description",
						}).AddRow("list-123", created, updated, "user-123", "Groceries", "Weekly shopping"))
					mock.ExpectExec(regexp.QuoteMeta(`
						UPDATE lists SET name = ?, description = ?, updated = ? WHERE id = ?
					`)).
						WithArgs("Updated Groceries", "Bi-weekly shopping", sqlmock.AnyArg(), "list-123").
						WillReturnResult(sqlmock.NewResult(0, 1))
				},
				expectedStatus: routeit.StatusOK,
				assertBody: func(t *testing.T, res *routeit.TestResponse, created, updated time.Time) {
					var body UpdateListResponse
					res.BodyToJson(t, &body)
					if body.Name != "Updated Groceries" {
						t.Errorf("expected name 'Updated Groceries', got %q", body.Name)
					}
					if body.Description != "Bi-weekly shopping" {
						t.Errorf("expected description 'Bi-weekly shopping', got %q", body.Description)
					}
				},
			},
			{
				name:           "missing user header returns 401",
				listId:         "list-123",
				addUser:        false,
				body:           []byte(`{"name":"Foo"}`),
				expectedStatus: routeit.StatusUnauthorized,
			},
			{
				name:    "list not found returns 404",
				listId:  "missing-list",
				addUser: true,
				body:    []byte(`{"name":"Something","description":"Whatever"}`),
				mockSetup: func(mock sqlmock.Sqlmock, created, updated time.Time) {
					mock.ExpectQuery(regexp.QuoteMeta(`
						SELECT id, created, updated, user_id, name, description
						FROM lists
						WHERE id = ?
					`)).
						WithArgs("missing-list").
						WillReturnError(sql.ErrNoRows)
				},
				expectedStatus: routeit.StatusNotFound,
			},
			{
				name:    "list belongs to another user returns 403",
				listId:  "list-xyz",
				addUser: true,
				body:    []byte(`{"name":"Hack attempt"}`),
				mockSetup: func(mock sqlmock.Sqlmock, created, updated time.Time) {
					mock.ExpectQuery(regexp.QuoteMeta(`
						SELECT id, created, updated, user_id, name, description
						FROM lists
						WHERE id = ?
					`)).
						WithArgs("list-xyz").
						WillReturnRows(sqlmock.NewRows([]string{
							"id", "created", "updated", "user_id", "name", "description",
						}).AddRow("list-xyz", created, updated, "different-user", "Other list", "Not yours"))
				},
				expectedStatus: routeit.StatusForbidden,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				db.WithUnitTestConnection(t, func(dbConn *sql.DB, mock sqlmock.Sqlmock) {
					created := time.Now().Add(-time.Hour)
					updated := created.Add(30 * time.Minute)
					if tc.mockSetup != nil {
						tc.mockSetup(mock, created, updated)
					}
					repo := db.NewTodoListRepository(dbConn)
					handler := ListsIndividualHandler(repo)

					req := routeit.NewTestRequest(t,
						fmt.Sprintf("/lists/%s", tc.listId),
						routeit.PUT,
						routeit.TestRequestOptions{
							PathParams: map[string]string{"list": tc.listId},
							Body:       tc.body,
							Headers:    []string{"Content-Type", "application/json"},
						},
					)

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
						if tc.assertBody != nil {
							tc.assertBody(t, res, created, updated)
						}
					}
					if err := mock.ExpectationsWereMet(); err != nil {
						t.Errorf("unmet SQL mock expectations: %v", err)
					}
				})
			})
		}
	})
}
