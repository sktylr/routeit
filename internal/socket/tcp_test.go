package socket

import (
	"net"
	"strings"
	"testing"
	"time"
)

func TestTcpSocketListenAndServe(t *testing.T) {
	// Using port 0 here means we let the kernel choose the port, reducing
	// flakiness as it is incredibly unlikely that there are no ports available.
	s := NewTcpSocket(0).(*tcp)
	if err := s.Bind(); err != nil {
		t.Fatalf("Bind failed: %v", err)
	}
	defer s.ln.Close()

	called := make(chan struct{})

	go s.Serve(
		func(conn net.Conn) {
			defer conn.Close()
			close(called)
		},
		func(err error) {
			if strings.Contains(err.Error(), "use of closed network connection") {
				// This is shutdown noise and is returned by the net library.
				// We can ignore this type of error as it is due to the fact
				// that the TCP socket will loop indefinitely, until we close
				// it.
				return
			}
			t.Errorf("unexpected error: %v", err)
		},
	)

	addr := s.ln.Addr().String()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	conn.Close()

	select {
	case <-called:
	case <-time.After(time.Second):
		t.Fatal("consumer was not called in time")
	}
}

func TestTcpSocketClose(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(Socket)
		wantErr bool
	}{
		{
			name: "success",
			setup: func(s Socket) {
				if err := s.Bind(); err != nil {
					t.Fatalf("unexpected error while binging: %+v", err)
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
			s := NewTcpSocket(0)
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
