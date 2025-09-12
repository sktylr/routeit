package socket

import (
	"crypto/rand"
	"crypto/rsa"
	ctls "crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"strings"
	"testing"
	"time"
)

func TestTlsSocketListenAndServe(t *testing.T) {
	conf := newTestTLSConfig()
	s := NewTlsSocket(0, conf).(*tls)

	if err := s.Bind(); err != nil {
		t.Fatalf("Bind failed: %v", err)
	}
	defer s.ln.Close()

	called := make(chan struct{})

	go s.Serve(
		func(conn net.Conn) {
			defer conn.Close()
			if _, err := conn.Write([]byte("OK")); err != nil {
				t.Errorf("failed to write to connection: %+v", err)
			}
			close(called)
		}, func(err error) {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			t.Errorf("unexpected error: %v", err)
		},
	)

	addr := s.ln.Addr().String()
	conn, err := ctls.Dial("tcp", addr, conf)
	if err != nil {
		t.Fatalf("failed to dial TLS server: %v", err)
	}
	conn.Close()

	select {
	case <-called:
	case <-time.After(time.Second):
		t.Fatal("consumer was not called in time")
	}
}

func TestTlsSocketClose(t *testing.T) {
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
			s := NewTlsSocket(0, newTestTLSConfig())
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

// newTestTLSConfig returns a minimal TLS config with an ephemeral self-signed
// cert. Suitable for tests only.
func newTestTLSConfig() *ctls.Config {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	der, _ := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	cert := ctls.Certificate{Certificate: [][]byte{der}, PrivateKey: priv}

	return &ctls.Config{Certificates: []ctls.Certificate{cert}, InsecureSkipVerify: true}
}
