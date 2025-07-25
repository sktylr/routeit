package routeit

import (
	"fmt"
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
