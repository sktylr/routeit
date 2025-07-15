package routeit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
)

var (
	GET  = HttpMethod{"GET"}
	HEAD = HttpMethod{"HEAD"}
	POST = HttpMethod{"POST"}
)

var methodLookup = map[string]HttpMethod{
	"GET":  GET,
	"HEAD": HEAD,
	"POST": POST,
}

type Request struct {
	mthd HttpMethod
	// TODO: need to normalise this properly and have a (private) method for trimming the namespace etc.
	uri     uri
	headers headers
	// TODO: consider byte slice here
	body string
	ct   ContentType
}

type HttpMethod struct {
	name string
}

type pathParameters map[string]string

type queryParameters map[string]string

// A composed structure representing the target of a request. It contains the
// parsed URL, which does not include the Host header and is always prefixed
// with a leading slash. Query parameters are extracted into a separate property
// and path parameters may be populated by the router.
type uri struct {
	url         string
	pathParams  pathParameters
	queryParams queryParameters
}

type protocolLine struct {
	mthd  HttpMethod
	path  string
	prtcl string
}

// Parses the raw byte slice of the request into a more usable request structure
//
// The request is made up of three components: the protocol line, headers and the
// body. For HTTP/1.1, at a bare minimum the Host header must be included, though
// the body is optional (and ignored for certain request methods such as GET).
//
// Each section is split using carriage returns (CRLF or \r\n). After the protocol
// line, and each header line is also split using a carriage return. The protocol
// line is always only a single line made up of three components - the request
// method, the path (or URI) and the http protocol, and a blank line (using a
// carriage return) also follows the headers before the optional body.
//
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Guides/Messages
func requestFromRaw(raw []byte) (*Request, *HttpError) {
	sections := bytes.Split(raw, []byte("\r\n"))

	// We are expecting 1 carriage return after the protocol line, 1 carriage
	// return after the Host header and 1 carriage return after all the headers.
	// This means there will be at least 4 sections.
	if len(sections) < 4 {
		return nil, BadRequestError()
	}

	prtclRaw := sections[0]
	hdrsRaw := sections[1 : len(sections)-1]
	// TODO: make sure that this is correct (i.e. if a body is not included, do the specs state that there must be a carriage return anyway, so we can safely take the body as the last carriage return object?)
	bdyRaw := sections[len(sections)-1]

	ptcl, err := parseProtocolLine(prtclRaw)
	if err != nil {
		return nil, err
	}

	uri, err := parseUri(ptcl.path)
	if err != nil {
		return nil, err
	}

	reqHdrs, err := headersFromRaw(hdrsRaw)
	if err != nil {
		return nil, err
	}

	// TODO: in future, we should verify that this is an allowed host
	_, hasHost := reqHdrs["Host"]
	if !hasHost {
		// The Host header is required as part of HTTP/1.1
		return nil, BadRequestError()
	}

	cLen := reqHdrs.contentLength()
	var body string
	// TODO: make the buffer size also depend on the server max allowed request
	if cLen <= 0 || ptcl.mthd == GET || ptcl.mthd == HEAD {
		// For GET or HEAD requests, the request body should be ignored even if
		// provided. Where we are consuming the body, we should only look for
		// Content-Length bytes and no more.
		body = ""
	} else {
		// TODO: we need to return 413 Payload Too Large if the total payload exceeds defined bounds
		reader := bytes.NewReader(bdyRaw)
		buf := make([]byte, cLen)
		_, err := io.ReadFull(reader, buf)
		if err != nil {
			// Http servers are expected to read **exactly** Content-Length bytes
			// from the request body. This error is returned if the reader contains
			// **less** than the requested number of bytes, so we cannot read it
			// all. Either the client has not sent it all (e.g. due to a slow
			// connection), or the request is malformed. Return 400 Bad Request
			// since the failure is with the client.
			return nil, BadRequestError()
		}
		body = string(buf)
	}

	ct := ContentType{}
	// TODO: can probably also reject the request body if it doesn't include a CT
	ctRaw := reqHdrs["Content-Type"]
	if ptcl.mthd != GET && ptcl.mthd != HEAD {
		ct = parseContentType(ctRaw)
	}

	req := Request{mthd: ptcl.mthd, uri: uri, headers: reqHdrs, body: body, ct: ct}
	return &req, nil
}

// Access the request's HTTP method
func (req *Request) Method() HttpMethod {
	return req.mthd
}

// The request's URL excluding the host. Does not include query parameters.
func (req *Request) Url() string {
	return req.uri.url
}

func (req *Request) PathParam(param string) (string, bool) {
	val, found := req.uri.pathParams[param]
	return val, found
}

// Access a header value, if present
func (req *Request) Header(key string) (string, bool) {
	val, found := req.headers[key]
	return val, found
}

// TODO: query params are currently not url decoded!
// Access a query parameter if present
func (req *Request) QueryParam(key string) (string, bool) {
	val, found := req.uri.queryParams[key]
	return val, found
}

// TODO: improve this in the future to provide an additional method that confirms the content type of the request and makes sure it is application/json
// TODO: should look into prohibiting (via a panic or error) this method for GET or HEAD requests, since they should not contain a req body and if they do the server shouldn't read it
// Parses the Json response body into the destination
func (req *Request) BodyToJson(to any) error {
	v := reflect.ValueOf(to)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		// We panic for now, this may change. This is due to an issue introduced
		// by the integrator, so we panic which will manifest itself as a 500:
		// Internal Server Error outside of the integrator's control.
		panic(fmt.Sprintf("BodyToJson requires a non-nil pointer destination, got %T", to))
	}
	return json.Unmarshal([]byte(req.body), to)
}

// Access the content type of the request body. Does not return a meaningful
// value for requests without the Content-Type header, or requests that should
// not contain a body, such as GET. The ContentType.Equals method can be used
// to perform equality checks.
func (req *Request) ContentType() ContentType {
	return req.ct
}

func parseProtocolLine(raw []byte) (protocolLine, *HttpError) {
	split := bytes.Split(raw, []byte(" "))
	if len(split) != 3 {
		return protocolLine{}, BadRequestError()
	}

	mthdRaw, path, prtcl := split[0], string(split[1]), string(split[2])
	mthd, found := methodLookup[string(mthdRaw)]
	if !found {
		return protocolLine{}, NotImplementedError()
	}
	if prtcl != "HTTP/1.1" {
		return protocolLine{}, HttpVersionNotSupportedError()
	}

	// TODO: need to return 414: URI Too Long if URI is too long
	return protocolLine{mthd, path, prtcl}, nil
}

func parseUri(raw string) (uri, *HttpError) {
	split := strings.Split(raw, "?")

	endpoint := split[0]
	if !strings.HasPrefix(endpoint, "/") {
		// Per FRC-9112 Section 3.2.1 guidance, origin-form request targets
		// must include a leading slash. This server adopts a lenient approach
		// that will prefix this slash if not present. If the URI is invalid it
		// will be found later by the router.
		endpoint = "/" + endpoint
	}

	queryParams := queryParameters{}
	uri := uri{url: endpoint, queryParams: queryParams}

	if len(split) == 1 {
		// No query string present
		return uri, nil
	}

	if len(split) > 2 {
		// There should only be 1 `?`. Any `?` that feature as part of the
		// query string should be URL encoded.
		return uri, BadRequestError()
	}

	for query := range strings.SplitSeq(split[1], "&") {
		kvp := strings.Split(query, "=")
		if len(kvp) != 2 {
			return uri, BadRequestError()
		}
		queryParams[kvp[0]] = kvp[1]
	}

	return uri, nil
}
