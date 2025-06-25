package routeit

import (
	"errors"
	"fmt"
	"net"
	"time"
)

const (
	Byte RequestSize = 1
	KiB              = 1024 * Byte
	MiB              = 1024 * KiB
)

type RequestSize uint32

type ServerConfig struct {
	Port          int
	RequestSize   RequestSize
	ReadDeadline  time.Duration
	WriteDeadline time.Duration
}

type server struct {
	conf   ServerConfig
	router *router
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
	return &server{conf: conf, router: newRouter()}
}

func (s *server) RegisterRoutes(rreg RouteRegistry) {
	s.router.registerRoutes(rreg)
}

// Attempts to start the server, panicking if that fails
func (s *server) StartOrPanic() {
	if err := s.Start(); err != nil {
		panic(fmt.Sprintf("failed to start server: %s", err))
	}
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

		go s.handleNewConnection(conn)
	}
}

func (s *server) handleNewConnection(conn net.Conn) {
	// TODO: need to choose between strings and bytes here!
	defer func() {
		conn.Close()
		fmt.Println("Response dispatched")
	}()

	buf := make([]byte, s.conf.RequestSize)
	_, err := conn.Read(buf)
	if err != nil {
		// TODO: should handle read timeouts here and return 408 Request Timeout
		fmt.Println(err)
		return
	}

	fmt.Printf("-------------\nReceived: %s\n----------\n", buf)

	res := s.handleNewRequest(buf)
	_, err = conn.Write(res.write())

	if err != nil {
		fmt.Printf("Error while responding to client: %s", err)
	}
}

func (s *server) handleNewRequest(raw []byte) (rw *ResponseWriter) {
	defer func() {
		// Prevent panics in the application code from crashing the
		// server entirely. We recover the panic and return a generic
		// 500 Internal Server Error since the fault is on the server,
		// not the client.
		if r := recover(); r != nil {
			fmt.Printf("Application code panicked: %s\n", r)
			rw = InternalServerError().toResponse()
		}
	}()

	req, httpErr := requestFromRaw(raw)
	if httpErr != nil {
		return httpErr.toResponse()
	}

	// Default to a 200 OK status code
	rw = newResponse(StatusOK)
	// TODO: will need to update this to properly set Allow headers etc if the route is valid but does not support that method.
	handler, found := s.router.route(req)
	if !found {
		return NotFoundError().toResponse()
	}
	err := handler.fn(rw, req)

	if err == nil {
		return rw
	}

	if errors.As(err, &httpErr) {
		return httpErr.toResponse()
	}

	// If the error is not a well formed httpError, then we consider it
	// to be unexpected and return a 500 Internal Server Error
	return InternalServerError().toResponse()
}
