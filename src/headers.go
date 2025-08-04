package routeit

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// RFC-2616 states that headers may have case insensitive keys. This is
// modelled within routeit by using a map that uses lower case keys. To
// preserve the original case of the (first appearance of the) key, the value
// of the map is a structure containing the value string and the original key
// string.
type headerVal struct {
	vals     []string
	original string
}

// TODO: does not handle where a request contains multiple headers of the same key, or a response does

type headers map[string]headerVal

// TODO:
type RequestHeaders struct {
	headers headers
}

// Parses a slice of byte slices into the headers type.
//
// Expects that the input has already been split on the carriage return symbol:
// \r\n, and will return the position of the last valid header line processed,
// stopping at the first blank line sequence if present. If a blank line is not
// present, we will return an error since per the RFC-9112 spec, the headers
// MUST be separated from the body by a blank CRLF line.
func headersFromRaw(raw [][]byte) (*RequestHeaders, int, *HttpError) {
	h := headers{}
	for i, line := range raw {
		if len(line) == 0 {
			// This is an empty line which is interpreted as the signal between
			// the end of the headers and the body. We return the current index
			// since this is the last valid "header" line we processed.
			return &RequestHeaders{headers: h}, i, nil
		}

		kvp := strings.SplitN(string(line), ":", 2)
		if len(kvp) == 1 || (strings.TrimSpace(kvp[0]) != kvp[0]) {
			// The key cannot contain any leading nor trailing whitespace per
			// RFC-9112. This may also be entered if the request does not
			// contain a valid empty line between headers and the body, which
			// we also reject.
			return nil, i - 1, ErrBadRequest()
		}
		h.Append(kvp[0], strings.TrimSpace(kvp[1]))
	}
	// If we get here, it means we have reached the end of the headers and
	// haven't encountered an empty line. This means the headers are malformed,
	// which we report to the caller by returning an error and reporting the
	// last valid index as the last element of the input slice.
	return &RequestHeaders{}, len(raw) - 1, ErrBadRequest()
}

// Writes the headers to the given string builder. Sanitises the keys and
// values before writing.
func (h headers) WriteTo(writer io.Writer) (int64, error) {
	total := int64(0)
	for _, v := range h {
		key := strings.Map(sanitiseHeader, strings.TrimSpace(v.original))
		for _, val := range v.vals {
			val := strings.Map(sanitiseHeader, val)
			written, err := fmt.Fprintf(writer, "%s: %s\r\n", key, val)
			total += int64(written)
			if err != nil {
				return total, err
			}
		}
	}
	return total, nil
}

// Sets a key-value pair in the headers. This is a case insensitive operation
// that will create a new entry in the map if needed or update an existing
// entry if already present.
func (h headers) Set(key, val string) {
	sKey := strings.Map(sanitiseHeader, key)
	sVal := strings.Map(sanitiseHeader, val)
	sKeyLower := strings.ToLower(sKey)
	actual, exists := h[sKeyLower]
	if !exists {
		h[sKeyLower] = headerVal{
			vals:     []string{sVal},
			original: sKey,
		}
	} else {
		actual.vals = []string{sVal}
		h[sKeyLower] = actual
	}
}

// TODO:
func (h headers) Append(key, val string) {
	sKey := strings.Map(sanitiseHeader, key)
	sVal := strings.Map(sanitiseHeader, val)
	sKeyLower := strings.ToLower(sKey)
	actual, exists := h[sKeyLower]
	if !exists {
		h[sKeyLower] = headerVal{
			vals:     []string{sVal},
			original: sKey,
		}
	} else {
		actual.vals = append(actual.vals, sVal)
		h[sKeyLower] = actual
	}

}

// TODO:
func (rh *RequestHeaders) All(key string) ([]string, bool) {
	vals, found := rh.headers.All(key)
	return vals, found
}

// TODO:
func (rh *RequestHeaders) First(key string) (string, bool) {
	vals, found := rh.headers.All(key)
	if !found || len(vals) == 0 {
		return "", false
	}
	return vals[0], found
}

func (h headers) All(key string) ([]string, bool) {
	val, found := h[strings.ToLower(key)]
	if found {
		return val.vals, true
	}
	return []string{}, false
}

// Extract the content length field from the header map, defaulting to 0 if not
// present
func (h headers) ContentLength() uint {
	cLenRaw, found := h.All("Content-Length")
	if !found {
		return 0
	}
	if len(cLenRaw) != 1 {
		return 0
	}
	cLen, err := strconv.Atoi(cLenRaw[0])
	if err != nil {
		return 0
	}
	return uint(cLen)
}

func sanitiseHeader(r rune) rune {
	// We only allow printable ASCII characters, which are between 32 and 126
	// (decimal) inclusive. Additionally, the HTAB character is allowed.
	if (r < ' ' || r > '~') && r != '\t' {
		return -1
	}
	return r
}
