package routeit

import (
	"fmt"
	"time"
)

// This middleware will prevent the server from acting upon HTTP requests and
// instruct the client to use HTTPS connections instead. When a plain HTTP
// request is received, we redirect the client to the equivalent resource using
// the HTTPS scheme. Many clients (e.g. browsers) will automatically handle
// this redirection. Once a HTTPS request is received, we want to tell the
// client to remember to use HTTPS for all future requests to this host, which
// is done using the Strict-Transport-Security header (RFC-6797). This header
// tells the client how long to cache the HTTPS upgrade instruction for, and
// also to respect this instruction on all subdomains of the host.
func upgradeToHttpsMiddleware(httpsPort uint16, maxAge time.Duration) Middleware {
	return func(c Chain, rw *ResponseWriter, req *Request) error {
		if req.tlsState != nil {
			// Inform the client that HTTPS is the preferred option using the
			// HTTP Strict Transport Security (HSTS) headers. Browsers will
			// cache this for as long as specified in max-age and will
			// automatically use HTTPS for subsequent requests.
			hsts := fmt.Sprintf("max-age=%d; includeSubdomains", int(maxAge.Seconds()))
			rw.Headers().Set("Strict-Transport-Security", hsts)
			return c.Proceed(rw, req)
		}

		host := req.host
		endpoint := req.RawPath()
		location := fmt.Sprintf("https://%s:%d%s", host, httpsPort, endpoint)

		// The request is over HTTP so we inform the client to redirect to the
		// equivalent HTTPS resource.
		rw.Headers().Set("Location", location)
		rw.Status(StatusMovedPermanently)
		return nil
	}
}
