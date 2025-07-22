package routeit

import (
	"net/url"
	"strings"
)

type pathParameters map[string]string

type queryParameters map[string]string

// A composed structure representing the target of a request. It contains the
// parsed URL, which does not include the Host header and is always prefixed
// with a leading slash. Query parameters are extracted into a separate property
// and path parameters may be populated by the router.
type uri struct {
	// The edge path is the path that the request reaches the edge of the
	// server with. This may be different to the rewritten path, if URL
	// rewrites are configured for the server.
	edgePath      string
	rewrittenPath string
	rawPath       string
	pathParams    pathParameters
	queryParams   queryParameters
}

// TODO: this doesn't enforce that all characters are legal
func parseUri(uriRaw string) (*uri, *HttpError) {
	// When the raw URI is "*", this means that the request is an OPTIONS
	// request for the whole server. At this point we know that if the URI is
	// "*", then it is a valid request. We can skip the rest of the parsing.
	if uriRaw == "*" {
		return &uri{edgePath: uriRaw, rawPath: uriRaw}, nil
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

	path, err := url.PathUnescape(rawPath)
	if err != nil {
		return nil, ErrBadRequest()
	}

	if !strings.HasPrefix(path, "/") {
		// Per FRC-9112 Section 3.2.1 guidance, origin-form request targets
		// must include a leading slash. This server adopts a lenient approach
		// that will prefix this slash if not present. If the URI is invalid it
		// will be found later by the router.
		path = "/" + path
		rawPath = "/" + rawPath
	}

	if path != "/" {
		path = strings.TrimSuffix(path, "/")
	}

	uri := &uri{edgePath: path, rawPath: rawPath, queryParams: queryParameters{}}

	if hasQuery {
		if err := parseQueryParams(rawQuery, &uri.queryParams); err != nil {
			return nil, err
		}
	}

	return uri, nil
}

func (u *uri) RewritePath(r *router) error {
	rewrite, didRewrite := r.Rewrite(u.edgePath)
	if !didRewrite {
		return nil
	}
	rewrittenPath, rewrittenQuery, hasQuery := strings.Cut(rewrite, "?")
	u.rewrittenPath = rewrittenPath
	if !hasQuery {
		return nil
	}
	return parseQueryParams(rewrittenQuery, &u.queryParams)
}

func parseQueryParams(rawQuery string, queryParams *queryParameters) *HttpError {
	if strings.Contains(rawQuery, "?") {
		// There should only be 1 `?`, which we have stripped off. Any `?` that
		// feature as part of the query string should be URL encoded.
		return ErrBadRequest()
	}

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
		(*queryParams)[key] = val
	}

	return nil
}
