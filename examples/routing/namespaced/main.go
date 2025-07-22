package main

import (
	"fmt"

	"github.com/sktylr/routeit"
)

func GetServer() *routeit.Server {
	srv := routeit.NewServer(routeit.ServerConfig{Namespace: "/api", Debug: true})

	registry := routeit.RouteRegistry{"/hello": EchoUrlHandler()}

	// Since we've configured a global namespace, all requests will be routed
	// under /api. So we are actually registering /api/hello, which will echo
	// the URL to the client.
	srv.RegisterRoutes(registry)

	// This creates another layer of namespacing and additionally registers
	// the provided routes under the provided namespace. This will register a
	// new route at /api/namespace/hello
	srv.RegisterRoutesUnderNamespace("/namespace", registry)

	return srv
}

func main() {
	GetServer().StartOrPanic()
}

func EchoUrlHandler() routeit.Handler {
	return routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
		path := req.Path()
		rw.Text(fmt.Sprintf(`Hello from "%s"`, path))
		return nil
	})
}
