package routeit

import (
	"net/http"
	"os"
	"path"
	"strings"
)

type HandlerFunc func(rw *ResponseWriter, req *Request) error

type Handler struct {
	get     HandlerFunc
	head    HandlerFunc
	post    HandlerFunc
	put     HandlerFunc
	delete  HandlerFunc
	patch   HandlerFunc
	options HandlerFunc
	trace   HandlerFunc
	allowed []HttpMethod
}

type MultiMethodHandler struct {
	Get    HandlerFunc
	Post   HandlerFunc
	Put    HandlerFunc
	Delete HandlerFunc
	Patch  HandlerFunc
}

// Creates a handler that will handle GET request. Internally this will also
// handle HEAD requests which behave the same as GET requests, except the
// response does not contain the body, instead it only contains the headers
// that the GET request would return.
func Get(fn HandlerFunc) Handler {
	return MultiMethod(MultiMethodHandler{Get: fn})
}

// Creates a handler that responds to POST requests
func Post(fn HandlerFunc) Handler {
	return MultiMethod(MultiMethodHandler{Post: fn})
}

// Creates a handler that responds to PUT requests
func Put(fn HandlerFunc) Handler {
	return MultiMethod(MultiMethodHandler{Put: fn})
}

// Creates a handler that responds to DELETE requests
func Delete(fn HandlerFunc) Handler {
	return MultiMethod(MultiMethodHandler{Delete: fn})
}

// Creates a handler that responds to PATCH requests
func Patch(fn HandlerFunc) Handler {
	return MultiMethod(MultiMethodHandler{Patch: fn})
}

// Creates a handler that responds to multiple HTTP methods (e.g. GET and POST
// on the same route). The router internally will decide which handler to
// invoke depending on the method of the request. An implementation does not
// need to be provided for each of the methods in [MultiMethodHandler], it is
// sufficient to only implement the methods that the endpoint should respond
// to. The handler will ensure that any non-implemented methods return a 405:
// Method Not Allowed response.
func MultiMethod(mmh MultiMethodHandler) Handler {
	h := Handler{get: mmh.Get, post: mmh.Post, put: mmh.Put, delete: mmh.Delete, patch: mmh.Patch}
	if mmh.Get != nil {
		h.head = func(rw *ResponseWriter, req *Request) error {
			// The HEAD method is the same as GET, except it does not return a
			// response body, only headers. It is often used to determine how
			// large a resource is before committing to downloading it. routeit
			// does not allow custom implementations of HEAD handler and
			// instead provides a baked-in implementation that is automatically
			// added to all routes that respond to GET requests.
			err := h.get(rw, req)
			rw.bdy = []byte{}
			return err
		}
	}

	allow := make([]HttpMethod, 0, 6)
	if h.get != nil {
		allow = append(allow, GET)
	}
	if h.head != nil {
		allow = append(allow, HEAD)
	}
	if h.post != nil {
		allow = append(allow, POST)
	}
	if h.put != nil {
		allow = append(allow, PUT)
	}
	if h.delete != nil {
		allow = append(allow, DELETE)
	}
	if h.patch != nil {
		allow = append(allow, PATCH)
	}
	h.allowed = allow

	if len(allow) != 0 {
		h.allowed = append(h.allowed, OPTIONS)
		h.options = func(rw *ResponseWriter, req *Request) error {
			// The OPTIONS request is used to ask the server what configuration
			// it accepts. A simple implementation tells the client which
			// request methods it can use on the given endpoint, through the
			// Allow response header. If we have at least 1 method supported on
			// this handler, then we add an OPTIONS handler to the endpoint.
			for _, allow := range h.allowed {
				rw.Headers().Append("Allow", allow.name)
			}
			return nil
		}
		h.trace = func(rw *ResponseWriter, req *Request) error {
			// We can register this handler regardless of whether the server
			// enables TRACE methods. We have additional middleware that will
			// ensure TRACE methods are only let through (or mentioned in Allow
			// headers) if the TRACE method is enabled for the server.
			rw.RawWithContentType(req.body, ContentType{part: "message", subtype: "http"})
			return nil
		}
	}
	return h
}

func (h *Handler) handle(rw *ResponseWriter, req *Request) error {
	switch req.Method() {
	case GET:
		if h.get != nil {
			return h.get(rw, req)
		}
	case HEAD:
		if h.head != nil {
			return h.head(rw, req)
		}
	case POST:
		if h.post != nil {
			return h.post(rw, req)
		}
	case PUT:
		if h.put != nil {
			return h.put(rw, req)
		}
	case DELETE:
		if h.delete != nil {
			return h.delete(rw, req)
		}
	case PATCH:
		if h.patch != nil {
			return h.patch(rw, req)
		}
	case OPTIONS:
		if h.options != nil {
			return h.options(rw, req)
		}
	case TRACE:
		if h.trace != nil {
			return h.trace(rw, req)
		}
	}

	return ErrMethodNotAllowed(h.allowed...)
}

// Dynamically loads static assets from disk.
func staticLoader(namespace []string) *Handler {
	h := Get(func(rw *ResponseWriter, req *Request) error {
		url, _ := req.uri.RemoveNamespace(namespace)
		path := path.Join(url...)

		// First determine the file's presence. This allows us to return more
		// meaningful errors - e.g. if the file is not present we can map that
		// to a 404.
		if _, err := os.Stat(path); err != nil {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		cType := parseContentType(http.DetectContentType(data))
		if cType.part == "text" && cType.subtype == "plain" {
			// [net/http.DetectContentType] typically cannot infer the content
			// type of CSS or JS files, and instead returns text/plain.
			// Browsers will typically not trust stylesheets or scripts that
			// have text/plain content type, so we need to adjust for the
			// proper content type here. This works by reading the file
			// extension, so it requires that the file extension allows for
			// more accurate inference.
			if strings.HasSuffix(path, ".css") {
				cType.subtype = "css"
			} else if strings.HasSuffix(path, ".js") {
				cType.subtype = "javascript"
			}
		}
		rw.RawWithContentType(data, cType)

		return nil
	})
	return &h
}

// After all middleware is processed, the last piece is for the server to
// handle the request itself, such as method restriction. To simplify the
// logic, this is done using middleware. We force the last piece of middleware
// to always be a handler that handles the request and returns the response.
func handlingMiddleware(handler *Handler, conf handlingConfig) Middleware {
	return func(c Chain, rw *ResponseWriter, req *Request) error {
		if handler == nil {
			return ErrNotFound().WithMessagef("Invalid route: %s", req.RawPath())
		}
		if req.Method() == TRACE && !conf.AllowTraceRequests {
			return ErrMethodNotAllowed(handler.allowed...)
		}
		err := handler.handle(rw, req)
		if !conf.StrictClientAcceptance || err != nil {
			return err
		}
		if !req.AcceptsContentType(rw.ct) {
			return ErrNotAcceptable()
		}
		return nil
	}
}
