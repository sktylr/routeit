package routeit

import "strings"

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

// TODO:
func (rh *RequestHeaders) All(key string) ([]string, bool) {
	vals, found := rh.headers.All(key)
	return vals, found
}

// TODO:
func (rh *RequestHeaders) Only(key string) (string, bool, error) {
	vals, found := rh.All(key)
	if !found {
		return "", false, nil
	}
	if len(vals) != 1 {
		return "", true, ErrBadRequest().WithMessagef("Header %#q should only be present once", key)
	}
	return vals[0], true, nil
}

// TODO:
func (rh *RequestHeaders) First(key string) (string, bool) {
	vals, found := rh.All(key)
	if !found || len(vals) == 0 {
		return "", false
	}
	return vals[0], found
}

// TODO:
func (rh *RequestHeaders) Last(key string) (string, bool) {
	vals, found := rh.All(key)
	length := len(vals)
	if !found || length == 0 {
		return "", false
	}
	return vals[length-1], found
}
