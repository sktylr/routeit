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
	srv.RegisterRoutesUnderNamespace("/auth", routeit.RouteRegistry{
		"/login":    handlers.LoginHandler(usersRepo),
		"/refresh":  handlers.RefreshTokenHandler(usersRepo),
		"/register": handlers.RegisterUserHandler(usersRepo),
	})
	return srv
}
