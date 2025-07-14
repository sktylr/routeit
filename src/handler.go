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
	post HandlerFunc
}

// Creates a handler that will handle GET request. Internally this will also
// handle HEAD requests which behave the same as GET requests, except the
// response does not contain the body, instead it only contains the headers
// that the GET request would return.
func Get(fn HandlerFunc) Handler {
	return Handler{get: fn}
}

// Creates a handler that responds to POST requests
func Post(fn HandlerFunc) Handler {
	return Handler{post: fn}
}

func (h *Handler) handle(rw *ResponseWriter, req *Request) error {
	if req.Method() == GET && h.get != nil {
		return h.get(rw, req)
	}
	if req.Method() == HEAD && h.get != nil {
		// The HEAD method is the same as GET, except it does not return a
		// response body, only headers. It is often used to determine how
		// large a resource is before committing to downloading it.
		err := h.get(rw, req)
		rw.bdy = []byte{}
		return err
	}
	if req.Method() == POST && h.post != nil {
		return h.post(rw, req)
	}

	err := MethodNotAllowedError()
	allow := make([]string, 0, 6)
	if h.get != nil {
		allow = append(allow, GET.name, HEAD.name)
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

		cType := http.DetectContentType(data)
		if strings.HasPrefix(cType, "text/plain") && strings.HasSuffix(path, ".css") {
			// net/http.DetectContentType is not capable of inferring CSS content
			// types. This causes issues with browsers since the inferred content
			// type is text/plain which cannot be understood as a stylesheet by some
			// browsers.
			cType = strings.Replace(cType, "text/plain", "text/css", 1)
		}
		rw.RawWithContentType(data, cType)

		return nil
	}}
}
