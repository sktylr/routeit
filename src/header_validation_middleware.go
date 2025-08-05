package routeit

import "github.com/sktylr/routeit/trie"

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
