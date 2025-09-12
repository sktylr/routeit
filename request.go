package routeit

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
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
	headers   *RequestHeaders
	body      []byte
	ct        ContentType
	host      string
	userAgent string
	ip        string
	accept    []ContentType
	id        string
	tlsState  *tls.ConnectionState
}

type HttpMethod struct {
	name string
}

type requestLine struct {
	mthd  HttpMethod
	prtcl string
	uri   uri
}

// Parses the raw byte slice of the request into a more usable request structure
//
// The request is made up of three components: the request line, headers and the
// body. For HTTP/1.1, at a bare minimum the Host header must be included, though
// the body is optional (and ignored for certain request methods such as GET).
//
// Each section is split using carriage returns (CRLF or \r\n). After the request
// line, and each header line is also split using a carriage return. The request
// line is always only a single line made up of three components - the request
// method, the path (or URI) and the HTTP protocol, and a blank line (using a
// carriage return) also follows the headers before the optional body.
//
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Guides/Messages
func requestFromRaw(raw []byte, maxSize RequestSize, ctx context.Context) (*Request, *HttpError) {
	sections := bytes.Split(raw, []byte("\r\n"))

	// We are expecting 1 carriage return after the request line and 1
	// carriage return after all the headers. This means there will be at least
	// 3 sections.
	if len(sections) < 3 {
		return nil, ErrBadRequest()
	}

	prtclRaw := sections[0]
	reqLine, err := parseRequestLine(prtclRaw)
	if err != nil {
		return nil, err
	}

	hdrsRaw := sections[1:]
	reqHdrs, lastHeader, err := headersFromRaw(hdrsRaw)
	if err != nil {
		return nil, err
	}

	ct := ContentType{}
	ctRaw, hasCType := reqHdrs.First("Content-Type")
	if hasCType && reqLine.mthd.canHaveBody() {
		ct = parseContentType(ctRaw)
	}
	cLen := reqHdrs.headers.ContentLength()

	if cLen > uint(maxSize) {
		return nil, ErrContentTooLarge()
	}

	if !ct.isValid() && cLen != 0 && reqLine.mthd.canHaveBody() {
		return nil, ErrBadRequest().WithMessage("Cannot specify a Content-Length without Content-Type")
	}

	bdyRaw := bytes.Join(hdrsRaw[lastHeader+1:], []byte("\r\n"))
	var body []byte
	if cLen == 0 || !reqLine.mthd.canHaveBody() {
		// For GET, HEAD or OPTIONS requests, the request body should be
		// ignored even if provided. Servers can technically accept request
		// bodies for OPTIONS requests, however it is up to the server
		// implementation, and routeit chooses not to. Where we are consuming
		// the body, we should only look for Content-Length bytes and no more.
		body = []byte{}
		if reqLine.mthd == TRACE {
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
	userAgent, _ := reqHdrs.First("User-Agent")
	req := Request{
		mthd:      reqLine.mthd,
		uri:       reqLine.uri,
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
// should be called with `"foO"`. This will always be a non-empty value
// corresponding to the named path segment in the request, unless the name does
// not match any segments, in which case it will be empty.
func (req *Request) PathParam(param string) string {
	return req.uri.pathParams[param]
}

// Access the headers of the request.
func (req *Request) Headers() *RequestHeaders {
	return req.headers
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

// The ID of the request. This will only be populated if an implementation is
// provided to [ServerConfig.RequestIdProvider].
func (req *Request) Id() string {
	return req.id
}

// The request's TLS connection state. This is nil when the request was
// received over HTTP and non-nil whenever the request was made using HTTPS.
// It contains details about the TLS connection, such as the version and cipher
// suites.
func (req *Request) Tls() *tls.ConnectionState {
	return req.tlsState
}

// Access the query parameters of the request URI. This will always return a
// non-nil pointer, even if the URI contains no query parameters. See the
// [QueryParams] type for access methods to retrieve individual keys.
func (req *Request) Queries() *QueryParams {
	return req.uri.queryParams
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
	err := json.Unmarshal([]byte(req.body), to)
	if err == nil {
		return nil
	}
	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return ErrBadRequest().WithCause(err).WithMessage("Failed to parse JSON request body.")
	}
	return err
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
// not contain a body, such as GET. The [ContentType.Matches] method can be
// used to determine whether two content types are equivalent.
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

// [ContextValueAs] is a shorthand to casting and using [Request.ContextValue]
// It will return a value and true whenever the request context has a key that
// matches the generic type, and will return false if either the context does
// not have a matching key, or the value stored under that key does not match
// the generic type.
func ContextValueAs[T any](req *Request, key any) (T, bool) {
	val := req.ctx.Value(key)
	if val == nil {
		var z T
		return z, false
	}
	casted, ok := val.(T)
	return casted, ok
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
// method, requested URI and the protocol. parseRequestLine will parse all
// components and group them into a [requestLine] struct, returning an error if
// the request line is malformed. The URI is only treated as being origin-form
// (the most common - e.g. "/foo/bar?baz=qux") or asterisk-form (used for
// global OPTIONS requests - "*"). There are two other formats the request can
// appear in - absolute-form ("http://example.com/") and authority-form
// ("www.example.com:80"). Authority-form is only used for CONNECT requests and
// absolute-form is used for non-CONNECT requests to a proxy. Since routeit
// does not support CONNECT requests and is not intended to be used as a proxy,
// we don't support absolute- or authority-form explicitly.
func parseRequestLine(raw []byte) (requestLine, *HttpError) {
	startLineSplit := bytes.Split(raw, []byte(" "))
	if len(startLineSplit) != 3 {
		return requestLine{}, ErrBadRequest()
	}

	mthdRaw, uriRaw, prtcl := startLineSplit[0], string(startLineSplit[1]), string(startLineSplit[2])
	mthd := HttpMethod{name: string(mthdRaw)}
	if !mthd.isValid() {
		return requestLine{}, ErrNotImplemented()
	}
	if prtcl != "HTTP/1.1" {
		return requestLine{}, ErrHttpVersionNotSupported()
	}
	if uriRaw == "*" && mthd != OPTIONS {
		return requestLine{}, ErrBadRequest().WithMessage("Invalid request-target for method")
	}

	if RequestSize(len(uriRaw)) > (8 * KiB) {
		return requestLine{}, ErrURITooLong()
	}

	uri, err := parseUri(uriRaw)
	if err != nil {
		return requestLine{}, err
	}

	return requestLine{mthd: mthd, prtcl: prtcl, uri: *uri}, nil
}
