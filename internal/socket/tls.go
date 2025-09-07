package socket

import (
	ctls "crypto/tls"
	"fmt"
	"net"
)

type tls struct {
	port uint16
	conf *ctls.Config
	ln   net.Listener
}

// Creates a new socket that uses TLS over TCP with the given config and port.
func NewTlsSocket(port uint16, conf *ctls.Config) Socket {
	return &tls{port: port, conf: conf.Clone()}
}

func (t *tls) Bind() error {
	ln, err := ctls.Listen("tcp", fmt.Sprintf(":%d", t.port), t.conf)
	if err != nil {
		return err
	}
	t.ln = ln
	return nil
}

func (t *tls) Serve(onConn onConnection, onErr onError) {
	defer t.Close()

	for {
		conn, err := t.ln.Accept()
		if err != nil {
			go onErr(err)
			continue
		}
		go onConn(conn)
	}
}

func (t *tls) Close() error {
	if t.ln == nil {
		return ErrSocketNotInUse
	}
	return t.ln.Close()
}
