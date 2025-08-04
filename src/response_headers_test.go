package routeit

import "testing"

func TestNewResponseHeaders(t *testing.T) {
	h := newResponseHeaders()

	if len(h.headers) != 1 {
		t.Errorf(`len(h) = %q, want match for %#q`, len(h.headers), 1)
	}
	verifyPresentAndMatches(t, h.headers, "Server", []string{"routeit"})
}
