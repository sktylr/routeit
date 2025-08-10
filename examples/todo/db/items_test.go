package db

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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
			wantErr: false,
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
			wantErr: false,
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
			wantErr: false,
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
			wantErr: false,
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
			wantErr: false,
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
			id:      "item-missing",
			newName: "No change",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`UPDATE items`).
					WithArgs("No change", sqlmock.AnyArg(), "item-missing").
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
