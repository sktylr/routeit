package routeit

import "fmt"

// TODO: should allow storing an actual message here as well
type httpError struct {
	Status HttpStatus
}

func (e *httpError) Error() string {
	return fmt.Sprintf("http error: %d %s", e.Status.code, e.Status.msg)
}

/*
 * 4xx Errors
 */

func BadRequestError() *httpError {
	return &httpError{StatusBadRequest}
}

func NotFoundError() *httpError {
	return &httpError{StatusNotFound}
}

/*
 * 5xx Errors
 */

func InternalServerError() *httpError {
	return &httpError{StatusInternalServerError}
}

func NotImplementedError() *httpError {
	return &httpError{StatusNotImplemented}
}

func (e *httpError) toResponse() *ResponseWriter {
	hdrs := newResponseHeaders()
	hdrs["Content-Type"] = "text/plain"
	body := []byte(fmt.Sprintf("%d: %s", e.Status.code, e.Status.msg))
	hdrs["Content-Length"] = fmt.Sprintf("%d", len(body))
	return &ResponseWriter{s: e.Status, hdrs: hdrs, bdy: body}
}
