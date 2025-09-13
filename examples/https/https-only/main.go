package main

import (
	"crypto/tls"

	"github.com/sktylr/routeit"
)

type InfoResponse struct {
	Version     string `json:"version"`
	CipherSuite string `json:"cipher_suite"`
	ServerName  string `json:"server_name"`
}

func GetServer() *routeit.Server {
	srv := routeit.NewServer(routeit.ServerConfig{
		Debug: true,
		HttpConfig: routeit.HttpConfig{
			HttpsPort: 8443,
			TlsConfig: routeit.NewTlsConfigForCertAndKey("../certs/localhost.crt", "../certs/localhost.key"),
		},
	})
	srv.RegisterRoutes(routeit.RouteRegistry{
		"/info": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			// Since the server will only accept TLS-backed HTTPS request over
			// port 8443, we can safely assume that [routeit.Request.Tls] is
			// non-nil.
			res := InfoResponse{
				Version:     tls.VersionName(req.Tls().Version),
				CipherSuite: tls.CipherSuiteName(req.Tls().CipherSuite),
				ServerName:  req.Tls().ServerName,
			}
			return rw.Json(res)
		}),
	})
	return srv
}

func main() { GetServer().StartOrPanic() }
