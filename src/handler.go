package routeit

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
)

type HandlerFunc func(rw *ResponseWriter, req *Request) error

type Handler struct {
	get  HandlerFunc
	head HandlerFunc
	post HandlerFunc
}

type MultiMethodHandler struct {
	Get  HandlerFunc
	Post HandlerFunc
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

// Creates a handler that responds to multiple HTTP methods (e.g. GET and POST
// on the same route). The router internally will decide which handler to
// invoke depending on the method of the request. An implementation does not
// need to be provided for each of the methods in [MultiMethodHandler], it is
// sufficient to only implement the methods that the endpoint should respond
// to. The handler will ensure that any non-implemented methods return a 405:
// Method Not Allowed response.
func MultiMethod(mmh MultiMethodHandler) Handler {
	h := Handler{get: mmh.Get, post: mmh.Post}
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
	return h
}

func (h *Handler) handle(rw *ResponseWriter, req *Request) error {
	if req.Method() == GET && h.get != nil {
		return h.get(rw, req)
	}
	if req.Method() == HEAD && h.head != nil {
		return h.head(rw, req)
	}
	if req.Method() == POST && h.post != nil {
		return h.post(rw, req)
	}

	err := MethodNotAllowedError()
	allow := make([]string, 0, 6)
	if h.get != nil {
		allow = append(allow, GET.name)
	}
	if h.head != nil {
		allow = append(allow, HEAD.name)
	}
	if h.post != nil {
		allow = append(allow, POST.name)
	}
	err.header("Allow", strings.Join(allow, ", "))
	return err
}

// Dynamically loads static assets from disk.
func staticLoader(namespace string) *Handler {
	return &Handler{get: func(rw *ResponseWriter, req *Request) error {
		// TODO: need more generic handling of this "with namespace", "without namespace" stuff
		// TODO: probably best to actually store that on the router.
		url := req.Url()
		if namespace != "" {
			url = strings.TrimPrefix(url, namespace+"/")
		}
		path := "." + path.Clean(url)

		// TODO: currently we have no understanding of the general namespace
		// TODO: need to have better general handling here! Make sure the path is valid etc.

		// First determine the file's presence. This allows us to return more
		// meaningful errors - e.g. if the file is not present we can map that
		// to a 404.
		if _, err := os.Stat(path); err != nil {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Println("Could not load!")
			fmt.Println(err)
			return err
		}

		// TODO: need to improve how we work with this!
		cType := http.DetectContentType(data)
		if strings.HasPrefix(cType, "text/plain") && strings.HasSuffix(path, ".css") {
			// net/http.DetectContentType is not capable of inferring CSS content
			// types. This causes issues with browsers since the inferred content
			// type is text/plain which cannot be understood as a stylesheet by some
			// browsers.
			cType = strings.Replace(cType, "text/plain", "text/css", 1)
		}
		rw.RawWithContentType(data, parseContentType(cType))

		return nil
	}}
}
