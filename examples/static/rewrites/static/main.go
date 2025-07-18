package main

import "github.com/sktylr/routeit"

type Hello struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

func GetServer() *routeit.Server {
	srv := routeit.NewServer(routeit.ServerConfig{
		Port:           3000,
		StaticDir:      "statics",
		URLRewritePath: "conf/rewrites.conf",
	})
	srv.RegisterRoutesUnderNamespace("/hello", routeit.RouteRegistry{
		"/there": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			hello := Hello{
				Name:    "/hello/there endpoint",
				Message: `This single handler responds to two edge request paths: "/hello/there" and "/rewrite/me/please", which is handled via URL rewrites.`,
			}

			return rw.Json(hello)
		}),
	})
	return srv
}

func main() {
	GetServer().StartOrPanic()
}
