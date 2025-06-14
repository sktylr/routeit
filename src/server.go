package routeit

import (
	"fmt"
	"net"
)

type server struct {
	port int
}

func NewServer(port int) *server {
	return &server{port}
}

func (s *server) Start() error {
	fmt.Printf("Starting server on port %d\n", s.port)
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		fmt.Printf("Failed to establish connection on port %d\n", s.port)
		return err
	}
	fmt.Print("Server started, ready for requests...\n")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("-------------\nReceived: %s\n----------\n", buf)

	_, err = conn.Write([]byte("HTTP/1.1 200 Ok\nServer: routeit\nContent-Type: application/json\nContent-Length: 5\nCache-Control: no-store\n\nHello"))
	if err != nil {
		fmt.Println(err)
	}
}
