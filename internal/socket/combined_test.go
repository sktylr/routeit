package socket

import (
	ctls "crypto/tls"
	"net"
	"strings"
	"testing"
	"time"
)

func TestCombinedSocketListenAndServe(t *testing.T) {
	conf := newTestTLSConfig()
	c := NewCombinedSocket(0, 0, conf).(*combined)

	if err := c.Bind(); err != nil {
		t.Fatalf("Bind failed: %v", err)
	}
	defer c.Close()

	tcpCalled := make(chan struct{})
	tlsCalled := make(chan struct{})

	go c.Serve(
		func(conn net.Conn) {
			defer conn.Close()
			if _, ok := conn.(*ctls.Conn); ok {
				close(tlsCalled)
			} else {
				close(tcpCalled)
			}
			if _, err := conn.Write([]byte("OK")); err != nil {
				t.Errorf("failed to write to connection: %+v", err)
			}
		},
		func(err error) {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			t.Errorf("unexpected error: %v", err)
		},
	)

	tcpAddr := c.tcp.(*tcp).ln.Addr().String()
	tcpConn, err := net.Dial("tcp", tcpAddr)
	if err != nil {
		t.Fatalf("failed to dial TCP server: %v", err)
	}
	tcpConn.Close()

	tlsAddr := c.tls.(*tls).ln.Addr().String()
	clientConf := conf.Clone()
	clientConf.InsecureSkipVerify = true
	tlsConn, err := ctls.Dial("tcp", tlsAddr, clientConf)
	if err != nil {
		t.Fatalf("failed to dial TLS server: %v", err)
	}
	tlsConn.Close()

	select {
	case <-tcpCalled:
	case <-time.After(time.Second):
		t.Fatal("TCP consumer was not called in time")
	}

	select {
	case <-tlsCalled:
	case <-time.After(time.Second):
		t.Fatal("TLS consumer was not called in time")
	}
}

func TestCombinedSocketClose(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(Socket)
		wantErr bool
	}{
		{
			name: "success",
			setup: func(s Socket) {
				if err := s.Bind(); err != nil {
					t.Fatalf("unexpected error while binding: %+v", err)
				}
			},
		},
		{
			name:    "socket not bound",
			setup:   func(s Socket) {},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := NewCombinedSocket(0, 0, newTestTLSConfig())
			tc.setup(s)

			err := s.Close()

			if err != nil && !tc.wantErr {
				t.Errorf("unexpected error: %+v", err)
			}
			if err == nil && tc.wantErr {
				t.Error("wanted error but not found")
			}
		})
	}
}
