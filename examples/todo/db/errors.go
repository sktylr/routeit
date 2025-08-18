package db

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/mattn/go-sqlite3"
)

var (
	ErrDuplicateKey  = errors.New("duplicate key")
	ErrItemNotFound  = errors.New("item not found")
	ErrListNotFound  = errors.New("list not found")
	ErrDatabaseIssue = errors.New("database issue")
)

func isDuplicateKeyErr(err error) bool {
	var mySqlErr *mysql.MySQLError
	if errors.As(err, &mySqlErr) && mySqlErr.Number == 1062 {
		// 1062 is MySQL's code to indicate a duplicate key error occurred.
		// For user creation this is either the id or email. Due to the way
		// UUID's work, it is most likely the email.
		return true
	}

	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) &&
		sqliteErr.Code == sqlite3.ErrConstraint &&
		sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
		// SQLite3 is used in E2E tests to provide a more realistic server
		// behaviour, so we need to use it here as well.
		return true
	}

	return false
}
