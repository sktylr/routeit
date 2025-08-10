package db

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMarkAsCompleted(t *testing.T) {
	tests := []struct {
		name      string
		itemId    string
		mockSetup func(sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name:   "successfully marks as completed",
			itemId: "123",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`UPDATE items`).
					WithArgs(sqlmock.AnyArg(), "123").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:   "no rows affected",
			itemId: "notfound",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`UPDATE items`).
					WithArgs(sqlmock.AnyArg(), "notfound").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: false,
		},
		{
			name:   "database error",
			itemId: "err-id",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`UPDATE items`).
					WithArgs(sqlmock.AnyArg(), "err-id").
					WillReturnError(errors.New("db failure"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			WithUnitTestConnection(t, func(sqlDB *sql.DB, mock sqlmock.Sqlmock) {
				tc.mockSetup(mock)
				repo := NewTodoItemRepository(sqlDB)

				err := repo.MarkAsCompleted(t.Context(), tc.itemId)

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
