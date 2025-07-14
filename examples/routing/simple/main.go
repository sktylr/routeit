package main

import (
	"fmt"

	"github.com/sktylr/routeit"
)

func GetServer() *routeit.Server {
	srv := routeit.NewServer(routeit.ServerConfig{})

	srv.RegisterRoutes(routeit.RouteRegistry{
		"/":                       hello("the root"),
		"/a":                      hello("/a"),
		"/a/heavily/nested":       hello("/a/heavily/nested"),
		"/a/heavily/nested/route": hello("/a/heavily/nested/route"),
	})

	return srv
}

func main() {
	GetServer().StartOrPanic()
}

func hello(msg string) routeit.Handler {
	return routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
		rw.Text(fmt.Sprintf("Hello from %s!", msg))
		rw.Status(routeit.StatusCreated)
		return nil
	})
}
