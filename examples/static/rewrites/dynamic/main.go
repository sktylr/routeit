package main

import (
	"fmt"

	"github.com/sktylr/routeit"
)

type ContactForm struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

func GetServer() *routeit.Server {
	srv := routeit.NewServer(routeit.ServerConfig{
		Debug:          true,
		StaticDir:      "assets",
		URLRewritePath: "conf/rewrites.conf",
	})
	srv.RegisterRoutesUnderNamespace("/api", routeit.RouteRegistry{
		"/contact": routeit.Post(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			var in ContactForm
			err := req.BodyFromJson(&in)
			if err != nil {
				return err
			}

			// In reality we should never ever log a user's email or name (or
			// any free form entered text) 😅
			fmt.Printf("Received contact from [%s, %s]: %#q\n", in.Name, in.Email, in.Message)

			rw.Textf("Thanks for your message %s!", in.Name)
			return nil
		}),
	})
	return srv
}

func main() {
	GetServer().StartOrPanic()
}
