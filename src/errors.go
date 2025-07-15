package routeit

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"strings"
)

type HttpError struct {
	status  HttpStatus
	message string
	headers headers
}

/*
 * 4xx Errors
 */

func BadRequestError() *HttpError {
	return &HttpError{status: StatusBadRequest, headers: headers{}}
}

func UnauthorizedError() *HttpError {
	return &HttpError{status: StatusUnauthorized, headers: headers{}}
}

func ForbiddenError() *HttpError {
	return &HttpError{status: StatusForbidden, headers: headers{}}
}

func NotFoundError() *HttpError {
	return &HttpError{status: StatusNotFound, headers: headers{}}
}

func MethodNotAllowedError() *HttpError {
	return &HttpError{status: StatusMethodNotAllowed, headers: headers{}}
}

func UnsupportedMediaTypeError(accepted ...ContentType) *HttpError {
	headers := headers{}
	if len(accepted) != 0 {
		var sb strings.Builder
		sb.WriteString(accepted[0].string())
		for _, accept := range accepted[1:] {
			sb.WriteString(", ")
			sb.WriteString(accept.string())
		}
		headers.Set("Accept", sb.String())
	}
	return &HttpError{status: StatusUnsupportedMediaType, headers: headers}
}

/*
 * 5xx Errors
 */

func InternalServerError() *HttpError {
	return &HttpError{status: StatusInternalServerError, headers: headers{}}
}

func NotImplementedError() *HttpError {
	return &HttpError{status: StatusNotImplemented, headers: headers{}}
}

func HttpVersionNotSupportedError() *HttpError {
	return &HttpError{status: StatusHttpVersionNotSupported, headers: headers{}}
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

func (he *HttpError) header(k string, v string) {
	he.headers.Set(k, v)
}

func (e *HttpError) toResponse() *ResponseWriter {
	rw := newResponse(e.status)
	rw.Text(e.Error())
	maps.Copy(rw.hdrs, e.headers)
	return rw
}
