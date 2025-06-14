package main

import (
	"log"

	"github.com/sktylr/routeit"
)

func main() {
	srv := routeit.NewServer(8080)
	err := srv.Start()
	if err != nil {
		log.Fatal(err)
	}
}
