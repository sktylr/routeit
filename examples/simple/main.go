package main

import (
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
	srv := routeit.NewServer(8080)
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
	})
	err := srv.Start()
	if err != nil {
		log.Fatal(err)
	}
}
