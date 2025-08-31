package routeit

type testChain struct {
	proceeded bool
}

func (c *testChain) Proceed(rw *ResponseWriter, req *Request) error {
	c.proceeded = true
	return nil
}

// The [TestMiddleware] function can be used in unit tests to test a
// middleware's implementation. A [TestResponse] is returned alongside a
// boolean flag and an error. The [TestResponse] allows assertions to be made
// on the [ResponseWriter], for example to assert that a particular header was
// included or not. The boolean confirms whether the middleware proceeded or
// not, and the error is the error that the middleware ultimately returned.
func TestMiddleware(m Middleware, tr *TestRequest) (*TestResponse, bool, error) {
	rw := newResponseForMethod(tr.req.Method())
	c := &testChain{proceeded: false}

	err := m(c, rw, tr.req)
	return &TestResponse{rw: rw}, c.proceeded, err
}
