package db

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sktylr/routeit/examples/todo/dao"
)

func TestUpdateStatus(t *testing.T) {
	type updateFunc func(ctx context.Context, id string) error
	tests := []struct {
		name      string
		update    func(repo *TodoItemRepository) updateFunc
		itemId    string
		wantQuery string
		wantArgs  []any
		mockSetup func(sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name:      "MarkAsCompleted - success",
			update:    func(repo *TodoItemRepository) updateFunc { return repo.MarkAsCompleted },
			itemId:    "123",
			wantQuery: `UPDATE items\s+SET status = 'COMPLETED',`,
			wantArgs:  []any{sqlmock.AnyArg(), "123"},
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`UPDATE items\s+SET status = 'COMPLETED',`).
					WithArgs(sqlmock.AnyArg(), "123").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:      "MarkAsPending - success",
			update:    func(repo *TodoItemRepository) updateFunc { return repo.MarkAsPending },
			itemId:    "456",
			wantQuery: `UPDATE items\s+SET status = 'PENDING',`,
			wantArgs:  []any{sqlmock.AnyArg(), "456"},
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`UPDATE items\s+SET status = 'PENDING',`).
					WithArgs(sqlmock.AnyArg(), "456").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:      "MarkAsCompleted - db error",
			update:    func(repo *TodoItemRepository) updateFunc { return repo.MarkAsCompleted },
			itemId:    "err-id",
			wantQuery: `UPDATE items\s+SET status = 'COMPLETED',`,
			wantArgs:  []any{sqlmock.AnyArg(), "err-id"},
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`UPDATE items\s+SET status = 'COMPLETED',`).
					WithArgs(sqlmock.AnyArg(), "err-id").
					WillReturnError(errors.New("db failure"))
			},
			wantErr: true,
		},
		{
			name:      "MarkAsPending - no rows affected",
			update:    func(repo *TodoItemRepository) updateFunc { return repo.MarkAsPending },
			itemId:    "no-match",
			wantQuery: `UPDATE items\s+SET status = 'PENDING',`,
			wantArgs:  []any{sqlmock.AnyArg(), "no-match"},
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`UPDATE items\s+SET status = 'PENDING',`).
					WithArgs(sqlmock.AnyArg(), "no-match").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			WithUnitTestConnection(t, func(sqlDB *sql.DB, mock sqlmock.Sqlmock) {
				tc.mockSetup(mock)
				repo := NewTodoItemRepository(sqlDB)

				err := tc.update(repo)(t.Context(), tc.itemId)

				if tc.wantErr && err == nil {
					t.Errorf("expected error, got nil")
				}
				if !tc.wantErr && err != nil {
					t.Errorf("did not expect error, got %v", err)
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Errorf("unmet sqlmock expectations: %v", err)
				}
			})
		})
	}
}

