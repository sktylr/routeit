package routeit

import (
	"log/slog"
	"time"
)

const (
	Byte RequestSize = 1
	KiB              = 1024 * Byte
	MiB              = 1024 * KiB
)

// The [RequestSize] type is used to define what size request the server is
// willing to accept from the client. It accounts for the entire request -
// including the request line and headers.
type RequestSize uint32

type ServerConfig struct {
	// The port the server listens on
	Port uint16
	// The maximum request size (headers, request line and body inclusive) that
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
	// By default, routeit registers TRACE handlers for all routes registered
	// to the server. If this is not desirable, leave AllowTraceRequests at its
	// default value of false. If you are happy to support TRACE requests, set
	// AllowTraceRequests to true.
	AllowTraceRequests bool
	// The logging handler used for server-specific logging. Logging within the
	// application code may use its own logging handler that is independent of
	// the given handler. If not supplied, this will use a [slog.JSONHandler]
	// that outputs to [os.Stdout]. Level defaults to INFO but will be set to
	// DEBUG if [ServerConfig.Debug] is true.
	LoggingHandler slog.Handler
	// Http request and response headers may appear multiple times in the
	// message. For certain headers, repeated entries can pose security risks
	// or not make sense when attempting to interpret the request. routeit will
	// block requests that repeat specific header values automatically. By
	// default, the list of headers where only 0 or 1 value is allowed is
	// "Authorization", "Content-Length", "Content-Type", "Cookie", "Expect",
	// "Host", "Origin", "Range", "Referer" and "User-Agent". If you have
	// additional headers that should be limited to 0 or 1 appearances, they
	// can be included in this property. Lookup is case-insensitive.
	StrictSingletonHeaders []string
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
	AllowTraceRequests     bool
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
		AllowTraceRequests:     sc.AllowTraceRequests,
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
