package routeit

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/sktylr/routeit/cmp"
	"github.com/sktylr/routeit/trie"
)

type CorsConfig struct {
	// A list of origins the server can accept cross-origin requests from.
	// These can either be literal matches for the origin (case sensitive), or
	// wildcard matches, denoted by the * character. There can only be at most
	// 1 wildcard character in each element in this list.
	//
	// Example:
	//	["http://localhost:*", "http://example.com", "http://*.example.com"]
	//
	// This will match against any requests from localhost, regardless of the
	// port, any request from example.com, and any request from a subdomain of
	// example.com. If the Origin header is present and does not match any of
	// these patterns, we will reject the request.
	AllowedOrigins []string
	// Set to true if all origins should be allowed. This is equivalent to
	// setting [CorsConfig.AllowedOrigins] = ["*"] or always returning true, ""
	// from [CorsConfig.AllowOriginFunc]. Takes precedence over
	// [CorsConfig.AllowedOrigins] if both are set.
	AllowAllOrigins bool
	// A function that determines if the request's origin should be allowed.
	// This is only called when the request contains the Origin header, and can
	// also return additional Vary header values. Will take precedence over
	// [CorsConfig.AllowAllOrigins] and [CorsConfig.AllowedOrigins].
	AllowOriginFunc AllowOriginFunc
	// The allowed methods for cross-origin requests. This will always contain
	// simple methods (GET, HEAD and POST), as well as OPTIONS, which is
	// required for pre-flight requests. Note: this only carries relevance when
	// handling pre-flight requests initiated by the browser. It will not block
	// actual requests containing these methods. If you would like to block
	// specific HTTP methods from reaching the server, it is best to implement
	// your own middleware. Otherwise, you can simply avoid providing an
	// implementation of the corresponding method (i.e. never implement a PUT
	// handler); in this way the server will never respond successfully to
	// requests of those methods.
	AllowedMethods []HttpMethod
	// A list of the headers that the server will also accept. These do not
	// need to include the CORS safe headers (Accept, Accept-Language,
	// Content-Language, Content-Type and Range) and should be headers that
	// this server is willing to accept in incoming requests. This only
	// confirms to the client that it is permitted to send a request with these
	// headers. The server may still employ additional middleware or a handler
	// function that rejects the request if the header is not correctly
	// populated.
	AllowedHeaders []string
	// The maximum age that the client (or intermediaries) should cache the
	// pre-flight response for. Set to 0 or leave unset if you don't want this
	// to be included. Set to a negative number to explicitly set to 0 - i.e.
	// the client cannot cache responses.
	MaxAge time.Duration
	// The headers that should be exposed to the client JavaScript when
	// receiving a response to a cross-origin request. By default, browsers
	// will expose the Cache-Control, Content-Language, Content-Type, Expires,
	// Last-Modified and Pragma response headers to the client JavaScript. If
	// additional headers should be included (e.g. X-My-Custom-Header), then
	// they should be included here. The casing of the header values does not
	// matter.
	ExposeHeaders []string
	// Determines whether the Access-Control-Allow-Credentials header should be
	// included in responses and set to "true". This is required if the client
	// and server want the client to send credentials (cookies, HTTP
	// authentication, TLS certificates etc.).
	IncludeCredentials bool
}

// This function should validate the provided origin, returning true if the
// server should accept cross-origin requests from this origin. Optionally, the
// function can choose to return values for the Vary response header, to ensure
// that caches are serving responses correctly. This will be appended to the
// existing Vary response header, which will already contain Origin.
type AllowOriginFunc func(*Request, string) (bool, string)

type cors struct {
	AllowsOrigin       AllowOriginFunc
	AllowedMethods     []HttpMethod
	AllowedHeaders     func(string) bool
	MaxAge             string
	ExposeHeaders      string
	IncludeCredentials bool
}

// Default CORS config that will allow all origins and all methods and not
// include any special headers or authentication details
func DefaultCors() CorsConfig {
	return CorsConfig{
		AllowAllOrigins: true,
		AllowedMethods:  []HttpMethod{PUT, DELETE, PATCH},
	}
}

