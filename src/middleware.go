package routeit

type middleware struct {
	mwares []Middleware
	last   Middleware
}

// The chain manages the arrangement of middleware and can be used to invoke
// the next piece of middleware.
type Chain struct {
	i uint
	m *middleware
}

func newMiddleware(last Middleware) *middleware {
	return &middleware{last: last, mwares: []Middleware{}}
}

func (m *middleware) NewChain() *Chain {
	return &Chain{i: 0, m: m}
}

// Register new middleware handlers to the middleware. The order of insertion
// matches the execution order when the middleware is invoked.
func (m *middleware) Register(ms ...Middleware) {
	m.mwares = append(m.mwares, ms...)
}

// Passes the request and response to the next piece of middleware in the
// chain. Should be called whenever the middleware does not wish to block the
// incoming request.
func (c *Chain) Proceed(rw *ResponseWriter, req *Request) error {
	length := uint(len(c.m.mwares))
	if c.i > length {
		return nil
	}

	if c.i == length {
		return c.m.last(c, rw, req)
	}

	next := c.m.mwares[c.i]
	c.i++
	return next(c, rw, req)
}

// A middleware function is called for all incoming requests that reach the
// server. It can choose to block the request, or pass it off to the next
// middleware in the chain using [Chain.Proceed]. Common use cases include
// authentication or rate-limiting. The order with which middleware is
// registered to the server is important, as it defines the order of the chain.
// If a middleware chooses to block a request (by returning an error), it will
// not be propagated through to the rest of the chain, nor the handler defined
// by the application for the route and method of the request.
type Middleware func(c *Chain, rw *ResponseWriter, req *Request) error
