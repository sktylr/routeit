package main

import "github.com/sktylr/routeit"

func GetServer() *routeit.Server {
	srv := routeit.NewServer(routeit.ServerConfig{StaticDir: "static", Debug: true})
	// We don't register any routes here, instead we simply enable static disk
	// access. All requests that are under the /static URL will attempt to load
	// the corresponding file from the /static directory, returning 404 if the
	// file does not exist, and 403 if there are permission errors accessing the
	// file.
	return srv
}

func main() {
	GetServer().StartOrPanic()
}
