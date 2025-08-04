package routeit

import (
	"net/url"
	"strings"
)

type pathParameters map[string]string

// A composed structure representing the target of a request. It contains the
// parsed URL, which does not include the Host header and is always prefixed
// with a leading slash. Query parameters are extracted into a separate property
// and path parameters may be populated by the router.
type uri struct {
	// The edge path is the path that the request reaches the edge of the
	// server with. This may be different to the rewritten path, if URL
	// rewrites are configured for the server.
	edgePath      []string
	rewrittenPath []string
	rawPath       string
	rewritten     bool
	globalOptions bool
	pathParams    pathParameters
	queryParams   *QueryParams
}

// TODO: this doesn't enforce that all characters are legal
func parseUri(uriRaw string) (*uri, *HttpError) {
	// When the raw URI is "*", this means that the request is an OPTIONS
	// request for the whole server. At this point we know that if the URI is
	// "*", then it is a valid request. We can skip the rest of the parsing.
	if uriRaw == "*" {
		return &uri{edgePath: []string{"*"}, rawPath: uriRaw, globalOptions: true}, nil
	}

	// The client (e.g. browsers) typically strips the fragment from the
	// request before sending it. routeit does not know how to respond to a URI
	// that contains a fragment, so we simply ignore it. We don't reject the
	// request since we can still interpret it without the fragment, and will
	// reject it later on if the URI is malformed.
	uriRaw, _, _ = strings.Cut(uriRaw, "#")

	// The URI should not contain any ASCII control characters
	for _, b := range uriRaw {
		if b < ' ' || b == 0x7F {
			return nil, ErrBadRequest()
		}
	}

	rawPath, rawQuery, hasQuery := strings.Cut(uriRaw, "?")

	var edgePath []string
	last := -1
	for i, rawPart := range strings.Split(rawPath, "/") {
		if i == 0 && rawPart == "" {
			continue
		}
		part, err := url.PathUnescape(rawPart)
		if err != nil {
			return nil, ErrBadRequest().WithCause(err)
		}
		edgePath = append(edgePath, part)
		last++
	}
	if last > 0 && edgePath[last] == "" {
		edgePath = edgePath[:last]
	}

	if !strings.HasPrefix(rawPath, "/") {
		// Per FRC-9112 Section 3.2.1 guidance, origin-form request targets
		// must include a leading slash. This server adopts a lenient approach
		// that will prefix this slash if not present. If the URI is invalid it
		// will be found later by the router.
		rawPath = "/" + rawPath
	}

	uri := &uri{edgePath: edgePath, rawPath: rawPath, queryParams: newQueryParams()}

	if hasQuery {
		if err := parseQueryParams(rawQuery, uri.queryParams); err != nil {
			return nil, err
		}
	}

	return uri, nil
}

func (u uri) Path() []string {
	if u.rewritten {
		return u.rewrittenPath
	}
	return u.edgePath
}

// Strips the namespace from the uri's path, returning false if that is not
// possible for some reason (e.g. the path does not start with the namespace).
func (u uri) RemoveNamespace(ns []string) ([]string, bool) {
	path := u.Path()

	if len(path) < len(ns) {
		return nil, false
	}

	nonNamespace := 0
	for i, seg := range ns {
		if seg != path[i] {
			return nil, false
		}
		nonNamespace++
	}
	return path[nonNamespace:], true
}
