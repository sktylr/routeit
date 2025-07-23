package main

import (
	"errors"
	"log"
	"time"

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
		Port:                   8080,
		AllowedHosts:           []string{".example.com", ".localhost", "[::1]"},
		StrictClientAcceptance: true,
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
			return routeit.ErrInternalServerError()
		}),
		"/panic": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			// Panics in application code are mapped to 500 Internal Server
			// Error, which helps keep the server running instead of fully
			// crashing
			panic("panicking!")
		}),
		"/bad-status": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			// routeit.HttpStatus can technically be instantiated outside the
			// package, using the following syntax. However, doing this will
			// cause the server to panic due to the client not being able to
			// set required properties on the HttpStatus struct. Doing this
			// will cause the status to be translated due to a 500: Internal
			// Server Error.
			rw.Status(routeit.HttpStatus{})
			return nil
		}),
		"/": routeit.Post(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			var body Example
			err := req.BodyFromJson(&body)
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

			return rw.Json(res)
		}),
		"/multi": routeit.MultiMethod(routeit.MultiMethodHandler{
			// The MultiMethod handler allows multiple HTTP method handlers to
			// be registered to a single route. An example would be /api/orders
			// which would respond to a GET request by listing all orders, but
			// would also respond to a POST request to create a new order. The
			// integrator does not have to provide implementations for all
			// methods and any others will by default return a 405: Method Not
			// Supported if the requested method does not have a corresponding
			// handler.
			Get: func(rw *routeit.ResponseWriter, req *routeit.Request) error {
				body := Example{
					Name: "From GET",
					Nested: Nested{
						Age:    100,
						Height: 2.0,
					},
				}
				return rw.Json(body)
			},
			Post: func(rw *routeit.ResponseWriter, req *routeit.Request) error {
				var in Nested
				err := req.BodyFromJson(&in)
				if err != nil {
					return routeit.ErrBadRequest()
				}

				res := Example{
					Name:   "From POST",
					Nested: in,
				}
				return rw.Json(res)
			},
		}),
		"/modify": routeit.Put(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			// This simple PUT endpoint just echoes the client's text/plain
			// request body back in response. It provides an example of how to
			// safely parse the text/plain request body and respond with
			// text/plain content.
			body, err := req.BodyFromText()
			if err != nil {
				return err
			}

			rw.Text(body)
			return nil
		}),
		"/delete": routeit.Delete(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			// Typically we would extract a path parameter or similar to
			// actually perform the deletion, but in this case we will just use
			// the default response for a deletion.
			return nil
		}),
		"/update": routeit.Patch(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			conflict, present := req.QueryParam("conflict")
			if !present {
				return routeit.ErrUnprocessableContent()
			}
			if conflict == "true" {
				return routeit.ErrConflict()
			}

			rw.Text("Resource updated successfully\n")
			return nil
		}),
		"/slow": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			// TODO: these timeouts are slightly unpleasant in tests :(
			time.Sleep(10*time.Second + 100*time.Millisecond)
			rw.Text("This should not be seen if timeout is working.")
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
