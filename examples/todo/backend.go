package main

import (
	"database/sql"

	"github.com/sktylr/routeit"
	"github.com/sktylr/routeit/examples/todo/db"
	"github.com/sktylr/routeit/examples/todo/errors"
	"github.com/sktylr/routeit/examples/todo/handlers"
	"github.com/sktylr/routeit/examples/todo/middleware"
	"github.com/sktylr/routeit/requestid"
)

func GetBackendServer(dbConn *sql.DB) *routeit.Server {
	usersRepo := db.NewUsersRepository(dbConn)
	listsRepo := db.NewTodoListRepository(dbConn)
	itemsRepo := db.NewTodoItemRepository(dbConn)
	srv := routeit.NewServer(routeit.ServerConfig{
		Debug:                  false,
		StrictClientAcceptance: true,
		AllowedHosts:           []string{".localhost"},
		ErrorMapper:            errors.ErrorMapper,
		RequestIdProvider:      requestid.NewUuidV7Provider(),
	})
	srv.RegisterMiddleware(
		routeit.CorsMiddleware(routeit.CorsConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
			AllowedMethods: []routeit.HttpMethod{routeit.PUT, routeit.DELETE, routeit.PATCH},
			AllowedHeaders: []string{"Authorization"},
		}),
		routeit.CorsMiddleware(routeit.DefaultCors()),
		middleware.JwtMiddleware(usersRepo),
		middleware.LoadListMiddleware(listsRepo),
		middleware.LoadItemMiddleware(itemsRepo),
	)
	srv.RegisterRoutesUnderNamespace("/auth", routeit.RouteRegistry{
		"/login":    handlers.LoginHandler(usersRepo),
		"/refresh":  handlers.RefreshTokenHandler(usersRepo),
		"/register": handlers.RegisterUserHandler(usersRepo),
	})
	srv.RegisterRoutes(routeit.RouteRegistry{
		"/lists": handlers.ListsMultiHandler(listsRepo),
	})
	srv.RegisterRoutesUnderNamespace("/lists/:list", routeit.RouteRegistry{
		"/":            handlers.ListsIndividualHandler(listsRepo),
		"/items":       handlers.ItemsMultiHandler(itemsRepo),
		"/items/:item": handlers.ItemsIndividualHandler(itemsRepo),
	})
	return srv
}
