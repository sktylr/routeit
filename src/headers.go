package routeit

import (
	"fmt"
	"strconv"
	"strings"
)

type headers map[string]string

// TODO: this should have a carriage return on everything!
func (h headers) writeTo(sb *strings.Builder) {
	for k, v := range h {
		key := strings.Map(sanitiseHeader, strings.TrimSpace(k))
		val := strings.Map(sanitiseHeader, v)
		sb.WriteString(fmt.Sprintf("%s: %s\n", key, val))
	}
}

func (h headers) set(key string, val string) {
	sKey := strings.Map(sanitiseHeader, key)
	sVal := strings.Map(sanitiseHeader, val)
	h[sKey] = sVal
}

func newResponseHeaders() headers {
	return headers{"Server": "routeit"}
}

// Parses a slice of byte slices into the headers type.
//
// Expects that the input has already been split on the carriage return symbol \r\n
func headersFromRaw(raw [][]byte) headers {
	h := headers{}
	for i, line := range raw {
		if len(line) == 0 {
			// Empty line which should indicate the end of the headers
			count := len(raw)
			if count != i+1 {
				// There are more headers after this, signalling a malformed request.
				// Ignore subsequent properties for now.
				fmt.Printf("found end of headers section, returning... (total headers = %d, on index = %d)\n", len(raw), i)
			}
			return h
		}

		kvp := strings.SplitN(string(line), ": ", 2)
		if len(kvp) != 2 {
			fmt.Printf("Malformed header: [%s]\n", string(line))
			continue
		}
		h[kvp[0]] = kvp[1]
	}
	return h
}

// Extract the content length field from the header map, defaulting to 0 if not present
func (h headers) contentLength() uint {
	cLenRaw, found := h["Content-Length"]
	if !found {
		return 0
	}
	cLen, err := strconv.Atoi(cLenRaw)
	if err != nil {
		fmt.Print(err)
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
