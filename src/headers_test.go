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
	want := "Content-Type: application/json\nDate: Fri, 13 Jun 2025 19:35:42 GMT\nContent-Length: 85\n"

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
	want := "Content-Length: 15\n"

	verifyWriteToOutput(t, h, want, "h.writeTo clears new lines")
}

func TestHeadersWriteToClearsInternalCarriageReturn(t *testing.T) {
	h := headers{
		"Content-Type": "application/\r\njson",
	}
	want := "Content-Type: application/json\n"

	verifyWriteToOutput(t, h, want, "h.writeTo clears internal carriage return")
}

func TestHeadersWriteToDoesNotClearInteriorWhitespace(t *testing.T) {
	h := headers{
		"Content-Type": "application/json; charset=utf-8",
	}
	want := "Content-Type: application/json; charset=utf-8\n"

	verifyWriteToOutput(t, h, want, "h.writeTo does not clear interior whitespace")
}

func TestHeadersWriteToAllowsHTab(t *testing.T) {
	h := headers{
		"Content-Type": "application/json\tcharset=utf-8",
	}
	want := "Content-Type: application/json\tcharset=utf-8\n"

	verifyWriteToOutput(t, h, want, "h.writeTo allows tabs")
}

func TestNewHeadersDefaultsServer(t *testing.T) {
	h := newHeaders()

	if len(h) != 1 {
		t.Errorf(`len(h) = %q, want match for %#q`, len(h), 1)
	}
	val, found := h["Server"]
	if !found {
		t.Error("Expected to find 'Server' in default headers")
	}
	if val != "routeit" {
		t.Errorf(`h["Server"] = %q, want "routeit"`, val)
	}
}

func TestSetOverwrites(t *testing.T) {
	h := headers{
		"Content-Length": "15",
	}

	h.set("Content-Length", "16")

	cl, found := h["Content-Length"]
	if !found {
		t.Error("set overwrites, wanted 'Content-Length' to be present")
	}
	if cl != "16" {
		t.Errorf(`set overwrites h["Content-Length"] = %q, want "16"`, cl)
	}
}

func TestSetSanitises(t *testing.T) {
	h := headers{}

	h.set("Content\r\n-Length", "16\n\n\t")
	want := "16\t"

	cl, found := h["Content-Length"]
	if !found {
		t.Error("set sanitises, wanted 'Content-Length' to be present")
	}
	if cl != want {
		t.Errorf(`set sanitises h["Content-Length"] = %q, want %#q`, cl, want)
	}

}

func verifyWriteToOutput(t *testing.T, h headers, want string, msg string) {
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
