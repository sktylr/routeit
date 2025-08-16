package main

import (
	"database/sql"

	"github.com/sktylr/routeit"
	"github.com/sktylr/routeit/examples/todo/db"
	"github.com/sktylr/routeit/examples/todo/handlers"
	"github.com/sktylr/routeit/examples/todo/middleware"
)

func GetBackendServer(dbConn *sql.DB) *routeit.Server {
	usersRepo := db.NewUsersRepository(dbConn)
	listsRepo := db.NewTodoListRepository(dbConn)
	srv := routeit.NewServer(routeit.ServerConfig{
		Debug:                  false,
		StrictClientAcceptance: true,
		AllowedHosts:           []string{".localhost"},
	})
	srv.RegisterMiddleware(
		routeit.CorsMiddleware(routeit.DefaultCors()),
		middleware.JwtMiddleware(usersRepo),
	)
	srv.RegisterRoutesUnderNamespace("/auth", routeit.RouteRegistry{
		"/login":    handlers.LoginHandler(usersRepo),
		"/refresh":  handlers.RefreshTokenHandler(usersRepo),
		"/register": handlers.RegisterUserHandler(usersRepo),
	})
	srv.RegisterRoutes(routeit.RouteRegistry{
		"/lists":       handlers.ListsMultiHandler(listsRepo),
		"/lists/:list": handlers.ListsIndividualHandler(listsRepo),
	})
	return srv
}
