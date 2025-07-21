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
	cause   error
	message string
	headers headers
}

// The [ErrorMapper] can be implemented to provide more granular control over
// error mapping in the server. This function will be called whenever an
// application error is returned or recovered from a panic and can provide
// extra control over how errors are transformed. By default, the server will
// handle known mappings, such as interpreting an ErrNotExists error as a 404:
// Not Found error. If no mapping can be found, the server will default to a
// 500: Internal Server Error. If the custom [ErrorMapper] cannot identify the
// [HttpError] it should return, it may return nil
type ErrorMapper func(e error) *HttpError

// The [ErrorResponseWriter] is very similar to the [ResponseWriter], except it
// does not allow mutation of the status code of the response. It is used in
// custom error handlers that can provide a uniform response to the client in
// the event of a 4xx or 5xx error. Common use cases include a custom 404
// response.
type ErrorResponseWriter struct {
	rw  *ResponseWriter
	err error
}

type ErrorResponseHandler func(erw *ErrorResponseWriter, req *Request)

type errorHandler struct {
	handlers map[HttpStatus]ErrorResponseHandler
	em       ErrorMapper
}

/*
 * 4xx Errors
 */

func ErrBadRequest() *HttpError {
	return httpErrorForStatus(StatusBadRequest)
}

func ErrUnauthorized() *HttpError {
	return httpErrorForStatus(StatusUnauthorized)
}

func ErrPaymentRequired() *HttpError {
	return httpErrorForStatus(StatusPaymentRequired)
}

func ErrForbidden() *HttpError {
	return httpErrorForStatus(StatusForbidden)
}

func ErrNotFound() *HttpError {
	return httpErrorForStatus(StatusNotFound)
}

func ErrMethodNotAllowed(allowed ...HttpMethod) *HttpError {
	allow := make([]string, 0, len(allowed))
	for _, m := range allowed {
		allow = append(allow, m.name)
	}
	headers := headers{}
	headers.Set("Allow", strings.Join(allow, ", "))
	return &HttpError{status: StatusMethodNotAllowed, headers: headers}
}

func ErrNotAcceptable() *HttpError {
	return httpErrorForStatus(StatusNotAcceptable)
}

func ErrProxyAuthenticationRequired() *HttpError {
	return httpErrorForStatus(StatusProxyAuthenticationRequired)
}

func ErrRequestTimeout() *HttpError {
	return httpErrorForStatus(StatusRequestTimeout)
}

func ErrConflict() *HttpError {
	return httpErrorForStatus(StatusConflict)
}

func ErrGone() *HttpError {
	return httpErrorForStatus(StatusGone)
}

func ErrLengthRequired() *HttpError {
	return httpErrorForStatus(StatusLengthRequired)
}

func ErrPreconditionFailed() *HttpError {
	return httpErrorForStatus(StatusPreconditionFailed)
}

func ErrContentTooLarge() *HttpError {
	return httpErrorForStatus(StatusContentTooLarge)
}

func ErrURITooLong() *HttpError {
	return httpErrorForStatus(StatusURITooLong)
}

