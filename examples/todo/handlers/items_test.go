package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sktylr/routeit"
	"github.com/sktylr/routeit/examples/todo/dao"
	"github.com/sktylr/routeit/examples/todo/db"
)

func TestItemsMultiHandler(t *testing.T) {
	t.Run("POST", func(t *testing.T) {
		tests := []struct {
			name       string
			body       []byte
			mockSetup  func(mock sqlmock.Sqlmock)
			wantErr    error
			assertBody func(t *testing.T, res *routeit.TestResponse)
		}{
			{
				name: "valid create item",
				body: func() []byte {
					bytes, _ := json.Marshal(CreateItemRequest{Name: "Milk"})
					return bytes
				}(),
				mockSetup: func(mock sqlmock.Sqlmock) {
					mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO items (id, created, updated, user_id, list_id, name, status) VALUES (?, ?, ?, ?, ?, ?, 'PENDING')`)).
						WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "user-123", "list-123", "Milk").
						WillReturnResult(sqlmock.NewResult(1, 1))
				},
				assertBody: func(t *testing.T, res *routeit.TestResponse) {
					var body CreateItemResponse
					res.BodyToJson(t, &body)
					if body.Name != "Milk" {
						t.Errorf("expected name Milk, got %q", body.Name)
					}
				},
			},
			{
				name:    "invalid JSON returns 400",
				body:    []byte("{invalid-json}"),
				wantErr: routeit.ErrBadRequest(),
			},
			{
				name: "db error propagated",
				body: func() []byte {
					bytes, _ := json.Marshal(CreateItemRequest{Name: "Eggs"})
					return bytes
				}(),
				mockSetup: func(mock sqlmock.Sqlmock) {
					mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO items (id, created, updated, user_id, list_id, name, status) VALUES (?, ?, ?, ?, ?, ?, 'PENDING')`)).
						WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "user-123", "list-123", "Eggs").
						WillReturnError(errors.New("db insert failed"))
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
					repo := db.NewTodoItemRepository(dbConn)
					handler := ItemsMultiHandler(repo)
					req := routeit.NewTestRequest(t,
						"/lists/list-123/items",
						routeit.POST,
						routeit.TestRequestOptions{
							PathParams: map[string]string{"list": "list-123"},
							Body:       tc.body,
							Headers:    []string{"Content-Type", "application/json"},
						},
					)
					req.NewContextValue("user", &dao.User{Meta: dao.Meta{Id: "user-123"}})

					res, err := routeit.TestHandler(handler, req)

					if err != nil {
						if tc.wantErr == nil {
							t.Fatalf("unexpected error: %v", err)
						}
						if !errors.Is(err, tc.wantErr) {
							t.Fatalf("err = %v, wanted %v", err, tc.wantErr)
						}
					} else {
						res.AssertStatusCode(t, routeit.StatusCreated)
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

	t.Run("GET", func(t *testing.T) {
		db.WithUnitTestConnection(t, func(dbConn *sql.DB, mock sqlmock.Sqlmock) {
			created := time.Now()
			updated := created
			mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, created, updated, user_id, list_id, name, status FROM items WHERE user_id = ? AND list_id = ? ORDER BY created DESC LIMIT ? OFFSET ?`)).
				WithArgs("user-123", "list-123", 10, 0).
				WillReturnRows(sqlmock.NewRows([]string{"id", "created", "updated", "user_id", "list_id", "name", "status"}).
					AddRow("item-1", created, updated, "user-123", "list-123", "Milk", "PENDING"))

			repo := db.NewTodoItemRepository(dbConn)
			handler := ItemsMultiHandler(repo)
			req := routeit.NewTestRequest(t, "/lists/list-123/items", routeit.GET, routeit.TestRequestOptions{
				PathParams: map[string]string{"list": "list-123"},
			})
			req.NewContextValue("user", &dao.User{Meta: dao.Meta{Id: "user-123"}})

			res, err := routeit.TestHandler(handler, req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			res.AssertStatusCode(t, routeit.StatusOK)
			var body ListItemsResponse
			res.BodyToJson(t, &body)
			if len(body.Items) != 1 || body.Items[0].Name != "Milk" {
				t.Errorf("unexpected body: %+v", body)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet SQL mock expectations: %v", err)
			}
		})
	})
}

func TestItemsIndividualHandler(t *testing.T) {
	t.Run("GET", func(t *testing.T) {
		handler := ItemsIndividualHandler(nil)
		req := routeit.NewTestRequest(t, "/lists/list-123/items/item-1", routeit.GET, routeit.TestRequestOptions{
			PathParams: map[string]string{"list": "list-123", "item": "item-1"},
		})
		now := time.Now()
		req.NewContextValue("item", &dao.TodoItem{
			Meta:   dao.Meta{Id: "item-1", Created: now, Updated: now},
			Name:   "Milk",
			Status: "PENDING",
		})

		res, err := routeit.TestHandler(handler, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		res.AssertStatusCode(t, routeit.StatusOK)
		var body GetItemResponse
		res.BodyToJson(t, &body)
		if body.Id != "item-1" || body.Name != "Milk" {
			t.Errorf("unexpected body: %+v", body)
		}
	})

	t.Run("PUT", func(t *testing.T) {
		tests := []struct {
			name    string
			body    []byte
			mock    func(sqlmock.Sqlmock)
			wantErr error
		}{
			{
				name: "successfully updates item name",
				body: []byte(`{"name":"New Milk"}`),
				mock: func(mock sqlmock.Sqlmock) {
					mock.ExpectExec(regexp.QuoteMeta(`UPDATE items SET name = ?, updated = ? WHERE id = ?`)).
						WithArgs("New Milk", sqlmock.AnyArg(), "item-1").
						WillReturnResult(sqlmock.NewResult(0, 1))
				},
			},
			{
				name:    "invalid JSON",
				body:    []byte(`{"invalid-json"}`),
				mock:    func(mock sqlmock.Sqlmock) {},
				wantErr: routeit.ErrBadRequest(),
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				db.WithUnitTestConnection(t, func(dbConn *sql.DB, mock sqlmock.Sqlmock) {
					repo := db.NewTodoItemRepository(dbConn)
					if tc.mock != nil {
						tc.mock(mock)
					}
					handler := ItemsIndividualHandler(repo)
					req := routeit.NewTestRequest(t,
						"/lists/list-123/items/item-1",
						routeit.PUT,
						routeit.TestRequestOptions{
							PathParams: map[string]string{"list": "list-123", "item": "item-1"},
							Body:       tc.body,
							Headers:    []string{"Content-Type", "application/json"},
						},
					)

					res, err := routeit.TestHandler(handler, req)
					if err != nil {
						if tc.wantErr == nil {
							t.Fatalf("unexpected error: %v", err)
						}
						if !errors.Is(err, tc.wantErr) {
							t.Fatalf("err = %v, wanted %v", err, tc.wantErr)
						}
					} else {
						res.AssertStatusCode(t, routeit.StatusOK)
					}
					if err := mock.ExpectationsWereMet(); err != nil {
						t.Errorf("unmet SQL mock expectations: %v", err)
					}
				})
			})
		}
	})

	t.Run("PATCH", func(t *testing.T) {
		tests := []struct {
			name    string
			body    []byte
			mock    func(sqlmock.Sqlmock)
			wantErr error
		}{
			{
				name: "mark as completed",
				body: []byte(`{"status":"COMPLETED"}`),
				mock: func(mock sqlmock.Sqlmock) {
					mock.ExpectExec(regexp.QuoteMeta(`UPDATE items SET status = 'COMPLETED', updated = ? WHERE id = ?`)).
						WithArgs(sqlmock.AnyArg(), "item-1").
						WillReturnResult(sqlmock.NewResult(0, 1))
				},
			},
			{
				name:    "invalid status",
				body:    []byte(`{"status":"UNKNOWN"}`),
				mock:    func(mock sqlmock.Sqlmock) {},
				wantErr: routeit.ErrBadRequest(),
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				db.WithUnitTestConnection(t, func(dbConn *sql.DB, mock sqlmock.Sqlmock) {
					repo := db.NewTodoItemRepository(dbConn)
					if tc.mock != nil {
						tc.mock(mock)
					}
					handler := ItemsIndividualHandler(repo)
					req := routeit.NewTestRequest(t,
						"/lists/list-123/items/item-1",
						routeit.PATCH,
						routeit.TestRequestOptions{
							PathParams: map[string]string{"list": "list-123", "item": "item-1"},
							Body:       tc.body,
							Headers:    []string{"Content-Type", "application/json"},
						},
					)

					_, err := routeit.TestHandler(handler, req)
					if err != nil {
						if tc.wantErr == nil {
							t.Fatalf("unexpected error: %v", err)
						}
						if !errors.Is(err, tc.wantErr) {
							t.Fatalf("err = %v, wanted %v", err, tc.wantErr)
						}
					}
					if err := mock.ExpectationsWereMet(); err != nil {
						t.Errorf("unmet SQL mock expectations: %v", err)
					}
				})
			})
		}
	})

	t.Run("DELETE", func(t *testing.T) {
		tests := []struct {
			name    string
			mock    func(sqlmock.Sqlmock)
			wantErr error
		}{
			{
				name: "successfully deletes item",
				mock: func(mock sqlmock.Sqlmock) {
					mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM items WHERE id = ?`)).
						WithArgs("item-1").
						WillReturnResult(sqlmock.NewResult(0, 1))
				},
			},
			{
				name: "delete not found",
				mock: func(mock sqlmock.Sqlmock) {
					mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM items WHERE id = ?`)).
						WithArgs("item-1").
						WillReturnResult(sqlmock.NewResult(0, 0))
				},
				wantErr: db.ErrItemNotFound,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				db.WithUnitTestConnection(t, func(dbConn *sql.DB, mock sqlmock.Sqlmock) {
					repo := db.NewTodoItemRepository(dbConn)
					if tc.mock != nil {
						tc.mock(mock)
					}
					handler := ItemsIndividualHandler(repo)
					req := routeit.NewTestRequest(t,
						"/lists/list-123/items/item-1",
						routeit.DELETE,
						routeit.TestRequestOptions{
							PathParams: map[string]string{"list": "list-123", "item": "item-1"},
						},
					)

					_, err := routeit.TestHandler(handler, req)
					if err != nil {
						if tc.wantErr == nil {
							t.Fatalf("unexpected error: %v", err)
						}
						if !errors.Is(err, tc.wantErr) {
							t.Fatalf("err = %v, wanted %v", err, tc.wantErr)
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
