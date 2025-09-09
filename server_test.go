package routeit

import (
	"context"
	"crypto/tls"
	"log/slog"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	t.Run("sensible defaults", func(t *testing.T) {
		// Internally the router strips the leading and any trailing slashes for
		// the global namespace, so it should be empty by default. The trie
		// structure will handle the routing beyond that.
		defaultConf := serverConfig{
			HttpPort:      8080,
			RequestSize:   KiB,
			ReadDeadline:  10 * time.Second,
			WriteDeadline: 10 * time.Second,
		}
		tests := []struct {
			name string
			in   ServerConfig
			want func(s serverConfig) serverConfig
		}{
			{
				name: "defaults",
				in:   ServerConfig{},
				want: func(s serverConfig) serverConfig { return s },
			},
			{
				name: "only http port",
				in:   ServerConfig{HttpConfig: HttpConfig{HttpPort: 3000}},
				want: func(s serverConfig) serverConfig {
					s.HttpPort = 3000
					return s
				},
			},
			{
				name: "https port and TLS config",
				in: ServerConfig{
					HttpConfig: HttpConfig{
						HttpsPort: 3000,
						TlsConfig: &tls.Config{},
					},
				},
				want: func(s serverConfig) serverConfig {
					s.HttpPort = 0
					s.HttpsPort = 3000
					return s
				},
			},
			{
				name: "TLS config only",
				in:   ServerConfig{HttpConfig: HttpConfig{TlsConfig: &tls.Config{}}},
				want: func(s serverConfig) serverConfig {
					s.HttpPort = 0
					s.HttpsPort = 443
					return s
				},
			},
			{
				name: "TLS config and http port",
				in: ServerConfig{
					HttpConfig: HttpConfig{
						HttpPort:  8080,
						TlsConfig: &tls.Config{},
					},
				},
				want: func(s serverConfig) serverConfig {
					s.HttpsPort = 443
					return s
				},
			},
			{
				name: "TLS config and upgrade http to https, no http port",
				in: ServerConfig{
					HttpConfig: HttpConfig{
						TlsConfig:                &tls.Config{},
						UpgradeToHttps:           true,
						UpgradeInstructionMaxAge: time.Hour,
					},
				},
				want: func(s serverConfig) serverConfig {
					s.HttpPort = 80
					s.HttpsPort = 443
					return s
				},
			},
			{
				name: "only request buffer size",
				in:   ServerConfig{RequestSize: 3 * MiB},
				want: func(s serverConfig) serverConfig {
					s.RequestSize = 3 * MiB
					return s
				},
			},
			{
				name: "only read deadline",
				in:   ServerConfig{ReadDeadline: 3 * time.Minute},
				want: func(s serverConfig) serverConfig {
					s.ReadDeadline = 3 * time.Minute
					return s
				},
			},
			{
				name: "only write deadline",
				in:   ServerConfig{WriteDeadline: 3 * time.Minute},
				want: func(s serverConfig) serverConfig {
					s.WriteDeadline = 3 * time.Minute
					return s
				},
			},
			{
				name: "only namespace",
				in:   ServerConfig{Namespace: "/api"},
				want: func(s serverConfig) serverConfig {
					s.Namespace = "/api"
					return s
				},
			},
			{
				name: "only debug",
				in:   ServerConfig{Debug: true},
				want: func(s serverConfig) serverConfig {
					s.Debug = true
					return s
				},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				s := NewServer(tc.in)
				want := tc.want(defaultConf)

				if s.conf.HttpPort != want.HttpPort {
					t.Errorf(`default http port = %d, want %d`, s.conf.HttpPort, want.HttpPort)
				}
				if s.conf.HttpsPort != want.HttpsPort {
					t.Errorf(`default https port = %d, want %d`, s.conf.HttpsPort, want.HttpsPort)
				}
				if s.conf.RequestSize != want.RequestSize {
					t.Errorf(`default request buffer size = %d, want %d`, s.conf.RequestSize, want.RequestSize)
				}
				if s.conf.ReadDeadline != want.ReadDeadline {
					t.Errorf(`default read timeout = %d, want %d`, s.conf.ReadDeadline, want.ReadDeadline)
				}
				if s.conf.WriteDeadline != want.WriteDeadline {
					t.Errorf(`default write timeout = %d, want %d`, s.conf.WriteDeadline, want.WriteDeadline)
				}
				if s.conf.Namespace != want.Namespace {
					t.Errorf(`default namespace = %#q, want %#q`, s.conf.Namespace, want.Namespace)
				}
				if s.conf.Debug != want.Debug {
					t.Errorf("Debug = %t, wanted %t", s.conf.Debug, want.Debug)
				}
				if want.Debug && !s.log.Enabled(context.Background(), slog.LevelDebug) {
					t.Error("DEBUG logging not enabled but wanted Debug to be on")
				}
			})
		}
	})

	t.Run("panics", func(t *testing.T) {
		tests := []struct {
			name string
			conf ServerConfig
		}{
			{
				name: "URL rewrite path not a .conf file",
				conf: ServerConfig{URLRewritePath: "foo.bar"},
			},
			{
				name: "https port provided but no TLS config",
				conf: ServerConfig{HttpConfig: HttpConfig{HttpsPort: 443}},
			},
			{
				name: "upgrade to HTTPS but not TLS config",
				conf: ServerConfig{HttpConfig: HttpConfig{UpgradeToHttps: true}},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				defer func() {
					if r := recover(); r == nil {
						t.Error("expected NewServer to panic, did not")
					}
				}()

				NewServer(tc.conf)
			})
		}
	})
}

func TestAtomicity(t *testing.T) {
	expectPanic := func(t *testing.T) {
		if r := recover(); r == nil {
			t.Error("expected panic, none came")
		}
	}
	srv := NewServer(ServerConfig{})
	srv.started.Store(true)

	t.Run("RegisterRoutes", func(t *testing.T) {
		defer expectPanic(t)
		srv.RegisterRoutes(RouteRegistry{
			"/hello": Get(func(rw *ResponseWriter, req *Request) error { return nil }),
		})
	})

	t.Run("RegisterRoutesUnderNamespace", func(t *testing.T) {
		defer expectPanic(t)
		srv.RegisterRoutesUnderNamespace("/api", RouteRegistry{
			"/hello": Get(func(rw *ResponseWriter, req *Request) error { return nil }),
		})
	})

	t.Run("RegisterMiddleware", func(t *testing.T) {
		defer expectPanic(t)
		srv.RegisterMiddleware(func(c Chain, rw *ResponseWriter, req *Request) error { return nil })
	})

	t.Run("StartOrPanic", func(t *testing.T) {
		defer expectPanic(t)
		srv.StartOrPanic()
	})

	t.Run("Start", func(t *testing.T) {
		err := srv.Start()
		if err == nil {
			t.Error("expected Start() to return an error when already started")
		}
		if err.Error() != "server has already been started" {
			t.Errorf(`Error() = %#q, wanted ""server has already been started"`, err.Error())
		}
	})
}
