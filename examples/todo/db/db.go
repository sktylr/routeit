package db

import (
	"database/sql"

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

	db, err := sql.Open("mysql", conf.FormatDSN())
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
