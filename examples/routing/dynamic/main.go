package main

import (
	"github.com/sktylr/routeit"
)

type InGreeting struct {
	Message string `json:"message"`
}

type OutGreeting struct {
	To      string `json:"to"`
	From    string `json:"from"`
	Message string `json:"message"`
}

type CustomGreeting struct {
	Greeting string `json:"greeting"`
}

func GetServer() *routeit.Server {
	srv := routeit.NewServer(routeit.ServerConfig{})
	srv.RegisterRoutes(routeit.RouteRegistry{
		// In this example, we demonstrate how routeit prioritises routes.
		// Requests received on the POST /hello/bob endpoint can match against
		// all requests below. Since it exactly matches a static route
		// (/hello/bob), we use that handler. If the /hello/bob handler was not
		// used, we would instead fallback to /hello/:name, since it is more
		// specific than /:greeting/bob.
		"/hello/:name": routeit.Post(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			name, _ := req.PathParam("name")
			return HelloHandler(rw, req, name, "routeit dynamic route")
		}),
		"/hello/bob": routeit.Post(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			return HelloHandler(rw, req, "bob", "routeit static route")
		}),
		"/:greeting/bob": routeit.Post(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			greeting, _ := req.PathParam("greeting")

			out := OutGreeting{
				To:      "bob",
				From:    "routeit custom greeting route",
				Message: greeting,
			}
			return rw.Json(out)
		}),
	})
	return srv
}

func HelloHandler(rw *routeit.ResponseWriter, req *routeit.Request, name string, from string) error {
	var in InGreeting
	err := req.BodyToJson(&in)
	if err != nil {
		return err
	}

	out := OutGreeting{
		To:      name,
		From:    from,
		Message: in.Message,
	}
	return rw.Json(out)
}

func main() {
	GetServer().StartOrPanic()
}
