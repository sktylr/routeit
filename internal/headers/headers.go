package headers

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

type headers map[string]headerVal

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

// Appends a value to the headers. This is not destructive and does not check
// for the presence of the value already within the list. This is preferred to
// setting the header to a comma separated string, unless the header values
// need to be strictly reset.
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
