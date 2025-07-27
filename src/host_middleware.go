package routeit

import (
	"strings"

	"github.com/sktylr/routeit/cmp"
)

// Middleware that is always registered as the second piece of middleware for
// the server, and rejects all incoming requests that do not match the server's
// expected Host header pattern. Per RFC-9112 Sec 7.2, the server MUST reject
// the request and return 400: Bad Request when it does not contain a Host
// header. We do the same when the Host header does not match any expected
// values.
func hostValidationMiddleware(allowed []string) Middleware {
	if len(allowed) == 0 {
		return func(c *Chain, rw *ResponseWriter, req *Request) error {
			return ErrBadRequest()
		}
	}

	hosts := make([]*cmp.ExactOrWildcard, 0, len(allowed))
	for _, h := range allowed {
		if strings.HasPrefix(h, ".") {
			hosts = append(hosts, cmp.NewDynamicWildcardMatcher("", h[1:], validSubdomain))
			hosts = append(hosts, cmp.NewExactMatcher(h[1:]))
		} else {
			hosts = append(hosts, cmp.NewExactMatcher(h))
		}
	}

	return func(c *Chain, rw *ResponseWriter, req *Request) error {
		host, hasHost := req.Header("Host")
		if !hasHost {
			return ErrBadRequest()
		}

		matches := false
		for _, h := range hosts {
			if h.Matches(host) {
				matches = true
				break
			}
		}

		if !matches {
			return ErrBadRequest()
		}

		req.host = host
		return c.Proceed(rw, req)
	}
}

func validSubdomain(seg string) bool {
	return strings.Count(seg, ".") == 1
}
