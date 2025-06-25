package routeit

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type ResponseWriter struct {
	bdy  []byte
	s    HttpStatus
	hdrs headers
}

func newResponse(status HttpStatus) *ResponseWriter {
	headers := newResponseHeaders()
	return &ResponseWriter{s: status, hdrs: headers}
}

func (rw *ResponseWriter) Json(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	rw.bdy = b
	rw.hdrs["Content-Length"] = fmt.Sprintf("%d", len(b))
	rw.hdrs["Content-Type"] = "application/json"
	return nil
}

func (rw *ResponseWriter) Text(text string) {
	rw.body([]byte(text), "text/plain")
}

func (rw *ResponseWriter) Status(s HttpStatus) {
	rw.s = s
}

// Destructively sets the body of the response and updates headers accordingly
func (rw *ResponseWriter) body(raw []byte, contentType string) {
	rw.bdy = raw
	rw.hdrs["Content-Length"] = fmt.Sprintf("%d", len(raw))
	// TODO: should define an enum type thing for this!
	rw.hdrs["Content-Type"] = contentType
}

func (rw *ResponseWriter) write() []byte {
	var sb strings.Builder

	// Status line
	sb.WriteString(fmt.Sprintf("HTTP/1.1 %d %s\r\n", rw.s.code, rw.s.msg))

	// Headers
	// TODO: we should probs set the content length header here to avoid it being overwritten
	now := time.Now().UTC()
	rw.hdrs["Date"] = now.Format("Mon, 02 Jan 2006 15:04:05 GMT")
	rw.hdrs.writeTo(&sb)

	// CRLF between headers and the response
	sb.WriteString("\r\n")

	sb.Write(rw.bdy)
	return []byte(sb.String())
}
