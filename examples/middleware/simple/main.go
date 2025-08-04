package main

import "github.com/sktylr/routeit"

func GetServer() *routeit.Server {
	srv := routeit.NewServer(routeit.ServerConfig{Debug: true})
	srv.RegisterRoutes(routeit.RouteRegistry{
		"/hello": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			rw.Text("Hello authorised user!")
			return nil
		}),
	})
	srv.RegisterMiddleware(AuthorisationMiddleware)
	return srv
}

// Simple middleware that confirms that the request is authorised. It performs
// a simple comparison between the value of the Authorization header and the
// string literal "LET ME IN". If equal, the request is permitted to progress
// through the rest of the middleware, otherwise it is rejected.
func AuthorisationMiddleware(c routeit.Chain, rw *routeit.ResponseWriter, req *routeit.Request) error {
	auth, found := req.Headers().First("Authorization")
	if !found || auth != "LET ME IN" {
		return routeit.ErrUnauthorized()
	}

	return c.Proceed(rw, req)
}

func main() {
	GetServer().StartOrPanic()
}
