package routeit

import (
	"sort"
	"strings"
	"testing"
)

// TODO: can simplify these write to tests!
func TestHeadersWriteTo(t *testing.T) {
	h := headers{}
	h.Set("Content-Type", "application/json")
	h.Set("date", "Fri, 13 Jun 2025 19:35:42 GMT")
	h.Set("Content-length", "85")
	want := "Content-Type: application/json\r\ndate: Fri, 13 Jun 2025 19:35:42 GMT\r\nContent-length: 85\r\n"

	verifyWriteToOutput(t, h, want, "h.WriteTo")
}

func TestHeadersWriteToEmpty(t *testing.T) {
	h := headers{}
	want := ""

	verifyWriteToOutput(t, h, want, "h.WriteTo empty")
}

func TestHeadersWriteToClearsNewlines(t *testing.T) {
	h := headers{}
	h.Set("Content-Length", "1\n5\n")
	want := "Content-Length: 15\r\n"

	verifyWriteToOutput(t, h, want, "h.WriteTo clears new lines")
}

func TestHeadersWriteToClearsInternalCarriageReturn(t *testing.T) {
	h := headers{}
	h.Set("Content-Type", "application/\r\njson")
	want := "Content-Type: application/json\r\n"

	verifyWriteToOutput(t, h, want, "h.WriteTo clears internal carriage return")
}

func TestHeadersWriteToDoesNotClearInteriorWhitespace(t *testing.T) {
	h := headers{}
	h.Set("Content-Type", "application/json; charset=utf-8")
	want := "Content-Type: application/json; charset=utf-8\r\n"

	verifyWriteToOutput(t, h, want, "h.WriteTo does not clear interior whitespace")
}

func TestHeadersWriteToAllowsHTab(t *testing.T) {
	h := headers{}
	h.Set("Content-Type", "application/json\tcharset=utf-8")
	want := "Content-Type: application/json\tcharset=utf-8\r\n"

	verifyWriteToOutput(t, h, want, "h.WriteTo allows tabs")
}

func TestNewResponseHeadersDefaultsServer(t *testing.T) {
	h := newResponseHeaders()

	if len(h) != 1 {
		t.Errorf(`len(h) = %q, want match for %#q`, len(h), 1)
	}
	verifyPresentAndMatches(t, h, "newResponseHeaders defaults server", "Server", "routeit")
}

func TestSetOverwrites(t *testing.T) {
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

			verifyPresentAndMatches(t, h, "set overwrites", "Content-Length", "16")
			if len(h) != 1 {
				t.Errorf(`len(h) = %d, wanted only 1 element`, len(h))
			}
		})
	}
}

func TestSetSanitises(t *testing.T) {
	h := headers{}

	h.Set("Content\r\n-Length", "16\n\n\t")
	want := "16\t"

	verifyPresentAndMatches(t, h, "set sanitises", "Content-Length", want)
}

func TestHeadersFromRaw(t *testing.T) {
	raw := [][]byte{[]byte("Host: localhost"), []byte("Content-Type: application/json")}

	h, err := headersFromRaw(raw)

	if err != nil {
		t.Errorf("expected error to be nil: %s", err)
	}
	if len(h) != 2 {
		t.Errorf(`headers from raw len(h) = %d, want match for 2`, len(h))
	}
	verifyPresentAndMatches(t, h, "headers from raw", "Host", "localhost")
	verifyPresentAndMatches(t, h, "headers from raw", "Content-Type", "application/json")
}

func TestHeadersFromRawExitsAfterEmptyLines(t *testing.T) {
	raw := [][]byte{[]byte(""), []byte("Host: localhost")}

	h, err := headersFromRaw(raw)
	if err != nil {
		t.Errorf("expected error to be nil: %s", err)
	}
	if len(h) != 0 {
		t.Errorf(`headers from raw empty line len(h) = %d, wanted 0`, len(h))
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

func TestHeadersLookupCaseInsensitive(t *testing.T) {
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
			verifyPresentAndMatches(t, base, "case insensitive", tc, "val")
		})
	}
}

func verifyWriteToOutput(t *testing.T, h headers, want string, msg string) {
	t.Helper()
	var sb strings.Builder
	h.WriteTo(&sb)
	actual := sb.String()

	if len(actual) != len(want) {
		t.Errorf(`%s length = %d, want %d`, msg, len(actual), len(want))
	}

	actualLines := make([]string, 0, len(actual))
	for l := range strings.Lines(actual) {
		actualLines = append(actualLines, l)
	}
	sort.Strings(actualLines)

	wantLines := make([]string, 0, len(want))
	for l := range strings.Lines(want) {
		wantLines = append(wantLines, l)
	}
	sort.Strings(wantLines)

	for i, l := range actualLines {
		wl := wantLines[i]
		if l != wl {
			t.Errorf(`%s = %q, want %#q`, msg, l, wl)
		}
	}
}

func verifyPresentAndMatches(t *testing.T, h headers, msg string, key string, want string) {
	t.Helper()
	got, exists := h.Get(key)
	if !exists {
		t.Errorf("%s, wanted %q to be present", msg, key)
	}
	if got != want {
		t.Errorf(`%s h[%q] = %q, want %#q`, msg, key, got, want)
	}
}
