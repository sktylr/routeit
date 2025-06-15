package routeit

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ResponseWriter struct {
	bdy  []byte
	s    HttpStatus
	hdrs headers
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

func (rw *ResponseWriter) Status(s HttpStatus) {
	rw.s = s
}

func (rw *ResponseWriter) write() []byte {
	var sb strings.Builder

	// HTTP line
	sb.WriteString(fmt.Sprintf("HTTP/1.1 %d %s\n", rw.s.code, rw.s.msg))

	// Headers
	rw.hdrs.writeTo(&sb)

	// Blank line between headers and the response
	sb.WriteString("\n")

	sb.Write(rw.bdy)
	return []byte(sb.String())
}
