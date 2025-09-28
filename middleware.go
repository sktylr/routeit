package routeit

import (
	"slices"

	"github.com/sktylr/routeit/internal/headers"
)

// A middleware function is called for all incoming requests that reach the
// server. It can choose to block the request, or pass it off to the next
// middleware in the chain using [Chain.Proceed]. Common use cases include
// authentication or rate-limiting. The order with which middleware is
// registered to the server is important, as it defines the order of the chain.
// If a middleware chooses to block a request (by returning an error), it will
// not be propagated through to the rest of the chain, nor the handler defined
// by the application for the route and method of the request. If headers are
// set on the response, using [ResponseWriter.Headers], the headers will be
// propagated to the response - even if the handler or intermediary middleware
// returns an error or panics. The error's headers take precedence and will
// overwrite any headers of the same name that are already set.
type Middleware func(c Chain, rw *ResponseWriter, req *Request) error

// The [Chain] manages the arrangement of middleware and can be used to invoke
// the next piece of middleware.
type Chain interface {
	Proceed(rw *ResponseWriter, req *Request) error
}

type middleware struct {
	mwares []Middleware
}

// The [realChain] is a real implementation of a middleware chain used in real
// requests and E2E tests.
type realChain struct {
	i    uint
	m    *middleware
	last HandlerFunc
}

func newMiddleware() *middleware {
	return &middleware{mwares: []Middleware{}}
}

func (m *middleware) NewChain(last HandlerFunc) *realChain {
	return &realChain{i: 0, m: m, last: last}
}

// Register new middleware handlers to the middleware. The order of insertion
// matches the execution order when the middleware is invoked.
func (m *middleware) Register(ms ...Middleware) {
	m.mwares = append(m.mwares, ms...)
}

// Passes the request and response to the next piece of middleware in the
// chain. Should be called whenever the middleware does not wish to block the
// incoming request.
func (c *realChain) Proceed(rw *ResponseWriter, req *Request) error {
	length := uint(len(c.m.mwares))
	if c.i > length {
		return nil
	}

	if c.i == length {
		return c.last(rw, req)
	}

	next := c.m.mwares[c.i]
	c.i++
	return next(c, rw, req)
}

// This middleware adds the TRACE method to the response's Allow header if the
// header is present and the server supports TRACE. By default (due to TRACE
// typically being disabled for security reasons), routeit does not serve TRACE
// requests and does not include it in the Allow header or anywhere else.
// However, if the user has explicitly enabled it, we want to make sure it
// appears where it should.
func allowTraceValidationMiddleware() Middleware {
	return func(c Chain, rw *ResponseWriter, req *Request) error {
		err := c.Proceed(rw, req)

		var h headers.Headers
		if err != nil {
			if e, ok := err.(*HttpError); ok {
				h = e.headers
			} else {
				h = headers.NewHeaders()
			}
		} else {
			h = rw.headers.headers
		}

		allow, hasAllow := h.All("Allow")
		if !hasAllow {
			return err
		}

		if !slices.Contains(allow, "TRACE") {
			h.Append("Allow", "TRACE")
		}

		return err
	}
}
