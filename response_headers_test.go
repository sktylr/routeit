package routeit

import (
	"reflect"
	"testing"

	"github.com/sktylr/routeit/internal/headers"
)

func TestNewResponseHeaders(t *testing.T) {
	h := newResponseHeaders()

	if len(h.headers) != 1 {
		t.Errorf(`len(h) = %q, want match for %#q`, len(h.headers), 1)
	}
	verifyPresentAndMatches(t, h.headers, "Server", []string{"routeit"})
}

// TODO: need to sort this out!
func verifyPresentAndMatches(t *testing.T, h headers.Headers, key string, want []string) {
	t.Helper()
	got, exists := h.All(key)
	if !exists {
		t.Errorf("wanted %q to be present", key)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf(`h[%q] = %v, want %v`, key, got, want)
	}
}
