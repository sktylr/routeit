package routeit

// A [RequestIdProvider] is used to provide each incoming request with an ID.
// Each valid request that reaches the server will be tagged with an ID
// returned from this function. The resultant ID will be available on the
// request using [Request.Id].
type RequestIdProvider func(*Request) string

// This middleware adds a request ID to each incoming request. When installed
// to the server, it is installed after all other built-in middlewares. The
// request ID is also added to each response using a header, which can help
// with debugging. This header defaults to "X-Request-Id", but can be
// customised.
func requestIdMiddleware(prov RequestIdProvider, header string) Middleware {
	if header == "" {
		header = "X-Request-Id"
	}

	return func(c Chain, rw *ResponseWriter, req *Request) error {
		req.id = prov(req)
		rw.Headers().Set(header, req.id)
		return c.Proceed(rw, req)
	}
}
