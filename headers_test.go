package routeit

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestHeadersWriteTo(t *testing.T) {
	tests := []struct {
		name string
		in   map[string][]string
		want string
	}{
		{
			name: "multiple multi-case headers",
			in: map[string][]string{
				"Content-Type":   {"application/json"},
				"date":           {"Fri, 13 Jun 2025 19:35:42 GMT"},
				"Content-length": {"85"},
			},
			want: "Content-Type: application/json\r\ndate: Fri, 13 Jun 2025 19:35:42 GMT\r\nContent-length: 85\r\n",
		},
		{
			name: "empty",
		},
		{
			name: "clears new lines",
			in:   map[string][]string{"Content-Length": {"1\n5\n"}},
			want: "Content-Length: 15\r\n",
		},
		{
			name: "clears internal carriage return",
			in:   map[string][]string{"Content-Type": {"application/\r\njson"}},
			want: "Content-Type: application/json\r\n",
		},
		{
			name: "does not clear internal whitespace",
			in:   map[string][]string{"Content-Type": {"application/json; charset=utf-8"}},
			want: "Content-Type: application/json; charset=utf-8\r\n",
		},
		{
			name: "allows HTAB",
			in:   map[string][]string{"Content-Type": {"application/json\tcharset=utf-8"}},
			want: "Content-Type: application/json\tcharset=utf-8\r\n",
		},
		{
			name: "handles multiple header values",
			in:   map[string][]string{"Allow": {"GET", "PUT", "POST"}},
			want: "Allow: GET\r\nAllow: PUT\r\nAllow: POST\r\n",
		},
		{
			name: "handles complex multiple header values",
			in:   map[string][]string{"Allow": {"GET,HEAD", "PUT,PATCH", "POST,DELETE,OPTIONS"}},
			want: "Allow: GET,HEAD\r\nAllow: PUT,PATCH\r\nAllow: POST,DELETE,OPTIONS\r\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := headers{}
			for k, vals := range tc.in {
				for _, v := range vals {
					h.Append(k, v)
				}
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

				verifyPresentAndMatches(t, h, "Content-Length", []string{"16"})
				if len(h) != 1 {
					t.Errorf(`len(h) = %d, wanted only 1 element`, len(h))
				}
			})
		}
	})

	t.Run("sanitises", func(t *testing.T) {
		h := headers{}

		h.Set("Content\r\n-Length", "16\n\n\t")
		want := []string{"16\t"}

		verifyPresentAndMatches(t, h, "Content-Length", want)
	})
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
				verifyPresentAndMatches(t, base.headers, tc, []string{"val"})
			})
		}
	})
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
			headers{"content-length": headerVal{[]string{"abc"}, "Content-Length"}},
			0,
		},
		{
			"valid",
			headers{"content-length": headerVal{[]string{"85"}, "Content-Length"}},
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

func verifyPresentAndMatches(t *testing.T, h headers, key string, want []string) {
	t.Helper()
	got, exists := h.All(key)
	if !exists {
		t.Errorf("wanted %q to be present", key)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf(`h[%q] = %v, want %v`, key, got, want)
	}
}
