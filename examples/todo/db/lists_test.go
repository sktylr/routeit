package db

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCreateList(t *testing.T) {
	tests := []struct {
		name      string
		userId    string
		listName  string
		desc      string
		mockSetup func(sqlmock.Sqlmock)
		expectErr bool
	}{
		{
			name:     "success",
			userId:   "user-123",
			listName: "Groceries",
			desc:     "Things to buy this week",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`INSERT INTO lists`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						"user-123", "Groceries", "Things to buy this week").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectErr: false,
		},
		{
			name:     "db error",
			userId:   "user-456",
			listName: "Chores",
			desc:     "Weekend cleaning tasks",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`INSERT INTO lists`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						"user-456", "Chores", "Weekend cleaning tasks").
					WillReturnError(errors.New("insert failed"))
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			WithUnitTestConnection(t, func(dbConn *sql.DB, mock sqlmock.Sqlmock) {
				tc.mockSetup(mock)
				repo := NewTodoListRepository(dbConn)

				list, err := repo.CreateList(t.Context(), tc.userId, tc.listName, tc.desc)

				if tc.expectErr && err == nil {
					t.Errorf("expected error, got nil")
				}
				if !tc.expectErr && err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !tc.expectErr {
					if list.UserId != tc.userId || list.Name != tc.listName || list.Description != tc.desc {
						t.Errorf("unexpected list: %+v", list)
					}
					if list.Created.IsZero() || list.Updated.IsZero() || list.Id == "" {
						t.Errorf("timestamps or id not set properly: %+v", list)
					}
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Errorf("unmet SQL expectations: %v", err)
				}
			})
		})
	}
}

func TestUpdateList(t *testing.T) {
	tests := []struct {
		name      string
		listId    string
		newName   string
		newDesc   string
		mockSetup func(sqlmock.Sqlmock)
		expectErr bool
		errCheck  func(error) bool
	}{
		{
			name:    "success",
			listId:  "list-123",
			newName: "Updated Groceries",
			newDesc: "Updated description",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`UPDATE lists`).
					WithArgs("Updated Groceries", "Updated description", sqlmock.AnyArg(), "list-123").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectErr: false,
		},
		{
			name:    "list not found",
			listId:  "missing-456",
			newName: "Doesn't matter",
			newDesc: "Nope",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`UPDATE lists`).
					WithArgs("Doesn't matter", "Nope", sqlmock.AnyArg(), "missing-456").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectErr: true,
			errCheck: func(err error) bool {
				var notFound *ErrListNotFound
				return errors.As(err, &notFound) && notFound.listId == "missing-456"
			},
		},
		{
			name:    "db error",
			listId:  "list-999",
			newName: "Broken",
			newDesc: "DB error",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`UPDATE lists`).
					WithArgs("Broken", "DB error", sqlmock.AnyArg(), "list-999").
					WillReturnError(errors.New("update failed"))
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			WithUnitTestConnection(t, func(dbConn *sql.DB, mock sqlmock.Sqlmock) {
				tc.mockSetup(mock)
				repo := NewTodoListRepository(dbConn)

				list, err := repo.UpdateList(t.Context(), tc.listId, tc.newName, tc.newDesc)

				if tc.expectErr {
					if err == nil {
						t.Errorf("expected error, got nil")
					} else if tc.errCheck != nil && !tc.errCheck(err) {
						t.Errorf("error check failed: %v", err)
					}
				} else {
					if err != nil {
						t.Errorf("unexpected error: %v", err)
					}
					if list.Name != tc.newName || list.Description != tc.newDesc {
						t.Errorf("unexpected list values: %+v", list)
					}
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Errorf("unmet SQL expectations: %v", err)
				}
			})
		})
	}
}
