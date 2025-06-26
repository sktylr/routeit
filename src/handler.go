package routeit

import (
	"fmt"
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

	rw.Raw(data)
	// TODO: this is a hacky workaround for css files. net/http's DetectContentType
	// does not know how to infer CSS files, so just returns text/plain.
	if strings.HasSuffix(path, "styles.css") {
		rw.hdrs.set("Content-Type", "text/css")
	}
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
