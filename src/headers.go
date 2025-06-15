package routeit

import (
	"fmt"
	"strings"
)

type headers map[string]string

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

func newHeaders() headers {
	return headers{"Server": "routeit"}
}

func sanitiseHeader(r rune) rune {
	// We only allow printable ASCII characters, which are between 32 and 126
	// (decimal) inclusive. Additionally, the HTAB character is allowed.
	if (r < ' ' || r > '~') && r != '\t' {
		return -1
	}

	return r
}
