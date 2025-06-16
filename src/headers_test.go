package routeit

import (
	"sort"
	"strings"
	"testing"
)

func TestHeadersWriteTo(t *testing.T) {
	h := headers{
		"Content-Type":   "application/json",
		"Date":           "Fri, 13 Jun 2025 19:35:42 GMT",
		"Content-Length": "85",
	}
	want := "Content-Type: application/json\r\nDate: Fri, 13 Jun 2025 19:35:42 GMT\r\nContent-Length: 85\r\n"

	verifyWriteToOutput(t, h, want, "h.writeTo")
}

func TestHeadersWriteToEmpty(t *testing.T) {
	h := headers{}
	want := ""

	verifyWriteToOutput(t, h, want, "h.writeTo empty")
}

func TestHeadersWriteToClearsNewlines(t *testing.T) {
	h := headers{
		"Content-Length": "1\n5\n",
	}
	want := "Content-Length: 15\r\n"

	verifyWriteToOutput(t, h, want, "h.writeTo clears new lines")
}

func TestHeadersWriteToClearsInternalCarriageReturn(t *testing.T) {
	h := headers{
		"Content-Type": "application/\r\njson",
	}
	want := "Content-Type: application/json\r\n"

	verifyWriteToOutput(t, h, want, "h.writeTo clears internal carriage return")
}

func TestHeadersWriteToDoesNotClearInteriorWhitespace(t *testing.T) {
	h := headers{
		"Content-Type": "application/json; charset=utf-8",
	}
	want := "Content-Type: application/json; charset=utf-8\r\n"

	verifyWriteToOutput(t, h, want, "h.writeTo does not clear interior whitespace")
}

func TestHeadersWriteToAllowsHTab(t *testing.T) {
	h := headers{
		"Content-Type": "application/json\tcharset=utf-8",
	}
	want := "Content-Type: application/json\tcharset=utf-8\r\n"

	verifyWriteToOutput(t, h, want, "h.writeTo allows tabs")
}

func TestNewResponseHeadersDefaultsServer(t *testing.T) {
	h := newResponseHeaders()

	if len(h) != 1 {
		t.Errorf(`len(h) = %q, want match for %#q`, len(h), 1)
	}
	verifyPresentAndMatches(t, h, "newResponseHeaders defaults server", "Server", "routeit")
}

func TestSetOverwrites(t *testing.T) {
	h := headers{
		"Content-Length": "15",
	}

	h.set("Content-Length", "16")

	verifyPresentAndMatches(t, h, "set overwrites", "Content-Length", "16")
}

func TestSetSanitises(t *testing.T) {
	h := headers{}

	h.set("Content\r\n-Length", "16\n\n\t")
	want := "16\t"

	verifyPresentAndMatches(t, h, "set sanitises", "Content-Length", want)
}

func TestHeadersFromRaw(t *testing.T) {
	raw := [][]byte{[]byte("Host: localhost"), []byte("Content-Type: application/json")}

	h := headersFromRaw(raw)

	if len(h) != 2 {
		t.Errorf(`headers from raw len(h) = %d, want match for 2`, len(h))
	}
	verifyPresentAndMatches(t, h, "headers from raw", "Host", "localhost")
	verifyPresentAndMatches(t, h, "headers from raw", "Content-Type", "application/json")
}

func TestHeadersFromRawExitsAfterEmptyLines(t *testing.T) {
	raw := [][]byte{[]byte(""), []byte("Host: localhost")}

	h := headersFromRaw(raw)
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
			headers{"Content-Length": "abc"},
			0,
		},
		{
			"valid",
			headers{"Content-Length": "85"},
			85,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cLen := tc.in.contentLength()
			if cLen != tc.want {
				t.Errorf(`h.contentLength() = %d, wanted %d`, cLen, tc.want)
			}
		})
	}
}

func verifyWriteToOutput(t *testing.T, h headers, want string, msg string) {
	t.Helper()
	var sb strings.Builder
	h.writeTo(&sb)
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
	got, exists := h[key]
	if !exists {
		t.Errorf("%s, wanted %q to be present", msg, key)
	}
	if got != want {
		t.Errorf(`%s h[%q] = %q, want %#q`, msg, key, got, want)
	}
}
