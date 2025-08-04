package routeit

// TODO:
type ResponseHeaders struct {
	headers headers
}

func newResponseHeaders() *ResponseHeaders {
	h := headers{}
	h.Set("Server", "routeit")
	return &ResponseHeaders{headers: h}
}

// TODO:
func (rh *ResponseHeaders) Set(key, val string) {
	rh.headers.Set(key, val)
}

// TODO:
func (rh *ResponseHeaders) Append(key, val string) {
	rh.headers.Append(key, val)
}
