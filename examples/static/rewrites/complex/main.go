package main

import "github.com/sktylr/routeit"

func GetServer() *routeit.Server {
	srv := routeit.NewServer(routeit.ServerConfig{
		StaticDir:      "static",
		URLRewritePath: "config/rewrites.conf",
	})
	return srv
}

func main() {
	GetServer().StartOrPanic()
}
