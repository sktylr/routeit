package routeit

import (
	"errors"
	"fmt"
	"net"
	"time"
)

type ServerConfig struct {
	Port          int
	RequestSize   RequestSize
	ReadDeadline  time.Duration
	WriteDeadline time.Duration
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
	if conf.ReadDeadline == 0 {
		conf.ReadDeadline = 10 * time.Second
	}
	if conf.WriteDeadline == 0 {
		conf.WriteDeadline = 10 * time.Second
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

		now := time.Now()
		if err = conn.SetReadDeadline(now.Add(s.conf.ReadDeadline)); err != nil {
			fmt.Println(err)
		}
		if err = conn.SetWriteDeadline(now.Add(s.conf.WriteDeadline)); err != nil {
			fmt.Println(err)
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

	req, err := requestFromRaw(buf)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Default to 200 OK status
	rw := ResponseWriter{s: StatusOK, hdrs: map[string]string{"Server": "routeit"}}
	err = handler(&rw, req)
	if err != nil {
		fmt.Println(err)
	}
}
