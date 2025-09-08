package socket

import (
	"fmt"
	"net"
	"sync/atomic"
)

type tcp struct {
	port   uint16
	ln     net.Listener
	closed atomic.Bool
}

// Creates a new socket that operates over TCP on the specified port.
func NewTcpSocket(port uint16) Socket {
	return &tcp{port: port}
}

func (t *tcp) Bind() error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", t.port))
	if err != nil {
		return err
	}
	t.ln = ln
	t.closed.Store(false)
	return nil
}

func (t *tcp) Serve(onConn onConnection, onErr onError) {
	defer t.Close()

	for {
		if t.closed.Load() {
			return
		}

		conn, err := t.ln.Accept()
		if err != nil {
			go onErr(err)
			continue
		}
		go onConn(conn)
	}
}

func (t *tcp) Close() error {
	if t.ln == nil {
		return ErrSocketNotInUse
	}
	t.closed.Store(true)
	return t.ln.Close()
}
