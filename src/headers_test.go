package routeit

import (
	"sort"
	"strings"
	"testing"
)

func TestHeadersWriteTo(t *testing.T) {
	tests := []struct {
		name string
		in   map[string]string
		want string
	}{
		{
			"multiple multi-case headers",
			map[string]string{
				"Content-Type":   "application/json",
				"date":           "Fri, 13 Jun 2025 19:35:42 GMT",
				"Content-length": "85",
			},
			"Content-Type: application/json\r\ndate: Fri, 13 Jun 2025 19:35:42 GMT\r\nContent-length: 85\r\n",
		},
		{
			"empty",
			map[string]string{},
			"",
		},
		{
			"clears new lines",
			map[string]string{"Content-Length": "1\n5\n"},
			"Content-Length: 15\r\n",
		},
		{
			"clears internal carriage return",
			map[string]string{"Content-Type": "application/\r\njson"},
			"Content-Type: application/json\r\n",
		},
		{
			"does not clear internal whitespace",
			map[string]string{"Content-Type": "application/json; charset=utf-8"},
			"Content-Type: application/json; charset=utf-8\r\n",
		},
		{
			"allows HTAB",
			map[string]string{"Content-Type": "application/json\tcharset=utf-8"},
			"Content-Type: application/json\tcharset=utf-8\r\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := headers{}
			for k, v := range tc.in {
				h.Set(k, v)
			}

			var sb strings.Builder
			h.WriteTo(&sb)
			actual := sb.String()

			if len(actual) != len(tc.want) {
				t.Errorf(`length = %d, want %d`, len(actual), len(tc.want))
			}

			actualLines := make([]string, 0, len(actual))
			for l := range strings.Lines(actual) {
				actualLines = append(actualLines, l)
			}
			sort.Strings(actualLines)

			wantLines := make([]string, 0, len(tc.want))
			for l := range strings.Lines(tc.want) {
				wantLines = append(wantLines, l)
			}
			sort.Strings(wantLines)

			for i, l := range actualLines {
				wl := wantLines[i]
				if l != wl {
					t.Errorf(`got %q, want %#q`, l, wl)
				}
			}
		})
	}
}

func TestNewResponseHeaders(t *testing.T) {
	h := newResponseHeaders()

	if len(h) != 1 {
		t.Errorf(`len(h) = %q, want match for %#q`, len(h), 1)
	}
	verifyPresentAndMatches(t, h, "Server", "routeit")
}

func TestHeadersSet(t *testing.T) {
	t.Run("inserts cases insensitive", func(t *testing.T) {
		keys := []string{
			"Content-Length",
			"content-length",
			"content-Length",
			"cOntent-LEngTh",
			"Content-length",
		}

		for _, k := range keys {
			t.Run(k, func(t *testing.T) {
				h := headers{}
				h.Set("Content-Length", "15")

				h.Set(k, "16")

				verifyPresentAndMatches(t, h, "Content-Length", "16")
				if len(h) != 1 {
					t.Errorf(`len(h) = %d, wanted only 1 element`, len(h))
				}
			})
		}
	})

	t.Run("sanitises", func(t *testing.T) {
		h := headers{}

		h.Set("Content\r\n-Length", "16\n\n\t")
		want := "16\t"

		verifyPresentAndMatches(t, h, "Content-Length", want)
	})
}

func TestHeadersFromRaw(t *testing.T) {
	tests := []struct {
		name string
		raw  [][]byte
		want map[string]string
	}{
		{
			"multi header",
			[][]byte{[]byte("Host: localhost"), []byte("Content-Type: application/json")},
			map[string]string{
				"Host":         "localhost",
				"Content-Type": "application/json",
			},
		},
		{
			"exits after empty lines",
			[][]byte{[]byte(""), []byte("Host: localhost")},
			map[string]string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h, err := headersFromRaw(tc.raw)

			if err != nil {
				t.Errorf("expected error to be nil: %v", err)
			}
			if len(h) != len(tc.want) {
				t.Errorf(`headers from raw len(h) = %d, want %d`, len(h), len(tc.want))
			}
			for k, v := range tc.want {
				verifyPresentAndMatches(t, h, k, v)
			}
		})
	}
}

func TestContentLength(t *testing.T) {
	tests := []struct {
		name string
		in   headers
		want uint
	}{
		{
			"not present",
			headers{},
			0,
		},
		{
			"not parsable",
			headers{"content-length": headerVal{"abc", "Content-Length"}},
			0,
		},
		{
			"valid",
			headers{"content-length": headerVal{"85", "Content-Length"}},
			85,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cLen := tc.in.ContentLength()
			if cLen != tc.want {
				t.Errorf(`h.ContentLength() = %d, wanted %d`, cLen, tc.want)
			}
		})
	}
}

func TestHeadersGet(t *testing.T) {
	t.Run("case insensitive", func(t *testing.T) {
		base := newResponseHeaders()
		base.Set("Key", "val")
		tests := []string{
			"key",
			"Key",
			"kEy",
			"keY",
			"KEy",
			"KeY",
			"kEY",
			"KEY",
		}

		for _, tc := range tests {
			t.Run(tc, func(t *testing.T) {
				verifyPresentAndMatches(t, base, tc, "val")
			})
		}
	})
}

func verifyPresentAndMatches(t *testing.T, h headers, key string, want string) {
	t.Helper()
	got, exists := h.Get(key)
	if !exists {
		t.Errorf("wanted %q to be present", key)
	}
	if got != want {
		t.Errorf(`h[%q] = %q, want %#q`, key, got, want)
	}
}
