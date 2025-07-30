package routeit

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"strings"
	"sync/atomic"
	"time"
)

type Server struct {
	conf         serverConfig
	router       *router
	log          *logger
	middleware   *middleware
	started      atomic.Bool
	errorHandler *errorHandler
}

// Constructs a new server given the config. Defaults are provided for all
// options to ensure that the server can run with sane values from the get-go.
func NewServer(conf ServerConfig) *Server {
	router := newRouter()
	router.GlobalNamespace(conf.Namespace)
	router.NewStaticDir(conf.StaticDir)
	errorHandler := newErrorHandler(conf.ErrorMapper)
	log := newLogger(conf.LoggingHandler, conf.Debug)
	s := &Server{conf: conf.internalise(), router: router, log: log, errorHandler: errorHandler}
	s.middleware = newMiddleware(s.handlingMiddleware)
	s.RegisterMiddleware(s.timeoutMiddleware)
	s.configureRewrites(conf.URLRewritePath)
	s.errorHandler = newErrorHandler(conf.ErrorMapper)
	if len(conf.AllowedHosts) == 0 && s.conf.Debug {
		conf.AllowedHosts = []string{".localhost", "127.0.0.1", "[::1]"}
	}
	s.RegisterMiddleware(hostValidationMiddleware(conf.AllowedHosts))
	if conf.AllowTraceRequests {
		s.RegisterMiddleware(allowTraceValidationMiddleware())
	}
	return s
}

// Register all routes in the provided registry to the router on the server.
// All routes will already obey the global namespace (if configured). This is
// a destructive operation, meaning that if there are multiple calls to
// RegisterRoutes with overlapping values, the latest value takes precedence.
func (s *Server) RegisterRoutes(rreg RouteRegistry) {
	s.panicIfStarted("register routes")
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
	s.panicIfStarted("register routes")
	s.router.RegisterRoutesUnderNamespace(namespace, rreg)
}

// Registers middleware to the server. The order of registration matters, where
// the first middleware registered will be the first middleware called in the
// chain, the second will be the second and so on.
func (s *Server) RegisterMiddleware(ms ...Middleware) {
	s.panicIfStarted("register middleware")
	s.middleware.Register(ms...)
}

// Register specific handlers for a given status code. These are called
// automatically after the entire request has finished processing, and allow
// the integrator to uniformly respond to certain 4xx or 5xx status codes.
// Common use cases include for 401 or 404 handling. For example, it may be
// desired for all 404 responses to return application/json content, which can
// be done in one place. The [RegisterErrorHandlers] method will panic if
// handlers are attempted to be registered for non 4xx or 5xx status codes.
func (s *Server) RegisterErrorHandlers(handlers map[HttpStatus]ErrorResponseHandler) {
	s.panicIfStarted("register error handlers")
	for st, h := range handlers {
		s.errorHandler.RegisterHandler(st, h)
	}
}

// Attempts to start the server, panicking if that fails
func (s *Server) StartOrPanic() {
	if err := s.Start(); err != nil {
		panic(fmt.Sprintf("failed to start server: %s", err))
	}
}

// Starts the server using the config and registered routes. This should be the
// last line of a main function as any code after this call is not executable.
// The server's config is thread-safe - meaning that if thread A initialised
// the server, registered routes and started the server, and thread B attempted
// to register additional routes to the same server, then thread B would panic.
// The server may also not be stared multiple times from different threads as
// this will also cause a panic
func (s *Server) Start() error {
	if !s.started.CompareAndSwap(false, true) {
		return errors.New("server has already been started")
	}
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
		rddl := now.Add(s.conf.ReadDeadline)
		if err = conn.SetReadDeadline(rddl); err != nil {
			s.log.Warn("Failed to set read deadline for incoming connection", "deadline", s.conf.ReadDeadline, "err", err)
		}
		if err = conn.SetWriteDeadline(rddl.Add(s.conf.ReadDeadline)); err != nil {
			s.log.Warn("Failed to set write deadline for incoming connection", "deadline", s.conf.WriteDeadline, "err", err)
		}

		go s.handleNewConnection(conn)
	}
}

// Handles an incoming connection. Extracts the raw request bytes and sends the
// raw response back to the client. Read and write deadlines are handled using
// the server config.
func (s *Server) handleNewConnection(conn net.Conn) {
	defer func() {
		conn.Close()
	}()

	buf := make([]byte, s.conf.RequestSize)
	_, err := conn.Read(buf)
	if err != nil {
		// TODO: should handle read timeouts here and return 408 Request Timeout
		s.log.Warn("Failed to read request from connection", "err", err)
		return
	}

	res := s.handleNewRequest(buf, conn.RemoteAddr())
	_, err = conn.Write(res.write())

	if err != nil {
		s.log.Error("Failed to respond to client", "err", err)
	}
}

