package routeit

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

type ResponseWriter struct {
	// TODO: may be able to avoid having this
	conn net.Conn
	b    []byte
	s    HttpStatus
}

func (rw *ResponseWriter) Json(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	rw.b = b
	return nil
}

func (rw *ResponseWriter) Status(s HttpStatus) {
	rw.s = s
}

func (rw *ResponseWriter) write() error {
	var sb strings.Builder

	// HTTP line
	sb.WriteString(fmt.Sprintf("HTTP/1.1 %d %s\n", rw.s.code, rw.s.msg))

	// Headers
	sb.WriteString("Server: routeit\n")
	// TODO: should come from the response
	sb.WriteString("Content-Type: application/json\n")
	sb.WriteString("Cache-Control: no-cache\n")
	sb.WriteString(fmt.Sprintf("Content-Length: %d\n", len(rw.b)))

	// Blank line between headers and the response
	sb.WriteString("\n")

	sb.Write(rw.b)

	_, err := rw.conn.Write([]byte(sb.String()))
	return err
}