func ErrUnsupportedMediaType(accepted ...ContentType) *HttpError {
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

func ErrRangeNotSatisfiable() *HttpError {
	return httpErrorForStatus(StatusRangeNotSatisfiable)
}

func ErrExpectationFailed() *HttpError {
	return httpErrorForStatus(StatusExpectationFailed)
}

func ErrImATeapot() *HttpError {
	return httpErrorForStatus(StatusImATeapot)
}

func ErrMisdirectedRequest() *HttpError {
	return httpErrorForStatus(StatusMisdirectedRequest)
}

func ErrUnprocessableContent() *HttpError {
	return httpErrorForStatus(StatusUnprocessableContent)
}

func ErrLocked() *HttpError {
	return httpErrorForStatus(StatusLocked)
}

func ErrFailedDependency() *HttpError {
	return httpErrorForStatus(StatusFailedDependency)
}

func ErrTooEarly() *HttpError {
	return httpErrorForStatus(StatusTooEarly)
}

func ErrUpgradeRequired() *HttpError {
	return httpErrorForStatus(StatusUpgradeRequired)
}

func ErrPreconditionRequired() *HttpError {
	return httpErrorForStatus(StatusPreconditionRequired)
}

func ErrTooManyRequests() *HttpError {
	return httpErrorForStatus(StatusTooManyRequests)
}

func ErrRequestHeaderFieldsTooLarge() *HttpError {
	return httpErrorForStatus(StatusRequestHeaderFieldsTooLarge)
}

func ErrUnavailableForLegalReasons() *HttpError {
	return httpErrorForStatus(StatusUnavailableForLegalReasons)
}

/*
 * 5xx Errors
 */

func ErrInternalServerError() *HttpError {
	return httpErrorForStatus(StatusInternalServerError)
}

func ErrNotImplemented() *HttpError {
	return httpErrorForStatus(StatusNotImplemented)
}

func ErrBadGateway() *HttpError {
	return httpErrorForStatus(StatusBadGateway)
}

func ErrServiceUnavailable() *HttpError {
	return httpErrorForStatus(StatusServiceUnavailable)
}

func ErrGatewayTimeout() *HttpError {
	return httpErrorForStatus(StatusGatewayTimeout)
}

func ErrHttpVersionNotSupported() *HttpError {
	return httpErrorForStatus(StatusHttpVersionNotSupported)
}

func ErrVariantAlsoNegotiates() *HttpError {
	return httpErrorForStatus(StatusVariantAlsoNegotiates)
}

func ErrInsufficientStorage() *HttpError {
	return httpErrorForStatus(StatusInsufficientStorage)
}

func ErrLoopDetected() *HttpError {
	return httpErrorForStatus(StatusLoopDetected)
}

func ErrNotExtended() *HttpError {
	return httpErrorForStatus(StatusNotExtended)
}

func ErrNetworkAuthenticationRequired() *HttpError {
	return httpErrorForStatus(StatusNetworkAuthenticationRequired)
}

func httpErrorForStatus(s HttpStatus) *HttpError {
	return &HttpError{status: s, headers: headers{}}
}

func newErrorHandler(em ErrorMapper) *errorHandler {
	return &errorHandler{handlers: map[HttpStatus]ErrorResponseHandler{}, em: em}
}

// Add a custom message to the response exception. This is destructive and
// overwrites the previous message if present.
func (he *HttpError) WithMessage(message string) *HttpError {
	he.message = message
	return he
}

func (he *HttpError) WithCause(cause error) *HttpError {
	he.cause = cause
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

// Returns the underlying error that caused the application to return a 4xx or
// 5xx response. This error is not always present, such as when the server
// returns 404: Not Found.
func (erw *ErrorResponseWriter) Error() (error, bool) {
	return erw.err, erw.err != nil
}

// TODO: need to rethink this API - should it return an error??
func (erw *ErrorResponseWriter) Json(v any) error {
	return erw.rw.Json(v)
}

func (erw *ErrorResponseWriter) Text(text string) {
	erw.rw.Text(text)
}

func (erw *ErrorResponseWriter) Textf(format string, a ...any) {
	erw.rw.Textf(format, a...)
}

func (erw *ErrorResponseWriter) Raw(raw []byte) {
	erw.rw.Raw(raw)
}

func (erw *ErrorResponseWriter) RawWithContentType(raw []byte, ct ContentType) {
	erw.rw.RawWithContentType(raw, ct)
}

func (eh *errorHandler) RegisterHandler(s HttpStatus, h ErrorResponseHandler) {
	if !s.isError() {
		panic(fmt.Errorf("cannot specify an error handler for status code %d", s.code))
	}
	eh.handlers[s] = h
}

func (eh *errorHandler) HandleErrors(r any, rw *ResponseWriter, req *Request) *ResponseWriter {
	var err error
	if r != nil {
		fmt.Printf("Application code panicked: %s\n", r)
		switch e := r.(type) {
		case (*HttpError):
			rw = e.toResponse()
			err = e.cause
		case error:
			rw = eh.toHttpError(e).toResponse()
			err = e
		default:
			rw = ErrInternalServerError().toResponse()
		}
	}
	if rw.s.isError() {
		h, found := eh.handlers[rw.s]
		if !found {
			return rw
		}

		erw := &ErrorResponseWriter{rw: rw, err: err}
		h(erw, req)
	}
	return rw
}

func (e *HttpError) toResponse() *ResponseWriter {
	rw := newResponseWithStatus(e.status)
	rw.Text(e.Error())
	maps.Copy(rw.hdrs, e.headers)
	return rw
}

func (e *HttpError) isValid() bool {
	return e.status.isValid()
}

// Converts from a general error into a HttpError, is possible. Falls back to
// a 500: Internal Server Error if no match is possible.
func (eh *errorHandler) toHttpError(err error) *HttpError {
	mapped := eh.em(err)
	if mapped != nil && mapped.isValid() {
		return mapped
	}
	if errors.Is(err, fs.ErrPermission) {
		return ErrForbidden().WithCause(err)
	}
	if errors.Is(err, fs.ErrNotExist) {
		return ErrNotFound().WithCause(err)
	}
	return ErrInternalServerError().WithCause(err)
}