// Parses the raw request received from a connection and transforms it into a
// response. Handles the bulk of the server logic, such as routing, middleware
// and error handling.
func (s *Server) handleNewRequest(raw []byte, addr net.Addr) (rw *ResponseWriter) {
	ctx, cancel := context.WithTimeout(context.Background(), s.conf.WriteDeadline)
	defer cancel()

	req, httpErr := requestFromRaw(raw, s.conf.RequestSize, ctx)
	if httpErr != nil {
		rw := newResponse()
		httpErr.toResponse(rw)
		return rw
	}

	if tcpAddr, ok := addr.(*net.TCPAddr); ok {
		req.ip = tcpAddr.IP.String()
	} else {
		req.ip = addr.String()
	}

	var err error
	// This comes after the parsing of the request, since the parsing cannot
	// panic. By doing this, it means that we have access to the parsed request
	// when handling application panics.
	defer func() {
		// Prevent panics in the application code from crashing the
		// server entirely. We recover the panic and return a generic
		// 500 Internal Server Error since the fault is on the server,
		// not the client.
		r := recover()
		if r == nil {
			r = err
		}
		rw = s.errorHandler.HandleErrors(r, rw, req)

		// In some cases, the HEAD request will fail - e.g. a panic or error
		// returned. In those cases, we still return the error response, but
		// must make sure the body is removed. We keep the headers as they
		// would be had the GET request succeeded, so Content-Length and
		// Content-Type are left untouched.
		if req.mthd == HEAD {
			rw.bdy = []byte{}
		}

		go s.log.LogRequestAndResponse(rw, req)
	}()

	req.uri.RewritePath(s.router)
	rw = newResponseForMethod(req.mthd)
	chain := s.middleware.NewChain()
	err = chain.Proceed(rw, req)

	// Error handling is all done in the defer block, so we can proceed here
	// without checking the error value. The reason for doing this is we have
	// multiple streams that the error can come from, since we also want to
	// avoid letting panics halt the whole server.
	return rw
}

// This is the outermost piece of middleware and ensures that the request does
// not exceed the write timeout described by the server's configuration.
func (s *Server) timeoutMiddleware(c Chain, rw *ResponseWriter, req *Request) error {
	done := make(chan any, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- r
			}
		}()
		done <- c.Proceed(rw, req)
	}()

	select {
	case result := <-done:
		switch x := result.(type) {
		case error:
			return x
		case nil:
		default:
			// This will be caught by the server and passed through the error
			// handling pipeline. We don't know what type of panic this is and
			// we (likely) did not cause it, so we don't coerce it to any other
			// type here.
			panic(x)
		}
	case <-req.ctx.Done():
		return req.ctx.Err()
	}
	return nil
}

// After all middleware is processed, the last piece is for the server to
// handle the request itself, such as routing. To simplify the logic, this is
// done using middleware. We force the last piece of middleware to always be a
// handler that routes the request and returns the response.
func (s *Server) handlingMiddleware(c Chain, rw *ResponseWriter, req *Request) error {
	handler, found := s.router.Route(req)
	if !found {
		return ErrNotFound().WithMessage(fmt.Sprintf("Invalid route: %s", req.Path()))
	}
	if req.Method() == TRACE && !s.conf.AllowTraceRequests {
		return ErrMethodNotAllowed(handler.allowed...)
	}
	err := handler.handle(rw, req)
	if !s.conf.StrictClientAcceptance || err != nil {
		return err
	}
	if !req.AcceptsContentType(rw.ct) {
		return ErrNotAcceptable()
	}
	return nil
}

// Parses the URL rewrite file, if provided, and adds all rewrite entries to
// the router. Will panic if the input is malformed or invalid in any way.
func (s *Server) configureRewrites(rewritePath string) {
	if rewritePath == "" {
		return
	}

	path := path.Clean(rewritePath)
	if !strings.HasSuffix(path, ".conf") {
		panic(fmt.Errorf(`URL rewrite file %#q is not a ".conf" file`, rewritePath))
	}

	file, err := os.Open(path)
	if err != nil {
		panic(fmt.Errorf("unable to open URL rewrite file %v", err))
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		s.router.NewRewrite(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		panic(fmt.Errorf(`error while parsing URL rewrite config %v`, err))
	}
}

func (s *Server) panicIfStarted(action string) {
	if s.started.Load() {
		panic(fmt.Errorf("cannot %s after starting the server", action))
	}
}
