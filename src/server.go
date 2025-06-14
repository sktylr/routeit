package routeit

import (
	"bytes"
	"fmt"
	"net"
	"strings"
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
			msg := rw.write()
			_, err = conn.Write(msg)
			return err
		})
	}
}

func handleConnection(conn net.Conn, handler HandlerFunc) {
	// TODO: need to choose between strings and bytes here!
	defer func() {
		conn.Close()
		fmt.Print("Response dispatched\n")
	}()

	// TODO: make this configurable
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("-------------\nReceived: %s\n----------\n", buf)

	lines := bytes.Split(buf, []byte("\n"))
	ptcl := bytes.SplitN(bytes.TrimSpace(lines[0]), []byte(" "), 3)
	if len(ptcl) != 3 {
		fmt.Print("Unexpected HTTP protocol line!\n")
		return
	}

	ver := string(ptcl[2])
	if ver != "HTTP/1.1" {
		fmt.Print("Unsupported HTTP version!\n")
		return
	}

	mthd, found := parseMethod(string(ptcl[0]))
	if !found {
		fmt.Print("Unsupported HTTP Method!\n")
		return
	}

	path := string(ptcl[1])
	pathParams := pathParameters{}
	foo := strings.Split(path, "?")
	endpt := foo[0]
	if len(foo) > 1 {
		if len(foo) > 2 {
			fmt.Print("Unexpected number of query options!\n")
			return
		}

		queries := foo[1]
		for _, query := range strings.Split(queries, "&") {
			kvp := strings.SplitN(query, "=", 2)
			if len(kvp) != 2 {
				fmt.Print("Query string malformed!\n")
				continue
			}
			pathParams[kvp[0]] = kvp[1]
		}
	}

	reqHdrs := headers{}
	var end int
	fmt.Printf("Number of lines: %d\n", len(lines))
	for i, line := range lines {
		// ?????
		end = i
		if i == 0 {
			continue
		}
		sline := strings.TrimSpace(string(line))
		if sline == "" {
			// Blank line between headers and body
			break
		}

		kvp := strings.SplitN(sline, ": ", 2)
		if len(kvp) != 2 {
			fmt.Printf("Malformed header: [%s]\n", sline)
			continue
		}
		reqHdrs[kvp[0]] = kvp[1]
	}
	fmt.Printf("Ending on %d\n", end)
	var sb strings.Builder
	for _, line := range lines[end:] {
		sb.Write(bytes.TrimSpace(line))
	}
	req := Request{mthd: mthd, url: endpt, pathParams: pathParams, headers: reqHdrs, body: sb.String()}

	// Default to 200 OK status
	rw := ResponseWriter{s: StatusOK, hdrs: map[string]string{"Server": "routeit"}}
	err = handler(rw, &req)
	if err != nil {
		fmt.Println(err)
	}
}
