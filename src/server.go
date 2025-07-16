package routeit

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
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
	// The location of the statically loaded files served by the server. All
	// requests to reach these files must also start with this prefix and may
	// also need the global Namespace if configured. The path must be a
	// subdirectory of the project's root and the server will panic if this is
	// not the case. The path does not have to point to a _valid_ directory as
	// this allows the server to dynamically write to disk and serve files from
	// there, though this is discouraged. The path is interpreted as a relative
	// path, not an absolute path, regardless of the presence of a leading slash.
	StaticDir string
	// Enables debug information, such as logs. Do not enable for production
	// servers. Example behaviour includes logging request bodies for 4xx or
	// 5xx responses.
	Debug bool
}

type Server struct {
	conf       ServerConfig
	router     *router
	log        *slog.Logger
	middleware *middleware
}

// Constructs a new server given the config. Defaults are provided for all
// options to ensure that the server can run with sane values from the get-go.
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
	router := newRouter()
	router.GlobalNamespace(conf.Namespace)
	router.NewStaticDir(conf.StaticDir)
	logOpts := slog.HandlerOptions{}
	if conf.Debug {
		logOpts.Level = slog.LevelDebug
	} else {
		logOpts.Level = slog.LevelInfo
	}
	jsonHandler := slog.NewJSONHandler(os.Stdout, &logOpts)
	s := &Server{conf: conf, router: router, log: slog.New(jsonHandler)}
	s.middleware = newMiddleware(s.handlingMiddleware)
	return s
}

// Register all routes in the provided registry to the router on the server.
// All routes will already obey the global namespace (if configured). This is
// a destructive operation, meaning that if there are multiple calls to
// RegisterRoutes with overlapping values, the latest value takes precedence.
func (s *Server) RegisterRoutes(rreg RouteRegistry) {
	s.router.RegisterRoutes(rreg)
}

// RegisterRoutesUnderNamespace registers all routes in the registry under a
// specific namespace. All routes already obey the global namespace (if
// configured). This is a destructive operation.
//
// For example, if the /api/foo route has already been registered, and this
// function is called with the /api namespace and the registry contains a /foo
// route, this function will overwrite the original routing entry.
//
// Examples:
//
//	Namespace = /api
//	RegisterRoutesUnderNamespace("/foo", {"/bar": ...})
//	The route registered will be /api/foo/bar
//
//	Namespace = <not initialised>
//	RegisterRoutesUnderNamespace("/foo/bar", {"/baz": ...})
//	The route will be registered under /foo/bar/baz
func (s *Server) RegisterRoutesUnderNamespace(namespace string, rreg RouteRegistry) {
	s.router.RegisterRoutesUnderNamespace(namespace, rreg)
}

// Registers middleware to the server. The order of registration matters, where
// the first middleware registered will be the first middleware called in the
// chain, the second will be the second and so on.
func (s *Server) RegisterMiddleware(ms ...Middleware) {
	s.middleware.Register(ms...)
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
	s.log.Info("Starting server, binding to port", "port", s.conf.Port)
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", s.conf.Port))
	if err != nil {
		s.log.Error("Failed to establish connection", "port", s.conf.Port, "err", err)
		return err
	}
	s.log.Info("Server started, ready for requests")

	for {
		conn, err := ln.Accept()
		if err != nil {
			s.log.Warn("Failed to accept incoming connection", "err", err)
			continue
		}

		now := time.Now()
		if err = conn.SetReadDeadline(now.Add(s.conf.ReadDeadline)); err != nil {
			s.log.Warn("Failed to set read deadline for incoming connection", "deadline", s.conf.ReadDeadline, "err", err)
		}
		if err = conn.SetWriteDeadline(now.Add(s.conf.WriteDeadline)); err != nil {
			s.log.Warn("Failed to set write deadline for incoming connection", "deadline", s.conf.WriteDeadline, "err", err)
		}

		go s.handleNewConnection(conn)
	}
}

// Handles an incoming connection. Extracts the raw request bytes and sends the
// raw response back to the client. Read and write deadlines are handled using
// the server config.
func (s *Server) handleNewConnection(conn net.Conn) {
	// TODO: need to choose between strings and bytes here!
	defer func() {
		// TODO: should probably look at the headers here to determine whether the conn should actually be closed, or put on a timeout for closure etc.
		conn.Close()
	}()

	buf := make([]byte, s.conf.RequestSize)
	_, err := conn.Read(buf)
	if err != nil {
		// TODO: should handle read timeouts here and return 408 Request Timeout
		s.log.Warn("Failed to read request from connection", "err", err)
		return
	}

	// TODO: remove
	fmt.Printf("-------------\nReceived: %s\n----------\n", buf)

	res := s.handleNewRequest(buf)
	_, err = conn.Write(res.write())

	if err != nil {
		s.log.Error("Failed to respond to client", "err", err)
	}
}

// Parses the raw request received from a connection and transforms it into a
// response. Handles the bulk of the server logic, such as routing, middleware
// and error handling.
func (s *Server) handleNewRequest(raw []byte) (rw *ResponseWriter) {
	req, httpErr := requestFromRaw(raw)
	if httpErr != nil {
		return httpErr.toResponse()
	}

	// This comes after the parsing of the request, since the parsing cannot
	// panic. By doing this, it means that we have access to the parsed request
	// when handling application panics.
	defer func() {
		// Prevent panics in the application code from crashing the
		// server entirely. We recover the panic and return a generic
		// 500 Internal Server Error since the fault is on the server,
		// not the client.
		if r := recover(); r != nil {
			fmt.Printf("Application code panicked: %s\n", r)
			switch e := r.(type) {
			case error:
				rw = toHttpError(e).toResponse()
			default:
				rw = InternalServerError().toResponse()
			}
		}
		// TODO: need improved error handling semantics here!
		if req.mthd == HEAD {
			rw.bdy = []byte{}
		}
	}()

	// Default to a 200 OK status code
	rw = newResponse(StatusOK)
	var err error
	chain := s.middleware.NewChain()
	err = chain.Proceed(rw, req)

	if err == nil {
		return rw
	}

	if errors.As(err, &httpErr) {
		return httpErr.toResponse()
	}

	// If the error is not a well formed httpError, then we attempt to infer
	// the type of Http error it corresponds to, falling back to 500: Internal
	// Server Error if that fails.
	return toHttpError(err).toResponse()
}

// After all middleware is processed, the last piece is for the server to
// handle the request itself, such as routing. To simplify the logic, this is
// done using middleware. We force the last piece of middleware to always be a
// handler that routes the request and returns the response.
func (s *Server) handlingMiddleware(c *Chain, rw *ResponseWriter, req *Request) error {
	handler, found := s.router.Route(req)
	if !found {
		return NotFoundError().WithMessage(fmt.Sprintf("Invalid route: %s", req.Path()))
	}
	return handler.handle(rw, req)
}
