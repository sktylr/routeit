package routeit

import (
	"fmt"
	"strconv"
	"strings"
)

// TODO: need to make this work for case insensitive lookup
type headers map[string]string

func newResponseHeaders() headers {
	return headers{"Server": "routeit"}
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

		// TODO: also need to ensure exactly 1 :
		// TODO: improve this properly
		kvp := strings.Split(string(line), ": ")
		if len(kvp) != 2 {
			// TODO: this should use structured logging
			fmt.Printf("Malformed header: [%s]\n", string(line))
			return h, BadRequestError()
		}
		h[kvp[0]] = strings.TrimPrefix(kvp[1], " ")
	}
	return h, nil
}

func (h headers) WriteTo(sb *strings.Builder) {
	for k, v := range h {
		key := strings.Map(sanitiseHeader, strings.TrimSpace(k))
		val := strings.Map(sanitiseHeader, v)
		fmt.Fprintf(sb, "%s: %s\r\n", key, val)
	}
}

func (h headers) Set(key string, val string) {
	sKey := strings.Map(sanitiseHeader, key)
	sVal := strings.Map(sanitiseHeader, val)
	h[sKey] = sVal
}

// Extract the content length field from the header map, defaulting to 0 if not
// present
func (h headers) ContentLength() uint {
	cLenRaw, found := h["Content-Length"]
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
