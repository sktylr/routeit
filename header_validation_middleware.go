package routeit

import "github.com/sktylr/routeit/internal/trie"

// This middleware is the second piece of middleware run on all server
// instances. It will block requests that illegally contain repeated header
// values. Some of the headers that are blocked are blocked for security
// reasons (e.g. multiple "Authorization" headers poses a security risk), while
// others are blocked due to being nonsensical (e.g. multiple "Content-Type"
// headers makes no sense and makes parsing unreliable). If the middleware
// detects multiple header values for any of the default or additionally
// supplied headers, the request will be blocked. For security purposes, the
// request is blocked on the first offender and only includes information about
// that offender.
func headerValidationMiddleware(disallow []string) Middleware {
	trie := trie.NewRuneTrie()
	defaults := []string{
		"Host",
		"Content-Length",
		"Content-Type",
		"User-Agent",
		"Authorization",
		"Origin",
		"Cookie",
		"Referer",
		"Range",
		"Expect",
	}
	// We use a trie here to benefit from not having repeated elements and
	// case-insensitive insertion.
	for _, d := range append(defaults, disallow...) {
		trie.Insert(d)
	}

	return func(c Chain, rw *ResponseWriter, req *Request) error {
		for header := range trie.Traverse() {
			if vals, found := req.Headers().All(header); found && len(vals) > 1 {
				return ErrBadRequest().WithMessagef("Header %#q cannot appear more than once", header)
			}
		}

		return c.Proceed(rw, req)
	}
}
