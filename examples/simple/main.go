package main

import (
	"errors"
	"log"

	"github.com/sktylr/routeit"
)

type example struct {
	Name   string `json:"name"`
	Nested nested `json:"nested"`
}

type nested struct {
	Age    int     `json:"age"`
	Height float32 `json:"height"`
}

func main() {
	srv := routeit.NewServer(routeit.ServerConfig{
		Port: 8080,
	})
	srv.RegisterRoutes(routeit.RouteRegistry{
		"/hello": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			ex := example{
				Name: "John Doe",
				Nested: nested{
					Age:    25,
					Height: 1.82,
				},
			}
			return rw.Json(ex)
		}),
		"/error": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			// This route returns an error which is mapped internally into a
			// 500 Internal Server Error and propagated into the response
			return errors.New("custom error")
		}),
		"/crash": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			// This route deliberately returns a 500 Internal Server Error,
			// which shows how an integration could deliberately return an
			// error that is propagated into the http response
			return routeit.InternalServerError()
		}),
		"/panic": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			// Panics in application code are mapped to 500 Internal Server
			// Error, which helps keep the server running instead of fully
			// crashing
			panic("panicking!")
		}),
	})
	err := srv.Start()
	if err != nil {
		log.Fatal(err)
	}
}
