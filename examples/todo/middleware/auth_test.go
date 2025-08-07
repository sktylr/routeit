package middleware

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sktylr/routeit"
	"github.com/sktylr/routeit/examples/todo/auth"
	"github.com/sktylr/routeit/examples/todo/dao"
	"github.com/sktylr/routeit/examples/todo/db"
)

func TestJwtMiddleware(t *testing.T) {
	user := &dao.User{
		Meta:  dao.Meta{Id: "123"},
		Name:  "Test User",
		Email: "test@example.com",
	}

	now := time.Now()

	validClaims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.Id,
			Issuer:    "todo-sample-app",
		},
		Type: "access",
	}
	expiredClaims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			Subject:   user.Id,
			Issuer:    "todo-sample-app",
		},
		Type: "access",
	}
	invalidClaims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.Id,
			Issuer:    "todo-sample-app",
		},
		Type: "refresh",
	}

	gen := func(cl auth.Claims) string {
		token := jwt.NewWithClaims(jwt.SigningMethodHS384, cl)
		s, err := token.SignedString([]byte("super-secret-key"))
		if err != nil {
			t.Fatalf("could not sign token: %v", err)
		}
		return "Bearer " + s
	}

	tests := []struct {
		name              string
		path              string
		setup             func(sqlmock.Sqlmock)
		headers           []string
		expectErrorStatus routeit.HttpStatus
		expectProceed     bool
		expectUserSet     bool
	}{
		{
			name:          "bypasses /auth path",
			path:          "/auth/login",
			expectProceed: true,
		},
		{
			name:              "missing Authorization header",
			path:              "/todos",
			expectErrorStatus: routeit.StatusUnauthorized,
		},
		{
			name:              "Authorization header present but without bearer prefix",
			path:              "/todos",
			headers:           []string{"Authorization", "Basic 123"},
			expectErrorStatus: routeit.StatusUnauthorized,
		},
		{
			name:              "malformed token",
			path:              "/todos",
			headers:           []string{"Authorization", "Bearer invalid.token.string"},
			expectErrorStatus: routeit.StatusUnauthorized,
		},
		{
			name:              "expired token",
			path:              "/todos",
			headers:           []string{"Authorization", gen(expiredClaims)},
			expectErrorStatus: routeit.StatusUnauthorized,
		},
		{
			name:              "invalid token type (refresh)",
			path:              "/todos",
			headers:           []string{"Authorization", gen(invalidClaims)},
			expectErrorStatus: routeit.StatusUnauthorized,
		},
		{
			name:    "user not found",
			path:    "/todos",
			headers: []string{"Authorization", gen(validClaims)},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name, email, password, created, updated FROM users WHERE id = \?`).
					WithArgs(user.Id).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "created", "updated"}))
			},
			expectErrorStatus: routeit.StatusUnauthorized,
		},
		{
			name:    "user lookup fails",
			path:    "/todos",
			headers: []string{"Authorization", gen(validClaims)},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name, email, password, created, updated FROM users WHERE id = \?`).
					WithArgs(user.Id).
					WillReturnError(sql.ErrConnDone)
			},
			expectErrorStatus: routeit.StatusUnauthorized,
		},
		{
			name:    "valid token, user found",
			path:    "/todos",
			headers: []string{"Authorization", gen(validClaims)},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name, email, password, created, updated FROM users WHERE id = \?`).
					WithArgs(user.Id).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "created", "updated"}).
						AddRow(user.Id, user.Name, user.Email, user.Password, user.Created, user.Updated))
			},
			expectProceed: true,
			expectUserSet: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db.WithUnitTestConnection(t, func(sqlDB *sql.DB, mock sqlmock.Sqlmock) {
				if tc.setup != nil {
					tc.setup(mock)
				}
				repo := db.NewUsersRepository(sqlDB)
				req := routeit.NewTestRequest(t, tc.path, routeit.GET, routeit.TestRequestOptions{
					Headers: tc.headers,
				})

				_, proceeded, err := routeit.TestMiddleware(JwtMiddleware(repo), req)

				if tc.expectErrorStatus.Is4xx() || tc.expectErrorStatus.Is5xx() {
					if err == nil {
						t.Fatal("expected error, got nil")
					}
					httpErr, ok := err.(*routeit.HttpError)
					if !ok {
						t.Fatalf("expected HttpError, got %T", err)
					}
					if httpErr.Status() != tc.expectErrorStatus {
						t.Errorf(`status = %+v, wanted %+v`, httpErr.Status(), tc.expectErrorStatus)
					}
					if proceeded {
						t.Errorf("expected middleware to block request, but it proceeded")
					}
				} else {
					if err != nil {
						t.Errorf("unexpected error: %v", err)
					}
					if !proceeded {
						t.Errorf("expected middleware to proceed, but it didn't")
					}
					if tc.expectUserSet {
						val, ok := req.ContextValue("user")
						if !ok {
							t.Errorf("expected user in context, but not found")
						}
						got, ok := val.(*dao.User)
						if !ok || got.Id != user.Id {
							t.Errorf("unexpected user in context: %+v", got)
						}
					}
				}

				if err := mock.ExpectationsWereMet(); err != nil {
					t.Errorf("unmet expectations: %v", err)
				}
			})
		})
	}
}
