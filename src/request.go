package routeit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

var (
	GET     = HttpMethod{name: "GET"}
	HEAD    = HttpMethod{name: "HEAD"}
	POST    = HttpMethod{name: "POST"}
	PUT     = HttpMethod{name: "PUT"}
	OPTIONS = HttpMethod{name: "OPTIONS"}
)

var methodLookup = map[string]HttpMethod{
	"GET":     GET,
	"HEAD":    HEAD,
	"POST":    POST,
	"PUT":     PUT,
	"OPTIONS": OPTIONS,
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

type protocolLine struct {
	mthd  HttpMethod
	prtcl string
	uri   uri
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
	// TODO: need to add support for all request-target forms (origin-form, absolute-form, authority-form, asterisk-form) that should be accepted by a HTTP only server.
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

	reqHdrs, err := headersFromRaw(hdrsRaw)
	if err != nil {
		return nil, err
	}

	// TODO: in future, we should verify that this is an allowed host
	// TODO: this is not being parsed properly due to the : in the port
	_, hasHost := reqHdrs.Get("Host")
	if !hasHost {
		// The Host header is required as part of HTTP/1.1
		return nil, BadRequestError()
	}

	cLen := reqHdrs.ContentLength()
	var body string
	// TODO: make the buffer size also depend on the server max allowed request
	if cLen <= 0 || ptcl.mthd == GET || ptcl.mthd == HEAD || ptcl.mthd == OPTIONS {
		// For GET, HEAD or OPTIONS requests, the request body should be
		// ignored even if provided. Servers can technically accept request
		// bodies for OPTIONS requests, however it is up to the server
		// implementation, and routeit chooses not to. Where we are consuming
		// the body, we should only look for Content-Length bytes and no more.
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
	ctRaw, hasCType := reqHdrs.Get("Content-Type")
	if hasCType && ptcl.mthd != GET && ptcl.mthd != HEAD {
		ct = parseContentType(ctRaw)
	}

	req := Request{mthd: ptcl.mthd, uri: ptcl.uri, headers: reqHdrs, body: body, ct: ct}
	return &req, nil
}

// Access the request's HTTP method
func (req *Request) Method() HttpMethod {
	return req.mthd
}

// The request's URL excluding the host. Does not include query parameters.
func (req *Request) Path() string {
	return req.uri.path
}

func (req *Request) PathParam(param string) (string, bool) {
	val, found := req.uri.pathParams[param]
	return val, found
}

// Access a header value, if present
func (req *Request) Header(key string) (string, bool) {
	val, found := req.headers.Get(key)
	return val, found
}

// Access a query parameter if present
func (req *Request) QueryParam(key string) (string, bool) {
	val, found := req.uri.queryParams[key]
	return val, found
}

// TODO: should look into prohibiting (via a panic or error) this method for GET or HEAD requests, since they should not contain a req body and if they do the server shouldn't read it
// Parses the Json request body into the destination. Ensures that the
// Content-Type header is application/json and will return a 415: Unsupported
// Media Type error if this is not the case. Will panic if the destination is
// not a pointer.
func (req *Request) BodyToJson(to any) error {
	if !req.ContentType().Equals(CTApplicationJson) {
		return UnsupportedMediaTypeError(CTApplicationJson)
	}
	return req.UnsafeBodyToJson(to)
}

// Parses the Json request body into the destination. Does not check the
// Content-Type header to confirm that the request body has application/json
// type body. Will panic if the destination is not a pointer.
func (req *Request) UnsafeBodyToJson(to any) error {
	v := reflect.ValueOf(to)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		// We panic for now, this may change. This is due to an issue introduced
		// by the integrator, so we panic which will manifest itself as a 500:
		// Internal Server Error outside of the integrator's control.
		panic(fmt.Sprintf("BodyToJson requires a non-nil pointer destination, got %T", to))
	}
	return json.Unmarshal([]byte(req.body), to)
}

// Parses the text/plain content from the request. This method checks that the
// Content-Type header is set to text/plain, returning a 415: Unsupported Media
// Type error if that is not the case.
func (req *Request) BodyToText() (string, error) {
	if !req.ContentType().Equals(CTTextPlain) {
		return "", UnsupportedMediaTypeError(CTTextPlain)
	}
	return req.body, nil
}

// Access the content type of the request body. Does not return a meaningful
// value for requests without the Content-Type header, or requests that should
// not contain a body, such as GET. The ContentType.Equals method can be used
// to perform equality checks.
func (req *Request) ContentType() ContentType {
	return req.ct
}

// Parses the first line of the request. This line should contain the HTTP
// method, requested URI and the protocol. parseProtocolLine will parse all
// components and group them into a [protocolLine] struct, returning an error
// if the protocol line is malformed.
func parseProtocolLine(raw []byte) (protocolLine, *HttpError) {
	startLineSplit := bytes.Split(raw, []byte(" "))
	if len(startLineSplit) != 3 {
		return protocolLine{}, BadRequestError()
	}

	mthdRaw, uriRaw, prtcl := startLineSplit[0], string(startLineSplit[1]), string(startLineSplit[2])
	mthd, found := methodLookup[string(mthdRaw)]
	if !found {
		return protocolLine{}, NotImplementedError()
	}
	if prtcl != "HTTP/1.1" {
		return protocolLine{}, HttpVersionNotSupportedError()
	}
	if uriRaw == "*" && mthd != OPTIONS {
		// TODO: check error message - should this be a 400 or something else?
		return protocolLine{}, BadRequestError()
	}

	// TODO: need to return 414: URI Too Long if URI is too long

	uri, err := parseUri(uriRaw)
	if err != nil {
		return protocolLine{}, err
	}

	return protocolLine{mthd: mthd, prtcl: prtcl, uri: *uri}, nil
}
