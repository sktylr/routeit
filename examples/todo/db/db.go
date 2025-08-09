package db

import (
	"context"
	"database/sql"
	"testing"

	_ "embed"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
)

//go:embed sql/users.sql
var usersSchema string

//go:embed sql/lists.sql
var listsSchema string

//go:embed sql/items.sql
var itemsSchema string

// [Connect] attempts to create a live connection to the sample MySQL database
// used for this application.
func Connect(ctx context.Context) (*sql.DB, error) {
	// We hardcode all config options, including the password. In a real app we
	// wouldn't do this, but this makes it easier to reproduce across different
	// systems.
	conf := mysql.NewConfig()
	conf.User = "todo_sample_user"
	conf.Passwd = "password"
	conf.Net = "tcp"
	conf.Addr = "127.0.0.1:3306"
	conf.DBName = "todo_sample_app"
	conf.ParseTime = true

	db, err := sql.Open("mysql", conf.FormatDSN())
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if _, err := db.ExecContext(ctx, usersSchema); err != nil {
		return nil, err
	}

	if _, err := db.ExecContext(ctx, listsSchema); err != nil {
		return nil, err
	}

	if _, err := db.ExecContext(ctx, itemsSchema); err != nil {
		return nil, err
	}

	return db, nil
}

// [WithUnitTestConnection] can be used to simulate a mock connection to a
// MySQL database in unit test contexts.
func WithUnitTestConnection(tb testing.TB, fn func(*sql.DB, sqlmock.Sqlmock)) {
	db, mock, err := sqlmock.New()
	if err != nil {
		tb.Fatalf("error while opening mock connection: %+v", err)
	}
	defer db.Close()

	fn(db, mock)
}

// [WithIntegrationTestConnection] spins up a SQLite3 in-memory database that
// can be used to run integration tests. Although slower than mocking, this
// allows the test to perform real SQL queries, so poses a more realistic base
// upon which the tests are executed.
func WithIntegrationTestConnection(tb testing.TB, fn func(*sql.DB)) {
	dbConn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		tb.Fatalf("error while opening db: %v", err)
	}
	defer dbConn.Close()

	if _, err = dbConn.ExecContext(tb.Context(), usersSchema); err != nil {
		tb.Fatalf(`failed to initialise db: %v`, err)
	}

	if _, err = dbConn.ExecContext(tb.Context(), listsSchema); err != nil {
		tb.Fatalf(`failed to initialise db: %v`, err)
	}

	if _, err = dbConn.ExecContext(tb.Context(), itemsSchema); err != nil {
		tb.Fatalf(`failed to initialise db: %v`, err)
	}

	fn(dbConn)
}
