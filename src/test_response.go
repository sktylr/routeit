package routeit

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

type TestResponse struct{ rw *ResponseWriter }

// Parses the Json response into a destination object. Fails if the Json
// parsing fails or if the response is not a Json response. The destination
// must be passed by reference and not by value.
func (tr *TestResponse) BodyToJson(t *testing.T, to any) {
	t.Helper()
	tr.AssertHeaderMatches(t, "Content-Type", "application/json")

	v := reflect.ValueOf(to)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		t.Fatalf("BodyToJson() requires a non-nil pointer destination, got %T", to)
	}

	err := json.Unmarshal(tr.rw.bdy, to)
	if err != nil {
		t.Fatalf(`failed to parse Json response: %v`, err)
	}
}

// Assert that the response body is nil or empty
func (tr *TestResponse) AssertBodyNilOrEmpty(t *testing.T) {
	t.Helper()
	if len(tr.rw.bdy) != 0 {
		t.Errorf(`expected nil or empty body, got %#q`, tr.rw.bdy)
	}
}

// Assert that a body contains the given substring. Supports improper
// substrings (i.e. where the substring is exactly equal to the superstring).
func (tr *TestResponse) AssertBodyContainsString(t *testing.T, want string) {
	t.Helper()
	if !strings.Contains(string(tr.rw.bdy), want) {
		t.Errorf(`body = %#q, wanted to contain %#q`, string(tr.rw.bdy), want)
	}
}

// Assert that a body exactly matches the given string
func (tr *TestResponse) AssertBodyMatchesString(t *testing.T, want string) {
	t.Helper()
	if string(tr.rw.bdy) != want {
		t.Errorf(`body = %#q, wanted to equal %#q`, string(tr.rw.bdy), want)
	}
}

// TODO: assert body matches f string

// Assert that a body starts with the given prefix. Supports improper
// substrings (i.e. where the prefix exactly equals the whole body).
func (tr *TestResponse) AssertBodyStartsWithString(t *testing.T, want string) {
	t.Helper()
	if !strings.HasPrefix(string(tr.rw.bdy), want) {
		t.Errorf(`body = %#q, wanted to start with %#q`, string(tr.rw.bdy), want)
	}
}

// Assert that a header is present and matches the given string
func (tr *TestResponse) AssertHeaderMatches(t *testing.T, header string, want string) {
	t.Helper()
	val, found := tr.rw.hdrs[header]
	if !found {
		t.Errorf(`expected %#q header to be present`, header)
	}
	if val != want {
		t.Errorf(`headers[%#q] = %#q, wanted %#q`, header, val, want)
	}
}

// Assert that the status code of the response matches
func (tr *TestResponse) AssertStatusCode(t *testing.T, want HttpStatus) {
	t.Helper()
	if tr.rw.s != want {
		t.Errorf(`status = %d, wanted %d`, tr.rw.s.code, want.code)
	}
}