func TestGetById(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		mockSetup func(sqlmock.Sqlmock)
		wantItem  *dao.TodoItem
		wantErr   bool
	}{
		{
			name: "success",
			id:   "item-123",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT id, created, updated, user_id, list_id, name, status FROM items WHERE id = \?`).
					WithArgs("item-123").
					WillReturnRows(sqlmock.NewRows([]string{"id", "created", "updated", "user_id", "list_id", "name", "status"}).
						AddRow("item-123", time.Now(), time.Now(), "user-1", "list-1", "Test Item", "PENDING"))
			},
			wantItem: &dao.TodoItem{
				Meta:       dao.Meta{Id: "item-123"},
				UserId:     "user-1",
				TodoListId: "list-1",
				Name:       "Test Item",
				Status:     "PENDING",
			},
		},
		{
			name: "not found",
			id:   "missing-item",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT id, created, updated, user_id, list_id, name, status FROM items WHERE id = \?`).
					WithArgs("missing-item").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name: "db error",
			id:   "error-item",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT id, created, updated, user_id, list_id, name, status FROM items WHERE id = \?`).
					WithArgs("error-item").
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			WithUnitTestConnection(t, func(sqlDB *sql.DB, mock sqlmock.Sqlmock) {
				tc.mockSetup(mock)
				repo := NewTodoItemRepository(sqlDB)

				item, err := repo.GetById(t.Context(), tc.id)

				if tc.wantErr && err == nil {
					t.Errorf("expected error, got nil")
				}
				if !tc.wantErr && err != nil {
					t.Errorf("did not expect error, got %v", err)
				}
				if !tc.wantErr && item != nil {
					if item.Id != tc.wantItem.Id ||
						item.UserId != tc.wantItem.UserId ||
						item.TodoListId != tc.wantItem.TodoListId ||
						item.Name != tc.wantItem.Name ||
						item.Status != tc.wantItem.Status ||
						item.Created.IsZero() ||
						item.Updated.IsZero() {
						t.Errorf("unexpected item: %+v", item)
					}
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Errorf("unmet sqlmock expectations: %v", err)
				}
			})
		})
	}
}

func TestCreateItem(t *testing.T) {
	type args struct {
		userId string
		listId string
		name   string
	}

	tests := []struct {
		name      string
		args      args
		mockSetup func(sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name: "success",
			args: args{
				userId: "user-123",
				listId: "list-456",
				name:   "Test item",
			},
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`INSERT INTO items`).
					WithArgs(
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						"user-123",
						"list-456",
						"Test item",
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name: "db error",
			args: args{
				userId: "user-err",
				listId: "list-err",
				name:   "Fail item",
			},
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`INSERT INTO items`).
					WithArgs(
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						"user-err",
						"list-err",
						"Fail item",
					).
					WillReturnError(errors.New("insert failed"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			WithUnitTestConnection(t, func(sqlDB *sql.DB, mock sqlmock.Sqlmock) {
				tc.mockSetup(mock)
				repo := NewTodoItemRepository(sqlDB)

				item, err := repo.CreateItem(t.Context(), tc.args.userId, tc.args.listId, tc.args.name)

				if tc.wantErr && err == nil {
					t.Errorf("expected error, got nil")
				}
				if !tc.wantErr && err != nil {
					t.Errorf("did not expect error, got %v", err)
				}
				if !tc.wantErr {
					if item.UserId != tc.args.userId || item.TodoListId != tc.args.listId ||
						item.Name != tc.args.name || item.Status != "PENDING" {
						t.Errorf("unexpected item returned: %+v", item)
					}
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Errorf("unmet sqlmock expectations: %v", err)
				}
			})
		})
	}
}

func TestUpdateName(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		newName   string
		mockSetup func(sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name:    "success",
			id:      "item-123",
			newName: "Updated name",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`UPDATE items`).
					WithArgs("Updated name", sqlmock.AnyArg(), "item-123").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:    "db error",
			id:      "item-err",
			newName: "Bad update",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`UPDATE items`).
					WithArgs("Bad update", sqlmock.AnyArg(), "item-err").
					WillReturnError(errors.New("update failed"))
			},
			wantErr: true,
		},
		{
			name:    "no rows affected",
			id:      "item-123",
			newName: "No change",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`UPDATE items`).
					WithArgs("No change", sqlmock.AnyArg(), "item-123").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			WithUnitTestConnection(t, func(sqlDB *sql.DB, mock sqlmock.Sqlmock) {
				tc.mockSetup(mock)
				repo := NewTodoItemRepository(sqlDB)

				err := repo.UpdateName(t.Context(), tc.id, tc.newName)

				if tc.wantErr && err == nil {
					t.Errorf("expected error, got nil")
				}
				if !tc.wantErr && err != nil {
					t.Errorf("did not expect error, got %v", err)
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Errorf("unmet sqlmock expectations: %v", err)
				}
			})
		})
	}
}

func TestDeleteItem(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		mockSetup func(sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name: "success",
			id:   "item-123",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`DELETE FROM items`).
					WithArgs("item-123").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "db error",
			id:   "item-err",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`DELETE FROM items`).
					WithArgs("item-err").
					WillReturnError(errors.New("delete failed"))
			},
			wantErr: true,
		},
		{
			name: "no rows affected",
			id:   "item-missing",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`DELETE FROM items`).
					WithArgs("item-missing").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			WithUnitTestConnection(t, func(sqlDB *sql.DB, mock sqlmock.Sqlmock) {
				tc.mockSetup(mock)
				repo := NewTodoItemRepository(sqlDB)

				err := repo.DeleteItem(t.Context(), tc.id)

				if tc.wantErr && err == nil {
					t.Errorf("expected error, got nil")
				}
				if !tc.wantErr && err != nil {
					t.Errorf("did not expect error, got %v", err)
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Errorf("unmet sqlmock expectations: %v", err)
				}
			})
		})
	}
}

func TestGetByListAndUser(t *testing.T) {
	tests := []struct {
		name      string
		userId    string
		listId    string
		page      int
		pageSize  int
		mockSetup func(sqlmock.Sqlmock)
		wantItems []*dao.TodoItem
		wantErr   bool
	}{
		{
			name:     "page 1, pageSize 2 - success",
			userId:   "user-1",
			listId:   "list-1",
			page:     1,
			pageSize: 2,
			mockSetup: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "created", "updated", "user_id", "list_id", "name", "status",
				}).
					AddRow("item-1", time.Now(), time.Now(), "user-1", "list-1", "Task 1", "PENDING").
					AddRow("item-2", time.Now(), time.Now(), "user-1", "list-1", "Task 2", "COMPLETED")
				m.ExpectQuery(`SELECT id, created, updated, user_id, list_id, name, status FROM items`).
					WithArgs("user-1", "list-1", 2, 0).
					WillReturnRows(rows)
			},
			wantItems: []*dao.TodoItem{
				{Meta: dao.Meta{Id: "item-1"}, UserId: "user-1", TodoListId: "list-1", Name: "Task 1", Status: "PENDING"},
				{Meta: dao.Meta{Id: "item-2"}, UserId: "user-1", TodoListId: "list-1", Name: "Task 2", Status: "COMPLETED"},
			},
		},
		{
			name:     "page 2, pageSize 3 - success with offset",
			userId:   "user-2",
			listId:   "list-2",
			page:     2,
			pageSize: 3,
			mockSetup: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "created", "updated", "user_id", "list_id", "name", "status",
				}).
					AddRow("item-10", time.Now(), time.Now(), "user-2", "list-2", "Other task", "PENDING")

				m.ExpectQuery(`SELECT id, created, updated, user_id, list_id, name, status FROM items`).
					WithArgs("user-2", "list-2", 3, 3).
					WillReturnRows(rows)
			},
			wantItems: []*dao.TodoItem{
				{Meta: dao.Meta{Id: "item-10"}, UserId: "user-2", TodoListId: "list-2", Name: "Other task", Status: "PENDING"},
			},
		},
		{
			name:     "no items found",
			userId:   "user-3",
			listId:   "list-3",
			page:     1,
			pageSize: 5,
			mockSetup: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "created", "updated", "user_id", "list_id", "name", "status",
				})
				m.ExpectQuery(`SELECT id, created, updated, user_id, list_id, name, status FROM items`).
					WithArgs("user-3", "list-3", 5, 0).
					WillReturnRows(rows)
			},
			wantItems: []*dao.TodoItem{},
		},
		{
			name:     "db query error",
			userId:   "user-err",
			listId:   "list-err",
			page:     1,
			pageSize: 1,
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT id, created, updated, user_id, list_id, name, status FROM items`).
					WithArgs("user-err", "list-err", 1, 0).
					WillReturnError(errors.New("db failure"))
			},
			wantErr: true,
		},
		{
			name:     "row scan error",
			userId:   "user-4",
			listId:   "list-4",
			page:     1,
			pageSize: 1,
			mockSetup: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "created", "updated", "user_id", "list_id", "name", "status",
				}).
					AddRow(nil, time.Now(), time.Now(), "user-4", "list-4", "Bad Row", "PENDING")
				m.ExpectQuery(`SELECT id, created, updated, user_id, list_id, name, status FROM items`).
					WithArgs("user-4", "list-4", 1, 0).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			WithUnitTestConnection(t, func(sqlDB *sql.DB, mock sqlmock.Sqlmock) {
				tc.mockSetup(mock)
				repo := NewTodoItemRepository(sqlDB)

				items, err := repo.GetByListAndUser(t.Context(), tc.userId, tc.listId, tc.page, tc.pageSize)

				if tc.wantErr && err == nil {
					t.Errorf("expected error, got nil")
				}
				if !tc.wantErr && err != nil {
					t.Errorf("did not expect error, got %v", err)
				}
				if !tc.wantErr {
					if len(items) != len(tc.wantItems) {
						t.Fatalf("got %d items, want %d", len(items), len(tc.wantItems))
					}
					for i := range items {
						if items[i].Id != tc.wantItems[i].Id ||
							items[i].UserId != tc.wantItems[i].UserId ||
							items[i].TodoListId != tc.wantItems[i].TodoListId ||
							items[i].Name != tc.wantItems[i].Name ||
							items[i].Status != tc.wantItems[i].Status {
							t.Errorf("unexpected item at index %d: %+v, want %+v", i, items[i], tc.wantItems[i])
						}
					}
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Errorf("unmet sqlmock expectations: %v", err)
				}
			})
		})
	}
}

