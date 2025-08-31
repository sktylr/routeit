package routeit

import (
	"fmt"
	"strings"
	"testing"
)

func BenchmarkCorsOriginValidation(b *testing.B) {
	for _, size := range []int{1, 10, 100, 1_000, 10_000} {
		b.Run(fmt.Sprintf("exact - %d origins", size), func(b *testing.B) {
			allowed := make([]string, size)
			for i := range size {
				allowed[i] = fmt.Sprintf("http://example%d.com", i)
			}
			cc := CorsConfig{AllowedOrigins: allowed}
			c := cc.toCors()

			nonMatch := "http://notallowed.com"

			positions := map[string]string{
				"first":  allowed[0],
				"middle": allowed[size/2],
				"last":   allowed[size-1],
			}

			for label, origin := range positions {
				b.Run(fmt.Sprintf("%s match", label), func(b *testing.B) {
					for b.Loop() {
						ok, _ := c.AllowsOrigin(&Request{}, origin)
						if !ok {
							b.Fatalf("expected %s origin match to be allowed", label)
						}
					}
				})
			}

			b.Run("non-match", func(b *testing.B) {
				for b.Loop() {
					ok, _ := c.AllowsOrigin(&Request{}, nonMatch)
					if ok {
						b.Fatalf("expected non-match origin to be disallowed")
					}
				}
			})
		})

		b.Run(fmt.Sprintf("wildcard - %d origins", size), func(b *testing.B) {
			makeOrigin := func(i int) string {
				return fmt.Sprintf("http://*.example%d.com", i)
			}
			makeMatch := func(i int) string {
				return fmt.Sprintf("http://sub.example%d.com", i)
			}

			origins := make([]string, size)
			for i := range size {
				origins[i] = makeOrigin(i)
			}
			cc := CorsConfig{AllowedOrigins: origins}
			c := cc.toCors()

			nonMatch := "http://doesnotmatch.com"

			positions := map[string]string{
				"first":  makeMatch(0),
				"middle": makeMatch(size / 2),
				"last":   makeMatch(size - 1),
			}

			for label, origin := range positions {
				b.Run(fmt.Sprintf("%s match", label), func(b *testing.B) {
					for b.Loop() {
						ok, _ := c.AllowsOrigin(&Request{}, origin)
						if !ok {
							b.Fatalf("expected wildcard %s origin to match", label)
						}
					}
				})
			}

			b.Run("non-match", func(b *testing.B) {
				for b.Loop() {
					ok, _ := c.AllowsOrigin(&Request{}, nonMatch)
					if ok {
						b.Fatalf("did not expect wildcard to match: %s", nonMatch)
					}
				}
			})
		})
	}
}

func BenchmarkCorsHeaderValidation(b *testing.B) {
	for _, size := range []int{1, 10, 100, 1_000} {
		b.Run(fmt.Sprintf("%d allowed headers", size), func(b *testing.B) {
			allowed := make([]string, size)
			for i := range size {
				allowed[i] = fmt.Sprintf("X-Header-%d", i)
			}

			cc := CorsConfig{AllowedHeaders: allowed}
			c := cc.toCors()

			headerCases := map[string]string{
				"first":  allowed[0],
				"middle": allowed[size/2],
				"last":   allowed[size-1],
			}

			for label, h := range headerCases {
				b.Run(fmt.Sprintf("match - %s", label), func(b *testing.B) {
					for b.Loop() {
						if !c.AllowedHeaders(h) {
							b.Fatalf("expected allowed header match for %s", h)
						}
					}
				})
			}

			b.Run("non match", func(b *testing.B) {
				for b.Loop() {
					if c.AllowedHeaders("X-Not-Allowed") {
						b.Fatalf("expected header to be disallowed")
					}
				}
			})

			if size >= 3 {
				start := allowed[0]
				mid := allowed[size/2]
				end := allowed[size-1]
				validHeaderCombo := fmt.Sprintf("%s, %s, %s", start, mid, end)

				b.Run("multi-match", func(b *testing.B) {
					for b.Loop() {
						if !c.AllowedHeaders(validHeaderCombo) {
							b.Fatalf("expected all headers in multi-match to be allowed")
						}
					}
				})

				// Comparison should be case insensitive and ignore leading /
				// trailing whitespace
				uppercaseTrimmed := fmt.Sprintf("   %s ,   %s , %s   ",
					strings.ToUpper(start),
					strings.ToUpper(mid),
					strings.ToUpper(end),
				)

				b.Run("multi-uppercase-trimmed", func(b *testing.B) {
					for b.Loop() {
						if !c.AllowedHeaders(uppercaseTrimmed) {
							b.Fatalf("expected uppercase + trimmed headers to be allowed")
						}
					}
				})

				uppercaseMixed := fmt.Sprintf("   %s , X-BAD-HEADER   ",
					strings.ToUpper(start),
				)

				b.Run("multi-uppercase-mixed", func(b *testing.B) {
					for b.Loop() {
						if c.AllowedHeaders(uppercaseMixed) {
							b.Fatalf("expected mixed uppercase headers to fail validation")
						}
					}
				})
			}

			if size >= 2 {
				start := allowed[0]
				invalid := "X-Invalid-Header"
				mixed := fmt.Sprintf("%s, %s", start, invalid)

				b.Run("multi-mixed", func(b *testing.B) {
					for b.Loop() {
						if c.AllowedHeaders(mixed) {
							b.Fatalf("expected mixed headers to fail validation")
						}
					}
				})
			}
		})
	}
}
