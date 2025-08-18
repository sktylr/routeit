package middleware

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sktylr/routeit"
	"github.com/sktylr/routeit/examples/todo/dao"
	"github.com/sktylr/routeit/examples/todo/db"
)

func TestLoadItemMiddleware(t *testing.T) {
	type expect struct {
		proceed  bool
		wantErr  error
		itemId   string
		saveItem bool
	}
	now := time.Now().UTC()
	cases := []struct {
		name       string
		uri        string
		listId     string
		itemId     string
		userId     string
		setupMocks func(sqlmock.Sqlmock)
		expect     expect
	}{
		{
			name:   "item found and belongs to user and list",
			uri:    "/lists/list-123/items/item-123",
			listId: "list-123",
			itemId: "item-123",
			userId: "user-123",
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, created, updated, user_id, list_id, name, status FROM items WHERE id = \?`).
					WithArgs("item-123").
					WillReturnRows(sqlmock.NewRows([]string{"id", "created", "updated", "user_id", "list_id", "name", "status"}).
						AddRow("item-123", now, now, "user-123", "list-123", "buy milk", "PENDING"))
			},
			expect: expect{proceed: true, itemId: "item-123", saveItem: true},
		},
		{
			name:   "item not found returns 404",
			uri:    "/lists/list-123/items/item-404",
			listId: "list-123",
			itemId: "item-404",
			userId: "user-123",
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, created, updated, user_id, list_id, name, status FROM items WHERE id = \?`).
					WithArgs("item-404").
					WillReturnError(sql.ErrNoRows)
			},
			expect: expect{proceed: false, wantErr: db.ErrItemNotFound},
		},
		{
			name:   "db error propagated",
			uri:    "/lists/list-123/items/item-err",
			listId: "list-123",
			itemId: "item-err",
			userId: "user-123",
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, created, updated, user_id, list_id, name, status FROM items WHERE id = \?`).
					WithArgs("item-err").
					WillReturnError(errors.New("db down"))
			},
			expect: expect{proceed: false, wantErr: db.ErrDatabaseIssue},
		},
		{
			name:   "item belongs to another list returns 404",
			uri:    "/lists/list-123/items/item-123",
			listId: "list-123",
			itemId: "item-123",
			userId: "user-123",
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, created, updated, user_id, list_id, name, status FROM items WHERE id = \?`).
					WithArgs("item-123").
					WillReturnRows(sqlmock.NewRows([]string{"id", "created", "updated", "user_id", "list_id", "name", "status"}).
						AddRow("item-123", now, now, "user-123", "other-list", "buy milk", "COMPLETED"))
			},
			expect: expect{proceed: false, wantErr: routeit.ErrNotFound()},
		},
		{
			name:   "item belongs to another user returns 403",
			uri:    "/lists/list-123/items/item-123",
			listId: "list-123",
			itemId: "item-123",
			userId: "user-123",
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, created, updated, user_id, list_id, name, status FROM items WHERE id = \?`).
					WithArgs("item-123").
					WillReturnRows(sqlmock.NewRows([]string{"id", "created", "updated", "user_id", "list_id", "name", "status"}).
						AddRow("item-123", now, now, "other-user", "list-123", "buy milk", "PENDING"))
			},
			expect: expect{proceed: false, wantErr: routeit.ErrForbidden()},
		},
		{
			name:       "path not under lists",
			uri:        "/auth/register",
			setupMocks: func(s sqlmock.Sqlmock) {},
			expect:     expect{proceed: true, saveItem: false},
		},
		{
			name:       "path under lists but not items",
			uri:        "/lists/list-123",
			setupMocks: func(s sqlmock.Sqlmock) {},
			expect:     expect{proceed: true, saveItem: false},
		},
		{
			name:       "path under lists but not individual items",
			uri:        "/lists/list-123/items",
			setupMocks: func(s sqlmock.Sqlmock) {},
			expect:     expect{proceed: true, saveItem: false},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db.WithUnitTestConnection(t, func(sqlDB *sql.DB, mock sqlmock.Sqlmock) {
				tc.setupMocks(mock)

				repo := db.NewTodoItemRepository(sqlDB)
				mw := LoadItemMiddleware(repo)

				req := routeit.NewTestRequest(t, tc.uri, routeit.GET, routeit.TestRequestOptions{
					PathParams: map[string]string{
						"list": tc.listId,
						"item": tc.itemId,
					},
				})
				req.NewContextValue("user", &dao.User{Meta: dao.Meta{Id: tc.userId}})
				if tc.listId != "" {
					req.NewContextValue("list", &dao.TodoList{Meta: dao.Meta{Id: tc.listId}})
				}

				_, proceed, err := routeit.TestMiddleware(mw, req)

				if tc.expect.wantErr == nil {
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
					if !proceed {
						t.Errorf("expected middleware to proceed")
					}
					raw, ok := req.ContextValue("item")
					if ok != tc.expect.saveItem {
						t.Fatalf("expected middleware to save item to context")
					}
					got, ok := raw.(*dao.TodoItem)
					if tc.expect.saveItem && (!ok || got.Id != tc.expect.itemId) {
						t.Errorf("expected item id %s in context, got %+v", tc.expect.itemId, got)
					}
				} else {
					if err == nil {
						t.Fatal("expected error, got none")
					}
					if !errors.Is(err, tc.expect.wantErr) {
						t.Fatalf(`error = %v, wanted %v`, err, tc.expect.wantErr)
					}
				}
			})
		})
	}
}
