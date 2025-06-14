package routeit

import "net"

type ResponseWriter struct{ conn net.Conn }

func (rw *ResponseWriter) Write(msg string) error {
	_, err := rw.conn.Write([]byte(msg))
	return err
}
