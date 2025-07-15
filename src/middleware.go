package routeit

type middlewareRegistry []Middleware

// The chain manages the arrangement of middleware and can be used to invoke
// the next piece of middleware.
type Chain struct {
	i   uint
	reg *middlewareRegistry
}

func newChain(reg *middlewareRegistry) *Chain {
	return &Chain{i: 0, reg: reg}
}

// Passes the request and response to the next piece of middleware in the
// chain. Should be called whenever the middleware does not wish to block the
// incoming request.
func (c *Chain) Proceed(rw *ResponseWriter, req *Request) error {
	if c.i >= uint(len(*c.reg)) {
		return nil
	}

	c.i++
	err := (*c.reg)[c.i-1](c, rw, req)
	return err
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
