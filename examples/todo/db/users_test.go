package db

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sktylr/routeit/examples/todo/auth"
	"github.com/sktylr/routeit/examples/todo/dao"
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

func TestGetUserByEmail(t *testing.T) {
	testEmail := "test@example.com"
	expectedUser := dao.User{
		Id:       "some-uuid",
		Name:     "Test User",
		Email:    testEmail,
		Password: "hashed-pw",
		Created:  time.Now().Unix(),
		Updated:  time.Now().Unix(),
	}

	tests := []struct {
		name      string
		email     string
		mockRows  *sqlmock.Rows
		mockErr   error
		wantUser  *dao.User
		wantFound bool
		wantErr   bool
	}{
		{
			name:  "user found",
			email: testEmail,
			mockRows: sqlmock.NewRows([]string{
				"id", "name", "email", "password", "created", "updated",
			}).AddRow(
				expectedUser.Id,
				expectedUser.Name,
				expectedUser.Email,
				expectedUser.Password,
				expectedUser.Created,
				expectedUser.Updated,
			),
			wantUser:  &expectedUser,
			wantFound: true,
			wantErr:   false,
		},
		{
			name:      "user not found",
			email:     "missing@example.com",
			mockErr:   sql.ErrNoRows,
			wantUser:  nil,
			wantFound: false,
			wantErr:   false,
		},
		{
			name:      "db error",
			email:     "error@example.com",
			mockErr:   errors.New("connection lost"),
			wantUser:  nil,
			wantFound: false,
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			WithTestConnection(t, func(db *sql.DB, mock sqlmock.Sqlmock) {
				repo := NewUsersRepository(db)

				query := regexp.QuoteMeta(`
					SELECT id, name, email, password, created, updated
					FROM users
					WHERE email = ?
				`)

				if tc.mockRows != nil {
					mock.ExpectQuery(query).
						WithArgs(tc.email).
						WillReturnRows(tc.mockRows)
				} else if tc.mockErr != nil {
					mock.ExpectQuery(query).
						WithArgs(tc.email).
						WillReturnError(tc.mockErr)
				}

				user, found, err := repo.GetUserByEmail(context.Background(), tc.email)

				if (err != nil) != tc.wantErr {
					t.Fatalf("unexpected error: %v", err)
				}

				if found != tc.wantFound {
					t.Errorf("found = %v, want %v", found, tc.wantFound)
				}

				if !reflect.DeepEqual(user, tc.wantUser) {
					t.Errorf("user = %+v, want %+v", user, tc.wantUser)
				}

				if err := mock.ExpectationsWereMet(); err != nil {
					t.Errorf("unmet mock expectations: %v", err)
				}
			})
		})
	}
}
