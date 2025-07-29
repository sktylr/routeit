package routeit

// The [TestHandler] function allows for isolated unit tests of a handler
// function. The [TestRequest] can be composed using [NewTestRequest] and can
// take whatever headers, body and context the test needs to test the handler
// appropriately. Under test circumstances, the handler will **not** reject the
// request if the URI does not match the URI the handler is registered to.
// Currently this does not allow path parameters (e.g. those that would be
// extracted from dynamic path components) to be configured, but does allow
// query parameters to be configured using the path.
func TestHandler(h Handler, tr *TestRequest) (*TestResponse, error) {
	rw := newResponseForMethod(tr.req.Method())
	err := h.handle(rw, tr.req)
	return &TestResponse{rw: rw}, err
}
