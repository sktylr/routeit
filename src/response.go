package routeit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type ResponseWriter struct {
	bdy  []byte
	s    HttpStatus
	hdrs headers
}

// Sets a sensible default for the status code of the response depending on the
// request method. For example, most POST requests return a 201: Created, while
// many DELETE and OPTIONS requests return 204: No Content. This can be
// overwritten by the integrator using [ResponseWriter.Status]
func newResponseForMethod(method HttpMethod) *ResponseWriter {
	switch method {
	case POST:
		return newResponseWithStatus(StatusCreated)
	case OPTIONS:
		return newResponseWithStatus(StatusNoContent)
	default:
		return newResponseWithStatus(StatusOK)
	}
}

func newResponseWithStatus(status HttpStatus) *ResponseWriter {
	headers := newResponseHeaders()
	return &ResponseWriter{s: status, hdrs: headers}
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
	rw.hdrs.Set("Content-Length", fmt.Sprintf("%d", len(raw)))
	rw.hdrs.Set("Content-Type", ct.string())
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

// Sets a header with the corresponding value. This is destructive, meaning
// repeated calls using the same key will preserve the last key. Header key and
// values will be sanitised per HTTP spec before being added to the server's
// response.
func (rw *ResponseWriter) Header(key string, val string) {
	// TODO: errors should use this!
	// TODO: probably want to define some allow list (e.g. to avoid overwriting Content-Length etc.)
	rw.hdrs.Set(key, val)
}

func (rw *ResponseWriter) write() []byte {
	var sb strings.Builder

	// Status line
	sb.WriteString(fmt.Sprintf("HTTP/1.1 %d %s\r\n", rw.s.code, rw.s.msg))

	// Headers
	// TODO: we should probs set the content length header here to avoid it being overwritten
	now := time.Now().UTC()
	rw.hdrs.Set("Date", now.Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	rw.hdrs.WriteTo(&sb)

	// CRLF between headers and the response
	sb.WriteString("\r\n")

	sb.Write(rw.bdy)
	return []byte(sb.String())
}
