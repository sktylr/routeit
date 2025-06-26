package routeit

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
)

// Dynamically loads static assets from disk.
var staticLoader = Handler{mthd: GET, fn: func(rw *ResponseWriter, req *Request) error {
	path := "." + path.Clean(req.Url())

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