func TestCountByListAndUser(t *testing.T) {
	tests := []struct {
		name      string
		userId    string
		listId    string
		mockSetup func(sqlmock.Sqlmock)
		want      uint
		wantErr   bool
	}{
		{
			name:   "success",
			userId: "user-1",
			listId: "list-1",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT COUNT\(id\) FROM items WHERE user_id = \? AND list_id = \?`).
					WithArgs("user-1", "list-1").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(15))
			},
			want: 15,
		},
		{
			name:   "db query error",
			userId: "user-1",
			listId: "list-1",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT COUNT\(id\) FROM items WHERE user_id = \? AND list_id = \?`).
					WithArgs("user-1", "list-1").
					WillReturnError(errors.New("db failure"))
			},
			wantErr: true,
		},
		{
			name:   "row scan error",
			userId: "user-1",
			listId: "list-1",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT COUNT\(id\) FROM items WHERE user_id = \? AND list_id = \?`).
					WithArgs("user-1", "list-1").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow("foobar"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			WithUnitTestConnection(t, func(sqlDB *sql.DB, mock sqlmock.Sqlmock) {
				tc.mockSetup(mock)
				repo := NewTodoItemRepository(sqlDB)

				count, err := repo.CountByListAndUser(t.Context(), tc.userId, tc.listId)

				if tc.wantErr && err == nil {
					t.Errorf("expected error, got nil")
				}
				if !tc.wantErr && err != nil {
					t.Errorf("did not expect error, got %v", err)
				}
				if !tc.wantErr {
					if count != tc.want {
						t.Errorf(`count = %d, wanted %d`, count, tc.want)
					}
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Errorf("unmet sqlmock expectations: %v", err)
				}
			})
		})
	}
}
