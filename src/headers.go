package routeit

import (
	"fmt"
	"strconv"
	"strings"
)

// RFC-2616 states that headers may have case insensitive keys. This is
// modelled within routeit by using a map that uses lower case keys. To
// preserve the original case of the (first appearance of) the key, the value
// of the map is a structure containing the value string and the original key
// string.
type headerVal struct {
	val      string
	original string
}

type headers map[string]headerVal

func newResponseHeaders() headers {
	h := headers{}
	h.Set("Server", "routeit")
	return h
}

// Parses a slice of byte slices into the headers type.
//
// Expects that the input has already been split on the carriage return symbol \r\n
func headersFromRaw(raw [][]byte) (headers, *HttpError) {
	h := headers{}
	for _, line := range raw {
		if len(line) == 0 {
			// Empty line which should indicate the end of the headers. If it
			// does not, we exit anyway and ignore the remaining headers, since
			// it should.
			// TODO: consider increasing the strictness here by rejecting the request
			return h, nil
		}

		kvp := strings.SplitN(string(line), ":", 2)
		k, v := kvp[0], kvp[1]
		if strings.TrimSpace(k) != k {
			// The key cannot contain any leading nor trailing whitespace per
			// RFC-9112
			return nil, BadRequestError()
		}
		h.Set(k, strings.TrimSpace(v))
	}
	return h, nil
}

// Writes the headers to the given string builder. Sanitises the keys and
// values before writing.
func (h headers) WriteTo(sb *strings.Builder) {
	for _, v := range h {
		key := strings.Map(sanitiseHeader, strings.TrimSpace(v.original))
		val := strings.Map(sanitiseHeader, v.val)
		fmt.Fprintf(sb, "%s: %s\r\n", key, val)
	}
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
