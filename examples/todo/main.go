package main

import (
	"sync"
)

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
