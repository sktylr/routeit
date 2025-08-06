package main

import (
	"database/sql"

	"github.com/sktylr/routeit"
	"github.com/sktylr/routeit/examples/todo/db"
	"github.com/sktylr/routeit/examples/todo/handlers"
)

func GetBackendServer(dbConn *sql.DB) *routeit.Server {
	usersRepo := db.NewUsersRepository(dbConn)
	srv := routeit.NewServer(routeit.ServerConfig{
		Debug:                  false,
		StrictClientAcceptance: true,
		AllowedHosts:           []string{".localhost"},
	})
	srv.RegisterMiddleware(routeit.CorsMiddleware(routeit.DefaultCors()))
	srv.RegisterRoutes(routeit.RouteRegistry{
		"/auth/register": handlers.RegisterUserHandler(usersRepo),
		"/auth/login":    handlers.LoginHandler(usersRepo),
	})
	return srv
}
