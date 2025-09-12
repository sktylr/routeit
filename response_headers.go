package routeit

// The [ResponseHeaders] can be used to mutate the headers for a given
// [ResponseWriter]. It allows writing of header values and uses case-
// insensitive insertion for the header key.
type ResponseHeaders struct {
	headers headers
}

func newResponseHeaders() *ResponseHeaders {
	h := headers{}
	h.Set("Server", "routeit")
	return &ResponseHeaders{headers: h}
}

// Use [ResponseHeaders.Set] to completely overwrite whatever headers are
// already set for the given (case-insensitive) key. Prefer
// [ResponseHeaders.Append] where possible. This is destructive, meaning
// repeated calls using the same key will preserve the last value. Header key
// and values will be sanitised per HTTP spec before being added to the
// response. It is the user's responsibility to ensure that the headers are
// safe and non-conflicting. For example, it is heavily discouraged to modify
// the Content-Type or Content-Length headers as they are managed implicitly
// whenever a body is written to a response and can cause issues on the client
// if they contain incorrect values.
func (rh *ResponseHeaders) Set(key, val string) {
	rh.headers.Set(key, val)
}

// Append the given value to the current header value for the key. Prefer this
// to [ResponseHeaders.Set] unless the header explicitly needs to be
// overwritten. Header key and values will be sanitised per HTTP spec before
// being added to the response. It is the user's responsibility to ensure that
// the headers are safe and non-conflicting. For example, it is heavily
// discouraged to modify the Content-Type or Content-Length headers as they are
// managed implicitly whenever a body is written to a response and can cause
// issues on the client if they contain incorrect values.
func (rh *ResponseHeaders) Append(key, val string) {
	rh.headers.Append(key, val)
}
