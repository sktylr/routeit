package routeit

import (
	"errors"
	"fmt"
	"io/fs"
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

func ForbiddenError() *HttpError {
	return &HttpError{status: StatusForbidden}
}

func NotFoundError() *HttpError {
	return &HttpError{status: StatusNotFound}
}

func MethodNotAllowedError() *HttpError {
	return &HttpError{status: StatusMethodNotAllowed}
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

// Converts from a general error into a HttpError, is possible. Falls back to
// a 500: Internal Server Error if no match is possible.
func toHttpError(err error) *HttpError {
	if errors.Is(err, fs.ErrPermission) {
		return ForbiddenError()
	}
	if errors.Is(err, fs.ErrNotExist) {
		return NotFoundError()
	}
	return InternalServerError()
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
