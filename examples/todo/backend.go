package main

import (
	"database/sql"

	"github.com/sktylr/routeit"
)

func GetBackendServer(db *sql.DB) *routeit.Server {
	return routeit.NewServer(routeit.ServerConfig{})
}
