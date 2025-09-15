package main

import (
	"time"

	"github.com/sktylr/routeit"
)

func GetServer() *routeit.Server {
	srv := routeit.NewServer(routeit.ServerConfig{
		AllowedHosts: []string{".localhost", "example.com", "127.0.0.1", "[::1]"},
		HttpConfig: routeit.HttpConfig{
			HttpPort:                 8123,
			HttpsPort:                8443,
			TlsConfig:                routeit.NewTlsConfigForCertAndKey("../certs/localhost.crt", "../certs/localhost.key"),
			UpgradeToHttps:           true,
			UpgradeInstructionMaxAge: time.Second,
		},
	})
	srv.RegisterRoutes(routeit.RouteRegistry{
		"/echo": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			msg, hasMsg := req.Queries().First("message")
			if !hasMsg {
				msg = "You didn't send a message!"
			}
			rw.Text(msg)
			return nil
		}),
	})
	return srv
}

func main() { GetServer().StartOrPanic() }
