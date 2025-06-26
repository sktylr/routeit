package routeit

import (
	"testing"
	"time"
)

func TestNewServerDefaults(t *testing.T) {
	srv := NewServer(ServerConfig{})

	verifyDefaultPort(t, srv.conf)
	verifyDefaultRequestSize(t, srv.conf)
	verifyDefaultReadTimeout(t, srv.conf)
	verifyDefaultWriteTimeout(t, srv.conf)
	verifyDefaultNamespace(t, srv.conf)
}

func TestNewServerOnlyPort(t *testing.T) {
	srv := NewServer(ServerConfig{Port: 3000})
	wantPort := uint16(3000)

	if srv.conf.Port != wantPort {
		t.Errorf(`custom port = %d, want %d`, srv.conf.Port, wantPort)
	}
	verifyDefaultRequestSize(t, srv.conf)
	verifyDefaultReadTimeout(t, srv.conf)
	verifyDefaultWriteTimeout(t, srv.conf)
	verifyDefaultNamespace(t, srv.conf)
}

func TestNewServerOnlyRequestBufferSize(t *testing.T) {
	srv := NewServer(ServerConfig{RequestSize: 3 * MiB})
	wantReqSize := RequestSize(1024 * 1024 * 3)

	verifyDefaultPort(t, srv.conf)
	if srv.conf.RequestSize != wantReqSize {
		t.Errorf(`custom request buffer size = %d, want %d`, srv.conf.RequestSize, wantReqSize)
	}
	verifyDefaultReadTimeout(t, srv.conf)
	verifyDefaultWriteTimeout(t, srv.conf)
	verifyDefaultNamespace(t, srv.conf)
}

func TestNewServerOnlyReadTimeout(t *testing.T) {
	srv := NewServer(ServerConfig{ReadDeadline: 3 * time.Minute})
	wantReadTmo := 3 * time.Minute

	verifyDefaultPort(t, srv.conf)
	verifyDefaultRequestSize(t, srv.conf)
	if srv.conf.ReadDeadline != wantReadTmo {
		t.Errorf(`custom read timeout = %d, want %d`, srv.conf.ReadDeadline, wantReadTmo)
	}
	verifyDefaultWriteTimeout(t, srv.conf)
	verifyDefaultNamespace(t, srv.conf)
}

func TestNewServerOnlyWriteTimeout(t *testing.T) {
	srv := NewServer(ServerConfig{WriteDeadline: 4 * time.Second})
	wantWriteTmo := 4 * time.Second

	verifyDefaultPort(t, srv.conf)
	verifyDefaultRequestSize(t, srv.conf)
	verifyDefaultReadTimeout(t, srv.conf)
	if srv.conf.WriteDeadline != wantWriteTmo {
		t.Errorf(`custom write timeout = %d, want %d`, srv.conf.WriteDeadline, wantWriteTmo)
	}
	verifyDefaultNamespace(t, srv.conf)
}

func TestNewServerOnlyNamespace(t *testing.T) {
	srv := NewServer(ServerConfig{Namespace: "/api"})

	verifyDefaultPort(t, srv.conf)
	verifyDefaultRequestSize(t, srv.conf)
	verifyDefaultReadTimeout(t, srv.conf)
	verifyDefaultWriteTimeout(t, srv.conf)
	if srv.conf.Namespace != "/api" {
		t.Errorf(`custom namespace = %q, wanted "/api"`, srv.conf.Namespace)
	}
}

func verifyDefaultPort(t *testing.T, conf ServerConfig) {
	t.Helper()
	if conf.Port != 8080 {
		t.Errorf(`default port = %d, want 8080`, conf.Port)
	}
}

func verifyDefaultRequestSize(t *testing.T, conf ServerConfig) {
	t.Helper()
	if conf.RequestSize != 1024 {
		t.Errorf(`default request buffer size = %d, want 1024`, conf.RequestSize)
	}
}

func verifyDefaultReadTimeout(t *testing.T, conf ServerConfig) {
	t.Helper()
	// 10s = 10^10 ns
	if conf.ReadDeadline != 10_000_000_000 {
		t.Errorf(`default read timeout = %d, want 10_000_000_000`, conf.ReadDeadline)
	}
}

func verifyDefaultWriteTimeout(t *testing.T, conf ServerConfig) {
	t.Helper()
	// 10s = 10^10 ns
	if conf.WriteDeadline != 10_000_000_000 {
		t.Errorf(`default write timeout = %d, want 10_000_000_000`, conf.WriteDeadline)
	}
}

func verifyDefaultNamespace(t *testing.T, conf ServerConfig) {
	t.Helper()
	// Internally the router strips the leading and any trailing slashes for
	// the global namespace, so this should be empty by default. The trie
	// structure will handle the routing beyond that.
	if conf.Namespace != "" {
		t.Errorf(`default namespace = %q, want ""`, conf.Namespace)
	}
}
