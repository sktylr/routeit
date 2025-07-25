package routeit

import (
	"fmt"
	"regexp"
	"strings"
)

// Middleware that is always registered as the second piece of middleware for
// the server, and rejects all incoming requests that do not match the server's
// expected Host header pattern. Per RFC-9112 Sec 7.2, the server MUST reject
// the request and return 400: Bad Request when it does not contain a Host
// header. We do the same when the Host header does not match any expected
// values.
func hostValidationMiddleware(allowed []string) Middleware {
	// TODO: would be interesting to benchmark regex vs string comparison here!
	if len(allowed) == 0 {
		return func(c *Chain, rw *ResponseWriter, req *Request) error {
			return ErrBadRequest()
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

	return func(c *Chain, rw *ResponseWriter, req *Request) error {
		host, hasHost := req.Header("Host")
		if !hasHost {
			return ErrBadRequest()
		}

		if !re.MatchString(host) {
			return ErrBadRequest()
		}

		req.host = host
		return c.Proceed(rw, req)
	}
}
