package routeit

import "fmt"

// TODO: should allow storing an actual message here as well
type httpError struct {
	Status HttpStatus
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

func HttpVersionNotSupportedError() *httpError {
	return &httpError{StatusHttpVersionNotSupported}
}

func (e *httpError) Error() string {
	return fmt.Sprintf("http error: %d %s", e.Status.code, e.Status.msg)
}

func (e *httpError) toResponse() *ResponseWriter {
	rw := newResponse(e.Status)
	body := fmt.Sprintf("%d: %s", e.Status.code, e.Status.msg)
	rw.Text(body)
	return rw
}
