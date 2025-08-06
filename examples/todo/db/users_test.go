package db

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sktylr/routeit/examples/todo/auth"
)

func TestCreateUser(t *testing.T) {
	t.Run("successfully creates user", func(t *testing.T) {
		WithTestConnection(t, func(db *sql.DB, mock sqlmock.Sqlmock) {
			repo := NewUsersRepository(db)

			ctx := context.Background()
			name := "Alice"
			email := "alice@example.com"
			password := "SecurePass123"

			mock.ExpectExec(regexp.QuoteMeta(
				"INSERT INTO users (id, name, email, password, created, updated) VALUES (?, ?, ?, ?, ?, ?)",
			)).
				WithArgs(
					sqlmock.AnyArg(),
					name,
					email,
					sqlmock.AnyArg(),
					sqlmock.AnyArg(),
					sqlmock.AnyArg(),
				).
				WillReturnResult(sqlmock.NewResult(1, 1))

			user, err := repo.CreateUser(ctx, name, email, password)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if user.Name != name || user.Email != email {
				t.Errorf("unexpected user returned: %+v", user)
			}

			if user.Id == "" {
				t.Errorf("expected non-empty user ID")
			}
			if ok := auth.ComparePassword(user.Password, password); !ok {
				t.Errorf("password hash did not match original")
			}
		})
	})

	t.Run("fails when insert fails", func(t *testing.T) {
		WithTestConnection(t, func(db *sql.DB, mock sqlmock.Sqlmock) {
			repo := NewUsersRepository(db)

			mock.ExpectExec(regexp.QuoteMeta(
				"INSERT INTO users (id, name, email, password, created, updated) VALUES (?, ?, ?, ?, ?, ?)",
			)).
				WithArgs(sqlmock.AnyArg(), "Bob", "bob@example.com", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
				WillReturnError(sqlmock.ErrCancelled)

			_, err := repo.CreateUser(context.Background(), "Bob", "bob@example.com", "MyPassword")
			if err == nil {
				t.Fatal("expected error but got nil")
			}
		})
	})
}
