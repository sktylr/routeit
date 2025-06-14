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

type example struct {
	Name   string `json:"name"`
	Nested nested `json:"nested"`
}

type nested struct {
	Age    int     `json:"age"`
	Height float32 `json:"height"`
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

		go handleConnection(conn, func(rw ResponseWriter, req *Request) error {
			ex := example{
				Name: "John Doe",
				Nested: nested{
					Age:    25,
					Height: 1.82,
				},
			}
			err := rw.Json(ex)
			if err != nil {
				return err
			}
			return rw.write()
		})
	}
}

func handleConnection(conn net.Conn, handler HandlerFunc) {
	defer func() {
		conn.Close()
		fmt.Print("Response dispatched\n")
	}()

	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("-------------\nReceived: %s\n----------\n", buf)

	// Default to 200 OK status
	rw := ResponseWriter{conn: conn, s: StatusOK}
	err = handler(rw, &Request{})
	if err != nil {
		fmt.Println(err)
	}
}
