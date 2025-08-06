package db

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
)

// [Connect] attempts to create a live connection to the sample MySQL database
// used for this application.
func Connect() (*sql.DB, error) {
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

	return db, nil
}

// [WithTestConnection] can be used to generate a mock connection to a MySQL
// database in test contexts.
func WithTestConnection(tb testing.TB, fn func(*sql.DB, sqlmock.Sqlmock)) {
	db, mock, err := sqlmock.New()
	if err != nil {
		tb.Fatalf("error while opening mock connection: %+v", err)
	}
	defer db.Close()

	fn(db, mock)
}
