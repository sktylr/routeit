package routeit

import (
	"fmt"
	"net/url"
	"strings"
)

type queryParameters map[string][]string

type QueryParams struct {
	q queryParameters
}

func newQueryParams() *QueryParams {
	return &QueryParams{q: queryParameters{}}
}

// Access a query parameter if present. This returns a slice that may contain
// multiple elements, some of which may be empty (e.g. if the client sends
// `?foo=`).
func (q *QueryParams) All(key string) ([]string, bool) {
	val, found := q.q[key]
	return val, found
}

// Access a query parameter, asserting that it is only present exactly once in
// the request URI. This will return false if the query parameter is not
// present, and a 400: Bad Request error if the query parameter is present more
// than once.
func (q *QueryParams) Only(key string) (string, bool, error) {
	val, found := q.All(key)
	if !found {
		return "", false, nil
	}
	if len(val) != 1 {
		msg := fmt.Sprintf("Query parameter %#q should only be present once", key)
		return "", true, ErrBadRequest().WithMessage(msg)
	}
	return val[0], true, nil
}

// Access the first query parameter for the given key, if present
func (q *QueryParams) First(key string) (string, bool) {
	val, found := q.All(key)
	if !found || len(val) == 0 {
		return "", false
	}
	return val[0], true
}

// Access the last query parameter for the given key, if present
func (q *QueryParams) Last(key string) (string, bool) {
	val, found := q.All(key)
	length := len(val)
	if !found || length == 0 {
		return "", false
	}
	return val[length-1], true
}

func parseQueryParams(rawQuery string, params *QueryParams) *HttpError {
	if strings.Contains(rawQuery, "?") {
		// There should only be 1 `?`, which we have stripped off. Any `?` that
		// feature as part of the query string should be URL encoded.
		return ErrBadRequest()
	}

	queryParams := params.q
	for query := range strings.SplitSeq(rawQuery, "&") {
		// Most servers interpret the query component "?foo=" or "?foo" to mean
		// that the value of "foo" is "".
		key, rest, _ := strings.Cut(query, "=")
		key, err := url.QueryUnescape(key)
		if err != nil {
			return ErrBadRequest()
		}
		val, err := url.QueryUnescape(rest)
		if err != nil {
			return ErrBadRequest()
		}
		queryParams[key] = append(queryParams[key], val)
	}

	return nil
}
