package main

import (
	"errors"
	"log"

	"github.com/sktylr/routeit"
)

type Greeting struct {
	From Example `json:"from"`
	To   Example `json:"to"`
}

type Example struct {
	Name   string `json:"name"`
	Nested Nested `json:"nested"`
}

type Nested struct {
	Age    int     `json:"age"`
	Height float32 `json:"height"`
}

func GetServer() *routeit.Server {
	srv := routeit.NewServer(routeit.ServerConfig{
		Port: 8080,
	})
	srv.RegisterRoutes(routeit.RouteRegistry{
		"/hello": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			ex := Example{
				Name: "John Doe",
				Nested: Nested{
					Age:    25,
					Height: 1.82,
				},
			}
			return rw.Json(ex)
		}),
		"/echo": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			msg, found := req.QueryParam("message")
			if !found {
				rw.Text("Looks like you didn't want me to echo anything!\n")
				return nil
			}

			rw.Textf("Received message to echo: %s\n", msg)
			return nil
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
		"/": routeit.Post(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			var body Example
			err := req.BodyToJson(&body)
			if err != nil {
				return err
			}

			res := Greeting{From: body}
			res.To = Example{
				Name: "Jane Doe",
				Nested: Nested{
					Age:    29,
					Height: 1.62,
				},
			}

			rw.Json(res)
			return nil
		}),
	})
	return srv
}

func main() {
	srv := GetServer()
	err := srv.Start()
	if err != nil {
		log.Fatal(err)
	}
}
