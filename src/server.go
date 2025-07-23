package routeit

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"regexp"
	"strings"
	"sync/atomic"
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
	// The path to the configuration file holding the URL rewrite information
	// for the server. It may be anywhere on the system, but must exist and be
	// readable by the server, otherwise the setup will panic. The file must be
	// a .conf file following the URL rewrite syntax. If the rewrites are
	// illegal (e.g. conflicting entries, invalid URI's or malformed in
	// general), the server setup will panic. Rewriting is not recursive, i.e.
	// if a server has rules /foo -> /bar and /bar -> /baz, an incoming request
	// to /foo will only be rewritten to /bar, it will not take the second step
	// to /baz. The [Request.Path] method always returns the request path
	// **after** rewriting.
	URLRewritePath string
	// An optional mapper to map application errors to routeit [HttpError]s
	// that are transformed to valid responses. Called whenever the application
	// code returns or panics an error
	ErrorMapper ErrorMapper
	// The allowed hosts that this server can serve from. If the incoming
	// request contains a Host header that is not satisfied by this list,
	// routeit will reject the request. Fully qualified domains can be specific
	// (e.g. www.example.com), or the domain can be prepended with . to match
	// against all subdomains. For example, .example.com will match against
	// api.example.com, www.example.com, example.com and any other subdomains
	// of example.com. Only a single layer of subdomains is considered (i.e.
	// this will not match against site.web.example.com). When Debug is enabled
	// this defaults to [".localhost", "127.0.0.1", "[::1]"] if no list is
	// specified.
	AllowedHosts []string
	// When enabled, the server will only return responses to the client that
	// strictly match the client's Accept header, if it is included in the
	// request. If the application code returns a non-compliant response
	// Content-Type, the server will automatically transform this to a 406: Not
	// Acceptable response.
	StrictClientAcceptance bool
}

// The internal server config, which only stores the necessary values
type serverConfig struct {
	Port                   uint16
	RequestSize            RequestSize
	ReadDeadline           time.Duration
	WriteDeadline          time.Duration
	Namespace              string
	Debug                  bool
	StrictClientAcceptance bool
}

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
	log := newLogger(conf.Debug)
	s := &Server{conf: conf.internalise(), router: router, log: log, errorHandler: errorHandler}
	s.middleware = newMiddleware(s.handlingMiddleware)
	s.configureRewrites(conf.URLRewritePath)
	s.errorHandler = newErrorHandler(conf.ErrorMapper)
	s.constructAllowedHosts(conf.AllowedHosts)
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
	req, httpErr := requestFromRaw(raw)
	if httpErr != nil {
		return httpErr.toResponse()
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
		// TODO: need improved error handling semantics here!
		if req.mthd == HEAD {
			rw.bdy = []byte{}
		}
		// TODO: could look into channels or goroutines here to avoid blocking on
		// the log call, since the response has been entirely computed here and can be returned to the user.
		s.log.LogRequestAndResponse(rw, req)
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

// After all middleware is processed, the last piece is for the server to
// handle the request itself, such as routing. To simplify the logic, this is
// done using middleware. We force the last piece of middleware to always be a
// handler that routes the request and returns the response.
func (s *Server) handlingMiddleware(c *Chain, rw *ResponseWriter, req *Request) error {
	handler, found := s.router.Route(req)
	if !found {
		return ErrNotFound().WithMessage(fmt.Sprintf("Invalid route: %s", req.Path()))
	}
	err := handler.handle(rw, req)
	if !s.conf.StrictClientAcceptance || err != nil {
		return err
	}
	// TODO: could store the content type on the ResponseWriter in its parsed form?
	ct, hasCt := rw.hdrs.Get("Content-Type")
	if !hasCt {
		return nil
	}
	if !req.AcceptsContentType(parseContentType(ct)) {
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

func (s *Server) constructAllowedHosts(allowed []string) {
	if len(allowed) == 0 {
		if s.conf.Debug {
			allowed = []string{".localhost", "127.0.0.1", "[::1]"}
		} else {
			s.RegisterMiddleware(hostValidationMiddleware(nil))
			return
		}
	}

	sbdmns, exact := []string{}, []string{}
	for _, host := range allowed {
		if strings.HasPrefix(host, ".") {
			sbdmns = append(sbdmns, host[1:])
		} else {
			exact = append(exact, host)
		}
	}

	var groups []string

	if len(sbdmns) > 0 {
		var parts []string
		for _, sbdmn := range sbdmns {
			parts = append(parts, regexp.QuoteMeta(sbdmn))
		}
		groups = append(groups, fmt.Sprintf(`([\w-]+\.)?(%s)`, strings.Join(parts, "|")))
	}

	for _, host := range exact {
		groups = append(groups, regexp.QuoteMeta(host))
	}

	re := regexp.MustCompile(fmt.Sprintf(`^(%s)(:\d+)?$`, strings.Join(groups, "|")))
	s.RegisterMiddleware(hostValidationMiddleware(re))
}

func (sc ServerConfig) internalise() serverConfig {
	out := serverConfig{
		Port:                   sc.Port,
		RequestSize:            sc.RequestSize,
		ReadDeadline:           sc.ReadDeadline,
		WriteDeadline:          sc.WriteDeadline,
		Namespace:              sc.Namespace,
		Debug:                  sc.Debug,
		StrictClientAcceptance: sc.StrictClientAcceptance,
	}
	if sc.RequestSize == 0 {
		out.RequestSize = KiB
	}
	if sc.Port == 0 {
		out.Port = 8080
	}
	if sc.ReadDeadline == 0 {
		out.ReadDeadline = 10 * time.Second
	}
	if sc.WriteDeadline == 0 {
		out.WriteDeadline = 10 * time.Second
	}
	return out
}
