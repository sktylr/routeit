package routeit

import (
	"fmt"
	"strings"
	"testing"

	"github.com/sktylr/routeit/internal/headers"
)

// Remember to use -benchtime=0.01s or similar to avoid the benchmarking
// hanging indefinitely 😬
func BenchmarkHostValidationMiddleware(b *testing.B) {
	for _, size := range []int{1, 10, 100, 1000} {
		b.Run(fmt.Sprintf("%d allowed hosts", size), func(b *testing.B) {
			exact := make([]string, size)
			subdomains := make([]string, size)
			for i := range size {
				exact[i] = fmt.Sprintf("host%d.example.com", i)
				subdomains[i] = fmt.Sprintf(".sub%d.example.com", i)
			}

			testCases := []struct {
				label    string
				allowed  []string
				testHost string
				wantErr  bool
			}{
				{"exact - first", exact, exact[0], false},
				{"exact - middle", exact, exact[size/2], false},
				{"exact - last", exact, exact[size-1], false},
				{"exact miss", exact, "notallowed.com", true},

				{"subdomain - first", subdomains, "api.sub0.example.com", false},
				{"subdomain - middle", subdomains, fmt.Sprintf("api.sub%d.example.com", size/2), false},
				{"subdomain - last", subdomains, fmt.Sprintf("api.sub%d.example.com", size-1), false},
				{"subdomain - miss", subdomains, "something.unmatched.com", true},
				{"subdomain - miss, too many subdomain levels", subdomains, fmt.Sprintf("site.web.sub%d.example.com", size-1), true},

				{"duplicate - exact", append([]string{exact[0]}, exact...), exact[0], false},
				{"duplicate - subdomain", append([]string{subdomains[0]}, subdomains...), fmt.Sprintf("api.%s", strings.TrimPrefix(subdomains[0], ".")), false},
			}

			for _, tc := range testCases {
				b.Run(tc.label, func(b *testing.B) {
					mw := newMiddleware()
					mw.Register(hostValidationMiddleware(tc.allowed))

					b.ResetTimer()
					for b.Loop() {
						b.StopTimer()
						headers := headers.Headers{}
						headers.Set("Host", tc.testHost)
						req := &Request{headers: &RequestHeaders{headers}}
						rw := &ResponseWriter{}
						chain := mw.NewChain(func(rw *ResponseWriter, req *Request) error {
							return nil
						})
						b.StartTimer()

						err := chain.Proceed(rw, req)

						if tc.wantErr && err == nil {
							b.Fatalf("expected error for host %q", tc.testHost)
						}
						if !tc.wantErr && err != nil {
							b.Fatalf("unexpected error for host %q: %v", tc.testHost, err)
						}
					}
				})
			}
		})
	}
}

func TestHostValidationMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		allowedHosts  []string
		hostHeader    string
		wantProceeded bool
		wantErr       bool
		wantHost      string
	}{
		{
			name:         "rejects when no Host header is present",
			allowedHosts: []string{"example.com"},
			wantErr:      true,
		},
		{
			name:         "rejects unmatched Host header",
			allowedHosts: []string{"example.com"},
			hostHeader:   "bad.com",
			wantErr:      true,
		},
		{
			name:          "accepts exact match host",
			allowedHosts:  []string{"example.com"},
			hostHeader:    "example.com",
			wantProceeded: true,
			wantHost:      "example.com",
		},
		{
			name:          "accepts wildcard subdomain match",
			allowedHosts:  []string{".example.com"},
			hostHeader:    "api.example.com",
			wantProceeded: true,
			wantHost:      "api.example.com",
		},
		{
			name:         "rejects deep subdomain for wildcard",
			allowedHosts: []string{".example.com"},
			hostHeader:   "deep.api.example.com",
			wantErr:      true,
		},
		{
			name:          "strips port from host before matching",
			allowedHosts:  []string{"example.com"},
			hostHeader:    "example.com:8080",
			wantProceeded: true,
			wantHost:      "example.com",
		},
		{
			name:          "rejects malformed ports",
			allowedHosts:  []string{"example.com"},
			hostHeader:    "example.com:8080A",
			wantProceeded: false,
			wantErr:       true,
		},
		{
			name:         "rejects if allowed host list is empty",
			allowedHosts: []string{},
			hostHeader:   "example.com",
			wantErr:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := NewTestRequest(t, "/path", GET, TestRequestOptions{
				Headers: func() []string {
					if tc.hostHeader == "" {
						return nil
					}
					return []string{"Host", tc.hostHeader}
				}(),
			})

			_, proceeded, err := TestMiddleware(hostValidationMiddleware(tc.allowedHosts), req)

			if proceeded != tc.wantProceeded {
				t.Errorf("proceeded = %v, want %v", proceeded, tc.wantProceeded)
			}
			if (err != nil) != tc.wantErr {
				t.Errorf("error = %v, want error? %v", err, tc.wantErr)
			}
			if tc.wantHost != "" && req.req.Host() != tc.wantHost {
				t.Errorf("host = %q, want %q", req.req.Host(), tc.wantHost)
			}
		})
	}
}
