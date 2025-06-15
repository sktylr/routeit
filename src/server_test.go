package routeit

import "testing"

func TestNewServerDefaults(t *testing.T) {
	srv := NewServer(ServerConfig{})

	if srv.conf.Port != 8080 {
		t.Errorf(`default port = %q, want 8080`, srv.conf.Port)
	}
	if srv.conf.RequestSize != 1024 {
		t.Errorf(`default request buffer size = %q, want 1024`, srv.conf.RequestSize)
	}
}

func TestNewServerOnlyPort(t *testing.T) {
	srv := NewServer(ServerConfig{Port: 3000})
	wantPort := 3000

	if srv.conf.Port != wantPort {
		t.Errorf(`custom port = %q, want %#q`, srv.conf.Port, wantPort)
	}
	if srv.conf.RequestSize != 1024 {
		t.Errorf(`default request buffer size = %q, want 1024`, srv.conf.RequestSize)
	}
}

func TestNewServerOnlyRequestBufferSize(t *testing.T) {
	srv := NewServer(ServerConfig{RequestSize: 3 * MiB})
	wantReqSize := RequestSize(1024 * 1024 * 3)

	if srv.conf.Port != 8080 {
		t.Errorf(`default port = %q, want 8080`, srv.conf.Port)
	}
	if srv.conf.RequestSize != wantReqSize {
		t.Errorf(`custom request buffer size = %q, want %#q`, srv.conf.RequestSize, wantReqSize)
	}
}
