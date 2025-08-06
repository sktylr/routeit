package db

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/sktylr/routeit/examples/todo/auth"
	"github.com/sktylr/routeit/examples/todo/dao"
)

func TestCreateUser(t *testing.T) {
	t.Run("successfully creates user", func(t *testing.T) {
		WithTestConnection(t, func(db *sql.DB, mock sqlmock.Sqlmock) {
			repo := NewUsersRepository(db)
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

			user, err := repo.CreateUser(t.Context(), name, email, password)

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

			_, err := repo.CreateUser(t.Context(), "Bob", "bob@example.com", "MyPassword")

			if err == nil {
				t.Fatal("expected error but got nil")
			}
		})
	})

	t.Run("fails with ErrDuplicateKey on duplicate email", func(t *testing.T) {
		WithTestConnection(t, func(db *sql.DB, mock sqlmock.Sqlmock) {
			repo := NewUsersRepository(db)
			sqlErr := &mysql.MySQLError{
				Number:  1062,
				Message: "Duplicate entry 'alice@example.com' for key 'users.email'",
			}
			mock.ExpectExec(regexp.QuoteMeta(
				"INSERT INTO users (id, name, email, password, created, updated) VALUES (?, ?, ?, ?, ?, ?)",
			)).
				WithArgs(sqlmock.AnyArg(), "Alice", "alice@example.com", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
				WillReturnError(sqlErr)

			_, err := repo.CreateUser(t.Context(), "Alice", "alice@example.com", "SomePass")

			if err == nil {
				t.Fatal("expected duplicate key error, got nil")
			}
			if !errors.Is(err, ErrDuplicateKey) {
				t.Fatalf("expected error to be ErrDuplicateKey, got: %v", err)
			}
			if !strings.Contains(err.Error(), sqlErr.Message) {
				t.Errorf("wrapped error message missing original SQL error: %v", err)
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
		Created:  time.Now(),
		Updated:  time.Now(),
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

				user, found, err := repo.GetUserByEmail(t.Context(), tc.email)

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

func TestGetUserById(t *testing.T) {
	WithTestConnection(t, func(db *sql.DB, mock sqlmock.Sqlmock) {
		repo := NewUsersRepository(db)

		const userID = "12345678-1234-1234-1234-123456789abc"
		const email = "test@example.com"
		const name = "John Doe"
		const hashedPassword = "$2a$10$somehashedpw"
		const created = int64(1700000000)
		const updated = int64(1700000500)

		tests := []struct {
			name       string
			setupMock  func()
			inputId    string
			wantFound  bool
			wantErr    bool
			wantUserId string
		}{
			{
				name:    "user found",
				inputId: userID,
				setupMock: func() {
					mock.ExpectQuery(`SELECT id, name, email, password, created, updated FROM users WHERE id = \?`).
						WithArgs(userID).
						WillReturnRows(
							sqlmock.NewRows([]string{"id", "name", "email", "password", "created", "updated"}).
								AddRow(userID, name, email, hashedPassword, created, updated),
						)
				},
				wantFound:  true,
				wantErr:    false,
				wantUserId: userID,
			},
			{
				name:    "user not found",
				inputId: "nonexistent-id",
				setupMock: func() {
					mock.ExpectQuery(`SELECT id, name, email, password, created, updated FROM users WHERE id = \?`).
						WithArgs("nonexistent-id").
						WillReturnError(sql.ErrNoRows)
				},
				wantFound: false,
				wantErr:   false,
			},
			{
				name:    "db error",
				inputId: "bad-id",
				setupMock: func() {
					mock.ExpectQuery(`SELECT id, name, email, password, created, updated FROM users WHERE id = \?`).
						WithArgs("bad-id").
						WillReturnError(errors.New("db failure"))
				},
				wantFound: false,
				wantErr:   true,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				tc.setupMock()

				user, found, err := repo.GetUserById(t.Context(), tc.inputId)

				if (err != nil) != tc.wantErr {
					t.Fatalf("unexpected error: got %v, want error: %v", err, tc.wantErr)
				}

				if found != tc.wantFound {
					t.Fatalf("found = %v, want %v", found, tc.wantFound)
				}

				if found && user.Id != tc.wantUserId {
					t.Errorf("user.Id = %v, want %v", user.Id, tc.wantUserId)
				}

				if err := mock.ExpectationsWereMet(); err != nil {
					t.Errorf("there were unfulfilled expectations: %v", err)
				}
			})
		}
	})
}