// Returns middleware that allows CORS requests from the client to function
// properly on the server. This should be placed early in the middleware stack
// and ideally before authentication middleware or most custom middleware. This
// middleware will handle any cross-origin request, including pre-flight and
// actual requests.
func CorsMiddleware(cc CorsConfig) Middleware {
	cors := cc.toCors()

	return func(c Chain, rw *ResponseWriter, req *Request) error {
		// Always set the Vary header to at least Origin. This ensures that
		// intermediary nodes, such as CDNs or proxies know not to return a
		// cached response if the Origin header is different as this could
		// cause CORS headers to leak over different origins.
		setOrAppend(rw, "Vary", "Origin")

		origin, hasOrigin := req.Header("Origin")
		if !hasOrigin {
			return c.Proceed(rw, req)
		}
		allowedOrigin, additionalVaryHeaders := cors.AllowsOrigin(req, origin)
		setOrAppend(rw, "Vary", additionalVaryHeaders)
		if !allowedOrigin {
			return ErrForbidden()
		}

		acrm, hasAcrm := req.Header("Access-Control-Request-Method")
		if req.Method() == OPTIONS && hasAcrm {
			// This is a pre-flight request sent by the browser for non-simple
			// requests. These requests are sent by browsers to inform the
			// server of the method and headers it will send so that the server
			// can say whether it will accept the request or not.

			// For the same reasons above, we only want the intermediaries
			// to return this exact cached response if the
			// Access-Control-Request-Method header is the same
			setOrAppend(rw, "Vary", "Access-Control-Request-Method")

			if !cors.IsAllowedMethod(acrm) {
				return ErrMethodNotAllowed(cors.AllowedMethods...)
			}

			// This is a hack to skip to the last piece of middleware, which
			// will respond with the methods the route supports
			if rc, ok := c.(*realChain); ok {
				rc.i = uint(len(rc.m.mwares))
			}
			c.Proceed(rw, req)

			allow, _ := rw.hdrs.Get("Allow")
			if !strings.Contains(allow, acrm) {
				return ErrMethodNotAllowed()
			}

			rw.Header("Access-Control-Allow-Origin", origin)
			rw.Header("Access-Control-Allow-Methods", acrm)
			if cors.MaxAge != "" {
				rw.Header("Access-Control-Max-Age", cors.MaxAge)
			}

			headers, requestsHeaders := req.Header("Access-Control-Request-Headers")
			if requestsHeaders && cors.AllowedHeaders(headers) {
				// For security purposes, we only confirm that exactly the
				// headers requested by the client will be accepted by the
				// server, even if there are others that may be.
				setOrAppend(rw, "Vary", "Access-Control-Request-Headers")
				rw.Header("Access-Control-Allow-Headers", headers)
			}

			if cors.IncludeCredentials {
				rw.Header("Access-Control-Allow-Credentials", "true")
			}

			// If we don't include CORS specific headers, the browser will know
			// to reject the request and not allow the cross origin request to
			// proceed.
			return nil
		}

		// Process the actual request. All we need to do is add the required
		// Access-Control-* headers and the client is expected to take care of
		// the rest.
		rw.Header("Access-Control-Allow-Origin", origin)
		if cors.ExposeHeaders != "" {
			rw.Header("Access-Control-Expose-Headers", cors.ExposeHeaders)
		}
		if cors.IncludeCredentials {
			rw.Header("Access-Control-Allow-Credentials", "true")
		}
		return c.Proceed(rw, req)
	}
}

func (c *cors) IsAllowedMethod(raw string) bool {
	method := HttpMethod{name: raw}
	return method.isValid() && slices.Contains(c.AllowedMethods, method)
}

func (cc CorsConfig) toCors() *cors {
	// By default we must allow OPTIONS requests, otherwise the server cannot
	// handle pre-flight requests. We also by default always support the simple
	// methods
	allowedMethods := stripDuplicates(
		append([]HttpMethod{OPTIONS, GET, HEAD, POST}, cc.AllowedMethods...),
	)

	// Despite being a CORS safe header, Content-Type may appear in the
	// requested headers header due to its value not being a simple
	// Content-Type value.
	allowedHeaders := stripDuplicates(append(cc.AllowedHeaders, "content-type"))

	trie := trie.NewRuneTrie()
	for _, h := range allowedHeaders {
		trie.Insert(h)
	}

	cors := &cors{
		AllowsOrigin:   cc.generateAllowsOrigin(),
		AllowedMethods: allowedMethods,
		AllowedHeaders: func(raw string) bool {
			split := strings.FieldsFunc(raw, func(r rune) bool { return r == ',' })
			for _, h := range split {
				if !trie.Contains(h) {
					return false
				}
			}
			return true
		},
		ExposeHeaders:      strings.Join(cc.ExposeHeaders, ", "),
		IncludeCredentials: cc.IncludeCredentials,
	}

	if cc.MaxAge > 0 {
		cors.MaxAge = strconv.FormatInt(int64(cc.MaxAge.Seconds()), 10)
	} else if cc.MaxAge < 0 {
		cors.MaxAge = "0"
	}

	return cors
}

func (cc CorsConfig) generateAllowsOrigin() AllowOriginFunc {
	acceptAll := func(req *Request, s string) (bool, string) { return true, "" }
	if cc.AllowOriginFunc != nil {
		return cc.AllowOriginFunc
	} else if cc.AllowAllOrigins {
		return acceptAll
	} else {
		allowedOrigins := stripDuplicates(cc.AllowedOrigins)
		origins := make([]*cmp.ExactOrWildcard, 0, len(allowedOrigins))
		for _, o := range allowedOrigins {
			if o == "*" {
				return acceptAll
			}
			stars := strings.Count(o, "*")
			if stars > 1 {
				panic(fmt.Errorf("cannot specify multiple wildcards in allowed origins: %s", o))
			}
			if stars == 1 {
				i := strings.IndexRune(o, '*')
				origin := cmp.NewWildcardMatcher(o[:i], o[i+1:])
				origins = append(origins, origin)
			} else {
				origins = append(origins, cmp.NewExactMatcher(o))
			}
		}

		return func(req *Request, s string) (bool, string) {
			for _, o := range origins {
				if o.Matches(s) {
					return true, ""
				}
			}
			return false, ""
		}
	}
}

func setOrAppend(rw *ResponseWriter, k, v string) {
	val, hasVal := rw.hdrs.Get(k)
	if !hasVal {
		rw.Header(k, v)
	} else if !strings.Contains(val, v) {
		rw.Header(k, fmt.Sprintf("%s, %s", val, v))
	}
}
