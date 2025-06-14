package routeit

type HandlerFunc func(rw *ResponseWriter, req *Request) error

type Handler struct {
	mthd HttpMethod
	fn   HandlerFunc
}

func Get(fn HandlerFunc) Handler {
	return Handler{
		mthd: GET,
		fn:   fn,
	}
}
