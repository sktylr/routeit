package routeit

type HandlerFunc func(rw ResponseWriter, req *Request) error
