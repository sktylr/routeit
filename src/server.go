package routeit

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strings"
)

type ServerConfig struct {
	Port        int
	RequestSize RequestSize
}

type server struct {
	conf   ServerConfig
	router router
}

func NewServer(conf ServerConfig) *server {
	if conf.RequestSize == 0 {
		conf.RequestSize = KiB
	}
	if conf.Port == 0 {
		conf.Port = 8080
	}
	return &server{conf: conf, router: router{}}
}

func (s *server) RegisterRoutes(rreg RouteRegistry) {
	s.router.registerRoutes(rreg)
}

func (s *server) Start() error {
	fmt.Printf("Starting server on port %d\n", s.conf.Port)
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", s.conf.Port))
	if err != nil {
		fmt.Printf("Failed to establish connection on port %d\n", s.conf.Port)
		return err
	}
	fmt.Print("Server started, ready for requests...\n")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		go handleConnection(conn, func(rw *ResponseWriter, req *Request) error {
			hndl, found := s.router.route(req)
			if !found {
				// TODO: will want to do something with this in the future
				return errors.New("unsupported route")
			}
			err := hndl.fn(rw, req)
			if err != nil {
				return err
			}
			msg := rw.write()
			_, err = conn.Write(msg)
			return err
		}, s.conf.RequestSize)
	}
}

type RequestSize uint32

const (
	Byte RequestSize = 1
	KiB              = 1024 * Byte
	MiB              = 1024 * KiB
)

func handleConnection(conn net.Conn, handler HandlerFunc, reqSize RequestSize) {
	// TODO: need to choose between strings and bytes here!
	defer func() {
		conn.Close()
		fmt.Print("Response dispatched\n")
	}()

	buf := make([]byte, reqSize)
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
	err = handler(&rw, &req)
	if err != nil {
		fmt.Println(err)
	}
}
