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
