package routeit

import (
	"crypto/tls"
	"fmt"
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
	// The maximum request size (headers, request line and body inclusive) that
	// the server will accept. Anything above this will be rejected.
	RequestSize RequestSize
	// The read deadline to leave the connection with the client open for.
	ReadDeadline time.Duration
	// The write deadline that the connection is left open with the client
	// for responses.
	WriteDeadline time.Duration
	// A global namespace that **all** routes are registered under. Common
	// examples include /api. Does not need to include a leading slash. The
	// global namespace may not contain dynamic routing segments - e.g. a
	// global namespace of /:foo will register all routes on the server
	// directly under the literal "/:foo" route.
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
	// Use [LogAttrExtractor] to include additional metadata in the default
	// request line that is dumped for all incoming requests.
	LogAttrExtractor LogAttrExtractor
	// Use [RequestIdProvider] to tag each incoming request with an ID. This
	// will automatically be logged and will be included as the "X-Request-Id"
	// header in the response. Note that the ID returned by this function is
	// allowed to be empty. In such cases, routeit will proceed the request as
	// normal, since the lack of request ID is not a strong enough reason to
	// entirely block a request. If such behaviour is desirable, it is
	// recommended to introduce custom middleware that blocks requests that
	// contain empty request ID's.
	RequestIdProvider RequestIdProvider
	// The header that each request ID is given in the response. This will
	// default to "X-Request-Id" if [ServerConfig.RequestIdProvider] is
	// non-nil. Otherwise, the header value will be ignored.
	RequestIdHeader string
	// Use [HttpConfig] to control whether the server responds to HTTP
	// requests, HTTPS requests, or both.
	HttpConfig
}

// The [HttpConfig] is used to specify details of the port(s) that the server
// will listen on. If left empty, the server will respond to HTTP requests on
// port 8080. If a [tls.Config] is provided, the server will respond to HTTPS
// requests, defaulting to listening on port 443. If it is desirable to listen
// to both HTTP and HTTPS requests, the HTTP port will need to be explicitly
// configured, commonly to port 80. In such cases, the HTTPS port only needs to
// be set if listening to HTTPS requests on port 443 is not desired. When HTTPS
// is enabled, the server can also instruct client to upgrade any HTTP
// communication to HTTPS, controlled using [HttpConfig.UpgradeToHttps] and
// [HttpConfig.UpgradeInstructionMaxAge]. If a HTTP port is not selected, but
// UpgradeToHttps is enabled, the server will listen for HTTP messages on port
// 80.
type HttpConfig struct {
	// This is the port that the HTTP listener will listen on. If the entire
	// [HttpConfig] is left empty, this will default to port 8080. If a
	// [tls.Config] is set, this will be left empty (meaning the server will
	// not respond to plain HTTP messages) unless explicitly configured.
	HttpPort uint16
	// The port that HTTPS messages are expected to be sent to. This only has
	// relevance if [HttpConfig.TlsConfig] is non-nil. If the HTTPS port is set
	// with no TLS config, server setup will panic. When a TLS config is
	// provided, the server by default listens on port 443 for HTTPS requests,
	// which can be changed with this property if required.
	HttpsPort uint16
	// The TLS config for the server. This is required if the server wishes to
	// receive and respond to HTTPS messages. When provided with no ports
	// configured, the server will listen for HTTPS messages on port 443, and
	// will not expect HTTP messages. Configure the [HttpConfig.HttpPort]
	// explicitly if it is desirable to listen to both HTTP and HTTPS requests.
	TlsConfig *tls.Config
	// Set this flag if you want the server to instruct clients to upgrade
	// their HTTP messages to HTTPS. Enabling this flag with a valid TLS config
	// and no HTTP port selected will default to the server listening for plain
	// HTTP messages on port 80, and HTTPS messages on the chosen port (or
	// defaulted to 443).
	UpgradeToHttps bool
	// Determines how long clients are instructed to remember to use HTTPS for
	// the server and its host subdomains. Set to 0 if the client should not
	// choose to remember. Only relevant when [HttpConfig.UpgradeToHttps] is
	// set, otherwise its value is ignored.
	UpgradeInstructionMaxAge time.Duration
}

// The internal server config, which only stores the necessary values
type serverConfig struct {
	HttpPort      uint16
	HttpsPort     uint16
	RequestSize   RequestSize
	ReadDeadline  time.Duration
	WriteDeadline time.Duration
	Namespace     string
	Debug         bool
	handlingConfig
}

type handlingConfig struct {
	AllowTraceRequests     bool
	StrictClientAcceptance bool
}

type httpsConfig struct {
	TlsConfig                *tls.Config
	UpgradeToHttps           bool
	UpgradeInstructionMaxAge time.Duration
}

// This is a convenience method for instantiating a TLS config with a single
// certificate and key. This will panic if the certificate or key cannot be
// loaded. [crypto/tls] sets sensible defaults for TLS config, so this is safe
// to use unless specific fine-grained control is needed.
func NewTlsConfigForCertAndKey(certPath, keyPath string) *tls.Config {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		panic(fmt.Errorf(`failed to load X509 key pair: %w`, err))
	}
	return &tls.Config{Certificates: []tls.Certificate{cert}}
}

func (sc ServerConfig) internalise() serverConfig {
	out := serverConfig{
		HttpPort:      sc.HttpPort,
		HttpsPort:     sc.HttpsPort,
		RequestSize:   sc.RequestSize,
		ReadDeadline:  sc.ReadDeadline,
		WriteDeadline: sc.WriteDeadline,
		Namespace:     sc.Namespace,
		Debug:         sc.Debug,
		handlingConfig: handlingConfig{
			StrictClientAcceptance: sc.StrictClientAcceptance,
			AllowTraceRequests:     sc.AllowTraceRequests,
		},
	}
	if sc.TlsConfig == nil {
		if sc.HttpsPort != 0 {
			panic("cannot choose a https port without a tls config")
		}
		if sc.UpgradeToHttps {
			panic("cannot upgrade to https without a tls config")
		}
	}

	if sc.RequestSize == 0 {
		out.RequestSize = KiB
	}
	if sc.TlsConfig != nil {
		if sc.HttpsPort == 0 {
			// We are using TLS so require a HTTPS port. If not supplied, we
			// default to 443
			out.HttpsPort = 443
		}
		if sc.UpgradeToHttps && sc.HttpPort == 0 {
			// Since we are explicitly upgrading HTTP to HTTPS, we need to
			// listen for HTTP requests. We don't have a port to listen on, so
			// default to 80
			out.HttpPort = 80
		}
	} else if sc.HttpPort == 0 && sc.HttpsPort == 0 {
		// No ports have been provided and we are not using TLS, so default to
		// HTTP over port 8080
		out.HttpPort = 8080
	}
	if sc.ReadDeadline == 0 {
		out.ReadDeadline = 10 * time.Second
	}
	if sc.WriteDeadline == 0 {
		out.WriteDeadline = 10 * time.Second
	}
	return out
}
