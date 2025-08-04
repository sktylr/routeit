package routeit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
)

var (
	GET     = HttpMethod{name: "GET"}
	HEAD    = HttpMethod{name: "HEAD"}
	POST    = HttpMethod{name: "POST"}
	PUT     = HttpMethod{name: "PUT"}
	DELETE  = HttpMethod{name: "DELETE"}
	PATCH   = HttpMethod{name: "PATCH"}
	OPTIONS = HttpMethod{name: "OPTIONS"}
	TRACE   = HttpMethod{name: "TRACE"}
)

type Request struct {
	ctx       context.Context
	mthd      HttpMethod
	uri       uri
	headers   headers
	body      []byte
	ct        ContentType
	host      string
	userAgent string
	ip        string
	accept    []ContentType
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
func requestFromRaw(raw []byte, maxSize RequestSize, ctx context.Context) (*Request, *HttpError) {
	// TODO: need to add support for all request-target forms (origin-form, absolute-form, authority-form, asterisk-form) that should be accepted by a HTTP only server.
	sections := bytes.Split(raw, []byte("\r\n"))

	// We are expecting 1 carriage return after the protocol line and 1
	// carriage return after all the headers. This means there will be at least
	// 3 sections.
	if len(sections) < 3 {
		return nil, ErrBadRequest()
	}

	prtclRaw := sections[0]
	ptcl, err := parseProtocolLine(prtclRaw)
	if err != nil {
		return nil, err
	}

	hdrsRaw := sections[1:]
	reqHdrs, lastHeader, err := headersFromRaw(hdrsRaw)
	if err != nil {
		return nil, err
	}

	ct := ContentType{}
	ctRaw, hasCType := reqHdrs.Get("Content-Type")
	if hasCType && ptcl.mthd.canHaveBody() {
		ct = parseContentType(ctRaw)
	}
	cLen := reqHdrs.ContentLength()

	if cLen > uint(maxSize) {
		return nil, ErrContentTooLarge()
	}

	if !ct.isValid() && cLen != 0 && ptcl.mthd.canHaveBody() {
		return nil, ErrBadRequest().WithMessage("Cannot specify a Content-Length without Content-Type")
	}

	bdyRaw := bytes.Join(hdrsRaw[lastHeader+1:], []byte("\r\n"))
	var body []byte
	if cLen == 0 || !ptcl.mthd.canHaveBody() {
		// For GET, HEAD or OPTIONS requests, the request body should be
		// ignored even if provided. Servers can technically accept request
		// bodies for OPTIONS requests, however it is up to the server
		// implementation, and routeit chooses not to. Where we are consuming
		// the body, we should only look for Content-Length bytes and no more.
		body = []byte{}
		if ptcl.mthd == TRACE {
			// TRACE requests should not have a body. However, they should
			// return the entire received request in their own response body.
			// To simplify data storage, we will store the raw request on the
			// body property. In reality, the integrator cannot design their
			// own custom handler for TRACE requests, so this difference is not
			// noticeable and easily managed within the framework.
			body = raw
		}
	} else {
		reader := bytes.NewReader(bdyRaw)
		body = make([]byte, cLen)
		_, err := io.ReadFull(reader, body)
		if err != nil {
			// Http servers are expected to read **exactly** Content-Length bytes
			// from the request body. This error is returned if the reader contains
			// **less** than the requested number of bytes, so we cannot read it
			// all. Either the client has not sent it all (e.g. due to a slow
			// connection), or the request is malformed. Return 400 Bad Request
			// since the failure is with the client.
			return nil, ErrBadRequest()
		}
	}

	accept := parseAcceptHeader(reqHdrs)
	userAgent, _ := reqHdrs.Get("User-Agent")
	req := Request{
		mthd:      ptcl.mthd,
		uri:       ptcl.uri,
		headers:   reqHdrs,
		body:      body,
		ct:        ct,
		userAgent: userAgent,
		accept:    accept,
		ctx:       ctx,
	}
	return &req, nil
}

// Access the request's HTTP method
func (req *Request) Method() HttpMethod {
	return req.mthd
}

// The request's URL excluding the host. Does not include query parameters.
// Where the server has URL rewrites configured, this will be the rewritten URL.
// This URL has been escaped correctly, so cannot be used for additional
// routing where dynamic path segment routing is used in the server, as decoded
// "/"s could inadvertently be treated as control characters. If additional
// routing needs to be performed based on the path delimiter, use
// [Request.RawPath].
func (req *Request) Path() string {
	uri := req.uri
	if uri.globalOptions {
		return "*"
	}
	return "/" + strings.Join(req.uri.Path(), "/")
}

// The raw path received at the edge of the server. This is not url-decoded and
// has not been rewritten if URL rewriting is enabled for the server.
func (req *Request) RawPath() string {
	return req.uri.rawPath
}

// Extract a path parameter from the request path. The name must exactly match
// the name of the parameter when it was registered to the router. For example,
// if the route was registered under `/:foO|prefix|suffix`, then this method
// should be called with `"foO"`. Returns a boolean indicating if the path
// parameter was extracted successfully.
func (req *Request) PathParam(param string) (string, bool) {
	val, found := req.uri.pathParams[param]
	return val, found
}

// Access a header value, if present
func (req *Request) Header(key string) (string, bool) {
	val, found := req.headers.Get(key)
	return val, found
}

// The Host header of the request. This will always be present and non-empty
func (req *Request) Host() string {
	return req.host
}

// The User-Agent header of the request. May be empty if not included by the
// client
func (req *Request) UserAgent() string {
	return req.userAgent
}

// The client's IP address that established connection with the server
func (req *Request) ClientIP() string {
	return req.ip
}

// Access a query parameter if present
func (req *Request) QueryParam(key string) (string, bool) {
	// TODO: need additional methods for this
	val, found := req.uri.queryParams[key]
	if found && len(val) > 0 {
		return val[0], true
	}
	return "", false
}

// Parses the Json request body into the destination. Ensures that the
// Content-Type header is application/json and will return a 415: Unsupported
// Media Type error if this is not the case. Will panic if the destination is
// not a pointer. Will also panic if the request cannot contain a request body,
// such as GET requests. Should be preferred to [Request.UnsafeBodyFromJson].
func (req *Request) BodyFromJson(to any) error {
	req.mustAllowBodyReading()
	if !req.ContentType().Matches(CTApplicationJson) {
		return ErrUnsupportedMediaType(CTApplicationJson)
	}
	return req.UnsafeBodyFromJson(to)
}

// Parses the Json request body into the destination. Does not check the
// Content-Type header to confirm that the request body has application/json
// type body. Will panic if the destination is not a pointer, or the request
// cannot support a body (such as GET requests).
func (req *Request) UnsafeBodyFromJson(to any) error {
	v := reflect.ValueOf(to)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		// We panic for now, this may change. This is due to an issue introduced
		// by the integrator, so we panic which will manifest itself as a 500:
		// Internal Server Error outside of the integrator's control.
		panic(fmt.Sprintf("BodyFromJson requires a non-nil pointer destination, got %T", to))
	}
	req.mustAllowBodyReading()
	return json.Unmarshal([]byte(req.body), to)
}

