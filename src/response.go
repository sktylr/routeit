package routeit

import (
	"encoding/json"
	"fmt"
	"net"
)

type ResponseWriter struct {
	conn net.Conn
	b    []byte
}

func (rw *ResponseWriter) Json(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	rw.b = b
	return nil
}

func (rw *ResponseWriter) write() error {
	_, err := rw.conn.Write([]byte("HTTP/1.1 200 Ok\nServer: routeit\nContent-Type: application/json\nCache-Control: no-cache\n"))
	if err != nil {
		return err
	}
	_, err = rw.conn.Write([]byte(fmt.Sprintf("Content-Length: %d\n\n", len(rw.b))))
	if err != nil {
		return err
	}
	_, err = rw.conn.Write(rw.b)
	return err
}
