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
	val      string
	original string
}

// TODO: does not handle where a request contains multiple headers of the same key, or a response does

type headers map[string]headerVal

func newResponseHeaders() headers {
	h := headers{}
	h.Set("Server", "routeit")
	return h
}

// Parses a slice of byte slices into the headers type.
//
// Expects that the input has already been split on the carriage return symbol:
// \r\n, and will return the position of the last valid header line processed,
// stopping at the first blank line sequence if present. If a blank line is not
// present, we will return an error since per the RFC-9112 spec, the headers
// MUST be separated from the body by a blank CRLF line.
func headersFromRaw(raw [][]byte) (headers, int, *HttpError) {
	h := headers{}
	for i, line := range raw {
		if len(line) == 0 {
			// This is an empty line which is interpreted as the signal between
			// the end of the headers and the body. We return the current index
			// since this is the last valid "header" line we processed.
			return h, i, nil
		}

		kvp := strings.SplitN(string(line), ":", 2)
		if len(kvp) == 1 || (strings.TrimSpace(kvp[0]) != kvp[0]) {
			// The key cannot contain any leading nor trailing whitespace per
			// RFC-9112. This may also be entered if the request does not
			// contain a valid empty line between headers and the body, which
			// we also reject.
			return nil, i - 1, ErrBadRequest()
		}
		h.Set(kvp[0], strings.TrimSpace(kvp[1]))
	}
	// If we get here, it means we have reached the end of the headers and
	// haven't encountered an empty line. This means the headers are malformed,
	// which we report to the caller by returning an error and reporting the
	// last valid index as the last element of the input slice.
	return headers{}, len(raw) - 1, ErrBadRequest()
}

// Writes the headers to the given string builder. Sanitises the keys and
// values before writing.
func (h headers) WriteTo(writer io.Writer) (int64, error) {
	total := int64(0)
	for _, v := range h {
		key := strings.Map(sanitiseHeader, strings.TrimSpace(v.original))
		val := strings.Map(sanitiseHeader, v.val)
		written, err := fmt.Fprintf(writer, "%s: %s\r\n", key, val)
		total += int64(written)
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

// Sets a key-value pair in the headers. This is a case insensitive operation
// that will create a new entry in the map if needed or update an existing
// entry if already present.
func (h headers) Set(key string, val string) {
	sKey := strings.Map(sanitiseHeader, key)
	sVal := strings.Map(sanitiseHeader, val)
	sKeyLower := strings.ToLower(sKey)
	actual, exists := h[sKeyLower]
	if !exists {
		h[sKeyLower] = headerVal{
			val:      sVal,
			original: sKey,
		}
	} else if actual.val != sVal {
		actual.val = sVal
		h[sKeyLower] = actual
	}
}

// Performs a case insensitive retrieval of the value associated with the given
// key, indicating a success or failure in the second return value.
func (h headers) Get(key string) (string, bool) {
	lower := strings.ToLower(key)
	val, found := h[lower]
	return val.val, found
}

// Extract the content length field from the header map, defaulting to 0 if not
// present
func (h headers) ContentLength() uint {
	cLenRaw, found := h.Get("Content-Length")
	if !found {
		return 0
	}
	cLen, err := strconv.Atoi(cLenRaw)
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