// Parses the text/plain content from the request. This method checks that the
// Content-Type header is set to text/plain, returning a 415: Unsupported Media
// Type error if that is not the case. Panics if the request method is GET,
// since GET requests cannot support bodies. Should be preferred to
// [Request.UnsafeBodyFromText].
func (req *Request) BodyFromText() (string, error) {
	req.mustAllowBodyReading()
	if !req.ContentType().Matches(CTTextPlain) {
		return "", ErrUnsupportedMediaType(CTTextPlain)
	}
	return string(req.body), nil
}

// Returns the raw body content as a string. Will panic if this is called on a
// method that cannot support a request body, such as GET, HEAD or OPTIONS.
func (req *Request) UnsafeBodyFromText() string {
	req.mustAllowBodyReading()
	return string(req.body)
}

// Returns the raw body content provided it matches the given content type,
// otherwise a 415: Unsupported Media Type error is returned. Will panic if
// this is called on a method that cannot support a request body, such as GET,
// HEAD or OPTIONS. Should be preferred to [Request.UnsafeBodyFromRaw].
func (req *Request) BodyFromRaw(ct ContentType) ([]byte, error) {
	req.mustAllowBodyReading()
	if !req.ContentType().Matches(ct) {
		return nil, ErrUnsupportedMediaType(ct)
	}
	return req.body, nil
}

