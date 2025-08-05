package routeit

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
)

type TestResponse struct{ rw *ResponseWriter }

// Parses the Json response into a destination object. Fails if the Json
// parsing fails or if the response is not a Json response. The destination
// must be passed by reference and not by value.
func (tr *TestResponse) BodyToJson(t testing.TB, to any) {
	t.Helper()
	tr.AssertHeaderMatchesString(t, "Content-Type", "application/json")

	v := reflect.ValueOf(to)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		t.Fatalf("BodyToJson() requires a non-nil pointer destination, got %T", to)
	}

	err := json.Unmarshal(tr.rw.bdy, to)
	if err != nil {
		t.Fatalf(`failed to parse Json response: %v`, err)
	}
}

// Assert that the response body is empty
func (tr *TestResponse) AssertBodyEmpty(t testing.TB) {
	t.Helper()
	if len(tr.rw.bdy) != 0 {
		t.Errorf(`expected empty body, got %#q`, tr.rw.bdy)
	}
}

// Assert that a body contains the given substring. Supports improper
// substrings (i.e. where the substring is exactly equal to the superstring).
func (tr *TestResponse) AssertBodyContainsString(t testing.TB, want string) {
	t.Helper()
	if !strings.Contains(string(tr.rw.bdy), want) {
		t.Errorf(`body = %#q, wanted to contain %#q`, string(tr.rw.bdy), want)
	}
}

// Assert that a body exactly matches the given string
func (tr *TestResponse) AssertBodyMatchesString(t testing.TB, want string) {
	t.Helper()
	if string(tr.rw.bdy) != want {
		t.Errorf(`body = %#q, wanted to equal %#q`, string(tr.rw.bdy), want)
	}
}

// Assert that a body exactly matches the given string with format options
// This is the same as formatting the string using fmt.Sprintf and calling
// AssertBodyMatchesString directly
func (tr *TestResponse) AssertBodyMatchesStringf(t testing.TB, wantf string, args ...any) {
	t.Helper()
	want := fmt.Sprintf(wantf, args...)
	tr.AssertBodyMatchesString(t, want)
}

// Assert that a body starts with the given prefix. Supports improper
// substrings (i.e. where the prefix exactly equals the whole body).
func (tr *TestResponse) AssertBodyStartsWithString(t testing.TB, want string) {
	t.Helper()
	if !strings.HasPrefix(string(tr.rw.bdy), want) {
		t.Errorf(`body = %#q, wanted to start with %#q`, string(tr.rw.bdy), want)
	}
}

// Assert that a header is present and matches the given slice of strings
func (tr *TestResponse) AssertHeaderMatches(t testing.TB, header string, want []string) {
	t.Helper()
	val, found := tr.rw.headers.headers.All(header)
	if !found {
		t.Errorf(`expected %#q header to be present`, header)
	}
	if !reflect.DeepEqual(val, want) {
		t.Errorf(`headers[%#q] = %+v, wanted %+v`, header, val, want)
	}
}

// Similar to [TestResponse.AssertHeaderMatches], except we assert there is
// only exactly 1 element in the header slice, and it matches the given string
// exactly.
func (tr *TestResponse) AssertHeaderMatchesString(t testing.TB, header, want string) {
	t.Helper()
	val, found := tr.rw.headers.headers.All(header)
	if !found {
		t.Errorf(`expected %#q header to be present`, header)
	}
	if len(val) != 1 {
		t.Errorf(`headers[%#q] = %+v with length %d, wanted exactly 1 element`, header, val, len(val))
	}
	if val[0] != want {
		t.Errorf(`headers[%#q] = %#q, wanted %#q`, header, val[0], want)
	}
}

// Assert a header is present and that it contains the given string. For
// clarity, this will only compare over a single slice element, it will not
// join multiple elements into 1 string for comparison.
func (tr *TestResponse) AssertHeaderContains(t testing.TB, header, want string) {
	t.Helper()
	val, found := tr.rw.headers.headers.All(header)
	if !found {
		t.Errorf(`expected %#q header to be present`, header)
	}
	if !slices.Contains(val, want) {
		t.Errorf(`headers[%#q] = %+v, wanted %#q`, header, val, want)
	}
}

// Asserts that a header key is not present in the response
func (tr *TestResponse) RefuteHeaderPresent(t testing.TB, header string) {
	t.Helper()
	val, found := tr.rw.headers.headers.All(header)
	if found {
		t.Errorf(`Headers[%#q] = %+v, did not expect to be present`, header, val)
	}
}

// Assert that the status code of the response matches
func (tr *TestResponse) AssertStatusCode(t testing.TB, want HttpStatus) {
	t.Helper()
	if tr.rw.s != want {
		t.Errorf(`status = %d, wanted %d`, tr.rw.s.code, want.code)
	}
}
