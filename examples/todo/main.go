package main

import (
	"sync"

	"github.com/sktylr/routeit"
)

func GetBackendServer() *routeit.Server {
	return nil
}

func GetFrontendServer() *routeit.Server {
	srv := routeit.NewServer(routeit.ServerConfig{
		Port:                   3000,
		StaticDir:              "static",
		URLRewritePath:         "conf/rewrites.conf",
		Debug:                  false,
		StrictClientAcceptance: true,
		AllowedHosts:           []string{".localhost", "127.0.0.1", "[::1]"},
	})
	return srv
}

func main() {
	// In the main function, we serve two servers, on different ports
	var wg sync.WaitGroup
	wg.Add(2)

	// The first is the backend, which is the piece we are actually testing
	go func() {
		defer wg.Done()
		// GetBackendServer().StartOrPanic()
	}()

	// The second is the front end, which is used to demonstrate easily how
	// the cors middleware works
	go func() {
		defer wg.Done()
		GetFrontendServer().StartOrPanic()
	}()

	wg.Wait()
}