// Returns the raw body content. Will panic if this is called on a method that
// cannot support a request body, such as GET, HEAD or OPTIONS. This does not
// assert that the body is present nor contains a corresponding Content-Type
// header
func (req *Request) UnsafeBodyFromRaw() []byte {
	req.mustAllowBodyReading()
	return req.body
}

// Access the content type of the request body. Does not return a meaningful
// value for requests without the Content-Type header, or requests that should
// not contain a body, such as GET. The ContentType.Equals method can be used
// to perform equality checks.
func (req *Request) ContentType() ContentType {
	return req.ct
}

// Can be used to determine whether the client will accept the provided
// [ContentType]
func (req *Request) AcceptsContentType(other ContentType) bool {
	for _, ct := range req.accept {
		if ct.Matches(other) {
			return true
		}
	}
	return false
}

// Access the request's context. This should be used for context aware
// calculations within handlers or middleware, for example to load an object
// from a database.
func (req *Request) Context() context.Context {
	return req.ctx
}

// Access a value stored on the request's context.
func (req *Request) ContextValue(key any) (any, bool) {
	val := req.ctx.Value(key)
	return val, val != nil
}

// Store a value on the request's context. This is particularly useful for
// middleware, which can set particular values the can be read in later
// middleware or the handler itself.
func (req *Request) NewContextValue(key any, val any) {
	req.ctx = context.WithValue(req.ctx, key, val)
}

func (req *Request) mustAllowBodyReading() {
	if !req.mthd.canHaveBody() {
		panic(fmt.Errorf("attempted to read body for request that cannot contain a body - method = %s", req.mthd.name))
	}
}

func (m HttpMethod) canHaveBody() bool {
	return m != GET && m != HEAD && m != OPTIONS && m != TRACE
}

func (m HttpMethod) isValid() bool {
	switch m {
	case GET, HEAD, POST, PUT, DELETE, PATCH, OPTIONS, TRACE:
		return true
	default:
		return false
	}
}

// Parses the first line of the request. This line should contain the HTTP
// method, requested URI and the protocol. parseProtocolLine will parse all
// components and group them into a [protocolLine] struct, returning an error
// if the protocol line is malformed.
func parseProtocolLine(raw []byte) (protocolLine, *HttpError) {
	startLineSplit := bytes.Split(raw, []byte(" "))
	if len(startLineSplit) != 3 {
		return protocolLine{}, ErrBadRequest()
	}

	mthdRaw, uriRaw, prtcl := startLineSplit[0], string(startLineSplit[1]), string(startLineSplit[2])
	mthd := HttpMethod{name: string(mthdRaw)}
	if !mthd.isValid() {
		return protocolLine{}, ErrNotImplemented()
	}
	if prtcl != "HTTP/1.1" {
		return protocolLine{}, ErrHttpVersionNotSupported()
	}
	if uriRaw == "*" && mthd != OPTIONS {
		return protocolLine{}, ErrBadRequest().WithMessage("Invalid request-target for method")
	}

	if RequestSize(len(uriRaw)) > (8 * KiB) {
		return protocolLine{}, ErrURITooLong()
	}

	uri, err := parseUri(uriRaw)
	if err != nil {
		return protocolLine{}, err
	}

	return protocolLine{mthd: mthd, prtcl: prtcl, uri: *uri}, nil
}
