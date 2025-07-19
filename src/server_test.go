package routeit

import (
	"context"
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
			Port:          8080,
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
				name: "only port",
				in:   ServerConfig{Port: 3000},
				want: func(s serverConfig) serverConfig {
					s.Port = 3000
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

				if s.conf.Port != want.Port {
					t.Errorf(`default port = %d, want %d`, s.conf.Port, want.Port)
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

	t.Run("panics when URL Rewrite Path not a .conf file", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected NewServer to panic, did not")
			}
		}()

		NewServer(ServerConfig{URLRewritePath: "foo.bar"})
	})
}
