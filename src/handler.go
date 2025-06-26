package routeit

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
)

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

// Dynamically loads static assets from disk.
func staticLoader() Handler {
	return Handler{mthd: GET, fn: func(rw *ResponseWriter, req *Request) error {
		path := "." + path.Clean(req.Url())

		// TODO: currently we have no understanding of the general namespace
		// TODO: need to have better general handling here! Make sure the path is valid etc.

		// First determine the file's presence. This allows us to return more
		// meaningful errors - e.g. if the file is not present we can map that
		// to a 404.
		if _, err := os.Stat(path); err != nil {
			// TODO: could probably move these errors to a more generic global handler
			if errors.Is(err, fs.ErrNotExist) {
				return NotFoundError()
			}

			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Println("Could not load!")
			fmt.Println(err)

			if errors.Is(err, fs.ErrPermission) {
				return ForbiddenError()
			}
			// TODO: work on this to return meaningful error
			return err
		}

		rw.Raw(data)
		return nil
	}}
}
