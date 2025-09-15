package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/sktylr/routeit"
)

type CreateRequest struct {
	Name string `json:"name"`
	Age  uint   `json:"age"`
}

type CreateResponse struct {
	Message string `json:"message"`
}

func GetServer() *routeit.Server {
	be := routeit.NewServer(routeit.ServerConfig{
		Debug: true,
	})
	be.RegisterRoutes(routeit.RouteRegistry{
		"/simple": routeit.MultiMethod(routeit.MultiMethodHandler{
			Get: func(rw *routeit.ResponseWriter, req *routeit.Request) error {
				// GET is a simple method, so we don't need any CORS for this
				// method. The browser should not send any pre-flight requests
				// before requesting this resource, but it is no problem if it does
				// anyway, since our server will handle it.
				rw.Text("Hello from GET simple!\n")
				return nil
			},
			Post: func(rw *routeit.ResponseWriter, req *routeit.Request) error {
				// POST is also a simple method, so long as it accepts
				// text/plain, multipart/form-data or
				// application/x-www-urlencoded
				body, err := req.BodyFromText()
				if err != nil {
					return err
				}
				rw.Textf("Hello from POST simple with message: %s\n", body)
				return nil
			},
		}),
		"/update": routeit.Patch(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			// Despite registering this middleware, any cross-origin request
			// initiated from a browser will not reach this endpoint. This is
			// because PATCH is not in our CORS allow list, so pre-flight
			// requests will be rejected. The endpoint is still reachable from
			// the same origin, or from requests that omit the Origin header
			// entirely.
			rw.Text("Hello from PATCH /update!\n")
			return nil
		}),
		"/create": routeit.Post(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			var body CreateRequest
			if err := req.BodyFromJson(&body); err != nil {
				return err
			}

			res := CreateResponse{
				Message: fmt.Sprintf("Hello %s (age %d), thanks for your message!", body.Name, body.Age),
			}
			return rw.Json(res)
		}),
		"/remove": routeit.Delete(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			rw.Text("Deleted!")
			return nil
		}),
	})
	be.RegisterMiddleware(routeit.CorsMiddleware(routeit.CorsConfig{
		AllowedOrigins: []string{"http://localhost:*"},
		AllowedMethods: []routeit.HttpMethod{routeit.DELETE, routeit.PUT},
		AllowedHeaders: []string{"X-Requested-With", "X-Custom-Header"},
		MaxAge:         15 * time.Second,
		ExposeHeaders:  []string{"X-Response-Header"},
	}),
		AddResponseHeader)
	return be
}

func AddResponseHeader(c routeit.Chain, rw *routeit.ResponseWriter, req *routeit.Request) error {
	rw.Headers().Set("X-Response-Header", fmt.Sprintf("%v: %s", req.Method(), req.Path()))
	return c.Proceed(rw, req)
}

func main() {
	// In the main function, we serve two servers, on different ports
	var wg sync.WaitGroup
	wg.Add(2)

	// The first is the backend, which is the piece we are actually testing
	go func() {
		defer wg.Done()
		GetServer().StartOrPanic()
	}()

	// The second is the front end, which is used to demonstrate easily how
	// the cors middleware works
	go func() {
		defer wg.Done()
		srv := routeit.NewServer(routeit.ServerConfig{
			HttpConfig:     routeit.HttpConfig{HttpPort: 3000},
			Debug:          true,
			StaticDir:      "assets",
			URLRewritePath: "conf/rewrite.conf",
		})
		srv.StartOrPanic()
	}()

	wg.Wait()
}
