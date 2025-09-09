package main

import "github.com/sktylr/routeit"

func GetServer() *routeit.Server {
	srv := routeit.NewServer(routeit.ServerConfig{
		Debug: true,
		HttpConfig: routeit.HttpConfig{
			HttpPort:  8080,
			HttpsPort: 8443,
			TlsConfig: routeit.NewTlsConfigForCertAndKey("../certs/localhost.crt", "../certs/localhost.key"),
		},
	})
	srv.RegisterRoutes(routeit.RouteRegistry{
		"/hello": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			if req.Tls() != nil {
				rw.Text("Hello world! Thanks for being secure!")
			} else {
				rw.Text("Hello world!")
			}
			return nil
		}),
	})
	return srv
}

func main() { GetServer().StartOrPanic() }
