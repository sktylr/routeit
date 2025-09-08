package socket

import ctls "crypto/tls"

// A [combined] socket is one that holds two TCP connections. The first serves
// and receives messages directly over TCP, while the second uses TLS over TCP
// for increased security. It is typically used when we would want to listen
// for HTTP and HTTPS messages, as the former will travel directly over TCP,
// while the latter will be transported via TLS over TCP.
type combined struct {
	tcp Socket
	tls Socket
}

// Use [NewCombinedSocket] to listen over two separate ports, listening over
// TCP on the first and TLS over TCP on the second. A valid, non-nil
// [ctls.Config] must be passed and the ports must be different, otherwise
// calling [Socket.Bind] will fail.
func NewCombinedSocket(tcpPort, tlsPort uint16, conf *ctls.Config) Socket {
	return &combined{
		tcp: NewTcpSocket(tcpPort),
		tls: NewTlsSocket(tlsPort, conf),
	}
}

func (c *combined) Bind() error {
	if err := c.tcp.Bind(); err != nil {
		return err
	}
	if err := c.tls.Bind(); err != nil {
		return err
	}
	return nil
}

func (c *combined) Serve(onConn onConnection, onErr onError) {
	ch := make(chan struct{}, 2)

	go func() {
		c.tcp.Serve(onConn, onErr)
		ch <- struct{}{}
	}()
	go func() {
		c.tls.Serve(onConn, onErr)
		ch <- struct{}{}
	}()

	// The individual [Socket.Serve] implementations for TCP and TLS are
	// blocking. Since we spin up two separate go-routines for this here, we
	// end up making this implementation non-blocking. For consistency, we use
	// channels within the go-routines so that if both underlying sockets
	// become closed (i.e. using the [Socket.Close] method), then we will exit
	// this method, otherwise we will block as is done on the underlying
	// sockets.
	<-ch
	<-ch
}

func (c *combined) Close() error {
	if err := c.tcp.Close(); err != nil {
		return err
	}
	if err := c.tls.Close(); err != nil {
		return err
	}
	return nil
}
