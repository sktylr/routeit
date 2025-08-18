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

func TestLoadListMiddleware(t *testing.T) {
	type expect struct {
		proceed  bool
		wantErr  error
		listId   string
		saveList bool
	}
	now := time.Now().UTC()
	cases := []struct {
		name       string
		uri        string
		listId     string
		userId     string
		setupMocks func(sqlmock.Sqlmock)
		expect     expect
	}{
		{
			name:   "list found and belongs to user",
			uri:    "/lists/list-123",
			listId: "list-123",
			userId: "user-123",
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, created, updated, user_id, name, description FROM lists WHERE id = \?`).
					WithArgs("list-123").
					WillReturnRows(sqlmock.NewRows([]string{"id", "created", "updated", "user_id", "name", "description"}).
						AddRow("list-123", now, now, "user-123", "Groceries", "Buy stuff"))
			},
			expect: expect{proceed: true, listId: "list-123", saveList: true},
		},
		{
			name:   "list not found returns 404",
			uri:    "/lists/list-404",
			listId: "list-404",
			userId: "user-123",
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, created, updated, user_id, name, description FROM lists WHERE id = \?`).
					WithArgs("list-404").
					WillReturnError(sql.ErrNoRows)
			},
			expect: expect{proceed: false, wantErr: db.ErrListNotFound},
		},
		{
			name:   "list belongs to another user returns 403",
			uri:    "/lists/list-123",
			listId: "list-123",
			userId: "user-123",
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, created, updated, user_id, name, description FROM lists WHERE id = \?`).
					WithArgs("list-123").
					WillReturnRows(sqlmock.NewRows([]string{"id", "created", "updated", "user_id", "name", "description"}).
						AddRow("list-123", now, now, "other-user", "Groceries", "Buy stuff"))
			},
			expect: expect{proceed: false, wantErr: routeit.ErrForbidden()},
		},
		{
			name:   "db error propagated",
			uri:    "/lists/list-err",
			listId: "list-err",
			userId: "user-123",
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, created, updated, user_id, name, description FROM lists WHERE id = \?`).
					WithArgs("list-err").
					WillReturnError(errors.New("db down"))
			},
			expect: expect{proceed: false, wantErr: db.ErrDatabaseIssue},
		},
		{
			name:       "path not for lists",
			uri:        "/auth/register",
			setupMocks: func(s sqlmock.Sqlmock) {},
			expect:     expect{proceed: true, saveList: false},
		},
		{
			name:       "path for lists, but not individual lists",
			uri:        "/lists",
			setupMocks: func(s sqlmock.Sqlmock) {},
			expect:     expect{proceed: true, saveList: false},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db.WithUnitTestConnection(t, func(sqlDB *sql.DB, mock sqlmock.Sqlmock) {
				tc.setupMocks(mock)
				repo := db.NewTodoListRepository(sqlDB)
				mw := LoadListMiddleware(repo)
				req := routeit.NewTestRequest(t, tc.uri, routeit.GET, routeit.TestRequestOptions{
					PathParams: map[string]string{"list": tc.listId},
				})
				req.NewContextValue("user", &dao.User{Meta: dao.Meta{Id: tc.userId}})

				_, proceed, err := routeit.TestMiddleware(mw, req)

				if tc.expect.wantErr == nil {
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
					if !proceed {
						t.Errorf("expected middleware to proceed")
					}
					raw, ok := req.ContextValue("list")
					if ok != tc.expect.saveList {
						t.Fatalf("expected middleware to save list to context")
					}
					got, ok := raw.(*dao.TodoList)
					if tc.expect.saveList && (!ok || got.Id != tc.expect.listId) {
						t.Errorf("expected list id %s in context, got %+v", tc.expect.listId, got)
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
