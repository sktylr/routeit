package main

import (
	"github.com/sktylr/routeit"
	"github.com/sktylr/routeit/requestid"
)

func GetFrontendServer() *routeit.Server {
	srv := routeit.NewServer(routeit.ServerConfig{
		Port:                   3000,
		StaticDir:              "static",
		URLRewritePath:         "conf/rewrites.conf",
		Debug:                  false,
		StrictClientAcceptance: true,
		AllowedHosts:           []string{".localhost", "127.0.0.1", "[::1]"},
		RequestIdProvider:      requestid.NewUuidV7Provider(),
	})
	return srv
}
