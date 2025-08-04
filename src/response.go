package routeit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ResponseWriter struct {
	bdy     []byte
	s       HttpStatus
	headers *ResponseHeaders
	ct      ContentType
}

// Sets a sensible default for the status code of the response depending on the
// request method. For example, most POST requests return a 201: Created, while
// many DELETE and OPTIONS requests return 204: No Content. This can be
// overwritten by the integrator using [ResponseWriter.Status]
func newResponseForMethod(method HttpMethod) *ResponseWriter {
	switch method {
	case POST:
		return newResponseWithStatus(StatusCreated)
	case DELETE, OPTIONS:
		return newResponseWithStatus(StatusNoContent)
	default:
		return newResponseWithStatus(StatusOK)
	}
}

func newResponseWithStatus(status HttpStatus) *ResponseWriter {
	rw := newResponse()
	rw.s = status
	return rw
}

func newResponse() *ResponseWriter {
	// TODO: should use constructor here!
	return &ResponseWriter{headers: newResponseHeaders()}
}

// Adds a JSON response body to the response and sets the corresponding
// Content-Length and Content-Type headers. This is a destructive operation,
// meaning repeated calls to Json(...) only preserve the last invocation.
func (rw *ResponseWriter) Json(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	rw.RawWithContentType(b, CTApplicationJson)
	return nil
}

// Adds a plaintext response body to the response and sets the corresponding
// Content-Length and Content-Type headers. This is a destructive operation,
// meaning repeated calls to Text(...) only preserve the last invocation.
func (rw *ResponseWriter) Text(text string) {
	rw.RawWithContentType([]byte(text), CTTextPlain)
}

// Shorthand for the Text function using a format string.
func (rw *ResponseWriter) Textf(format string, a ...any) {
	text := fmt.Sprintf(format, a...)
	rw.Text(text)
}

// Adds a raw response body to the response and sets the corresponding
// Content-Length and Content-Type headers. This is a destructive operation,
// meaning repeated calls to Raw(...) only preserve the last invocation. The
// mimetype of the body is inferred from its content.
func (rw *ResponseWriter) Raw(raw []byte) {
	cType := http.DetectContentType(raw)
	rw.RawWithContentType(raw, parseContentType(cType))
}

// Destructively sets the body of the response and updates headers accordingly
func (rw *ResponseWriter) RawWithContentType(raw []byte, ct ContentType) {
	rw.bdy = raw
	rw.headers.Set("Content-Length", fmt.Sprintf("%d", len(raw)))
	rw.headers.Set("Content-Type", ct.string())
	rw.ct = ct
}

// Sets the status of the response. The server sets an opinionated default
// depending on the incoming request. POST requests default to 201: Created and
// DELETE requests default to 204: No Content. All other request methods
// default to 200: OK.
func (rw *ResponseWriter) Status(s HttpStatus) {
	if !s.isValid() {
		panic(fmt.Errorf("invalid HTTP status code: %d, %s", s.code, s.msg))
	}
	rw.s = s
}

// TODO:
// Sets a header with the corresponding value. This is destructive, meaning
// repeated calls using the same key will preserve the last key. Header key and
// values will be sanitised per HTTP spec before being added to the server's
// response. It is the user's responsibility to ensure that the headers are
// safe and non-conflicting. For example, it is heavily discouraged to modify
// the Content-Type or Content-Length headers as they are managed implicitly
// whenever a body is written to a response and can cause issues on the client
// if they contain incorrect values.
// func (rw *ResponseWriter) Header(key string, val string) {
// 	rw.hdrs.Set(key, val)
// }

// TODO:
func (rw *ResponseWriter) Headers() *ResponseHeaders {
	return rw.headers
}

func (rw *ResponseWriter) clear() {
	rw.bdy = []byte{}
	delete(rw.headers.headers, "content-type")
	delete(rw.headers.headers, "content-length")
}

func (rw *ResponseWriter) write() []byte {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("HTTP/1.1 %d %s\r\n", rw.s.code, rw.s.msg))
	now := time.Now().UTC()
	rw.headers.Set("Date", now.Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	rw.headers.headers.WriteTo(&buf)
	buf.WriteString("\r\n")
	buf.Write(rw.bdy)
	return buf.Bytes()
}
