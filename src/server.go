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
	// The port the server listens on
	Port uint16
	// The maximum request size (headers, protocol and body inclusive) that
	// the server will accept. Anything above this will be rejected.
	RequestSize RequestSize
	// The read deadline to leave the connection with the client open for.
	ReadDeadline time.Duration
	// The write deadline that the connection is left open with the client
	// for responses.
	WriteDeadline time.Duration
	// A global namespace that **all** routes are registered under. Common
	// examples include /api. Does not need to include a leading slash.
	Namespace string
}

type Server struct {
	conf   ServerConfig
	router *router
}

func NewServer(conf ServerConfig) *Server {
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
	if conf.Namespace == "" {
		conf.Namespace = "/"
	}
	router := newRouter()
	router.globalNamespace(conf.Namespace)
	return &Server{conf: conf, router: router}
}

// Register all routes in the provided registry to the router on the server.
// All routes will already obey the global namespace (if configured). This is
// a destructive operation, meaning that if there are multiple calls to
// RegisterRoutes with overlapping values, the latest value takes precedence.
func (s *Server) RegisterRoutes(rreg RouteRegistry) {
	s.router.registerRoutes(rreg)
}

// Register all routes in the registry under a specific namespace. All routes
// already obey the global namespace (if configured). This is a destructive
// operation. For example, if the /api/foo route has already been registered,
// and this function is called with the /api namespace and the registry contains
// a /foo route, this function will overwrite the original routing entry.
//
// Examples:
// Namespace = /api
// RegisterRoutesUnderNamespace("/foo", {"/bar": ...})
// The route registered will be /api/foo/bar
//
// Namespace = <not initialised>
// RegisterRoutesUnderNamespace("/foo/bar", {"/baz": ...})
// The route will be registered under /foo/bar/baz
func (s *Server) RegisterRoutesUnderNamespace(namespace string, rreg RouteRegistry) {
	s.router.registerRoutesUnderNamespace(namespace, rreg)
}

// Attempts to start the server, panicking if that fails
func (s *Server) StartOrPanic() {
	if err := s.Start(); err != nil {
		panic(fmt.Sprintf("failed to start server: %s", err))
	}
}

// Starts the server using the config and registered routes. This should be the
// last line of a main function as any code after this call is not executable.
// The server's config is **not** thread-safe - meaning that if thread A
// initialised the server, registered routes and started the server, and thread
// B registered additional routes to the server, then thread B's modifications
// would come into affect on the live server once the scheduler gave thread B
// priority.
func (s *Server) Start() error {
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

func (s *Server) handleNewConnection(conn net.Conn) {
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

func (s *Server) handleNewRequest(raw []byte) (rw *ResponseWriter) {
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
		return NotFoundError().WithMessage(fmt.Sprintf("Invalid route: %s", req.Url())).toResponse()
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
