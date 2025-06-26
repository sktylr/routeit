package routeit

import (
	"fmt"
	"strings"
)

type HttpError struct {
	status  HttpStatus
	message string
}

/*
 * 4xx Errors
 */

func BadRequestError() *HttpError {
	return &HttpError{status: StatusBadRequest}
}

func NotFoundError() *HttpError {
	return &HttpError{status: StatusNotFound}
}

/*
 * 5xx Errors
 */

func InternalServerError() *HttpError {
	return &HttpError{status: StatusInternalServerError}
}

func NotImplementedError() *HttpError {
	return &HttpError{status: StatusNotImplemented}
}

func HttpVersionNotSupportedError() *HttpError {
	return &HttpError{status: StatusHttpVersionNotSupported}
}

// Add a custom message to the response exception. This is destructive and
// overwrites the previous message if present.
func (he *HttpError) WithMessage(message string) *HttpError {
	he.message = message
	return he
}

func (e *HttpError) Error() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d: %s", e.status.code, e.status.msg))
	if e.message != "" {
		sb.WriteString(". ")
		sb.WriteString(e.message)
	}
	return sb.String()
}

func (e *HttpError) toResponse() *ResponseWriter {
	rw := newResponse(e.status)
	rw.Text(e.Error())
	return rw
}
