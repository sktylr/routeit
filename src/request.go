package routeit

import (
	"bytes"
	"io"
	"strings"
)

var (
	GET  = HttpMethod{"GET"}
	HEAD = HttpMethod{"HEAD"}
)

var methodLookup = map[string]HttpMethod{
	"GET":  GET,
	"HEAD": HEAD,
}

type Request struct {
	mthd HttpMethod
	// TODO: need to normalise this properly and have a (private) method for trimming the namespace etc.
	url string
	// TODO: need to change these
	queries    queryParameters
	pathParams pathParameters
	headers    headers
	// TODO: consider byte slice here
	body string
}

type HttpMethod struct {
	name string
}

type pathParameters map[string]string
type queryParameters map[string]string

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
	bdyRaw := sections[len(sections)-1]

	ptcl, err := parseProtocolLine(prtclRaw)
	if err != nil {
		return nil, err
	}

	endpt, queryParams, err := parseQuery(ptcl.path)
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
	// TODO: prevent parsing if method is GET or HEAD
	// TODO: make the buffer size also depend on the server max allowed request
	if cLen > 0 {
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
	} else {
		body = ""
	}
	req := Request{mthd: ptcl.mthd, url: endpt, queries: queryParams, pathParams: pathParameters{}, headers: reqHdrs, body: body}
	return &req, nil
}

// Access the request's HTTP method
func (req *Request) Method() HttpMethod {
	return req.mthd
}

// The request's URL excluding the host
func (req *Request) Url() string {
	return req.url
}

func (req *Request) PathParam(param string) (string, bool) {
	val, found := req.pathParams[param]
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
	val, found := req.queries[key]
	return val, found
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

	return protocolLine{mthd, path, prtcl}, nil
}

func parseQuery(raw string) (string, queryParameters, *HttpError) {
	split := strings.Split(raw, "?")
	endpoint := split[0]
	queryParams := queryParameters{}

	if len(split) == 1 {
		// No query string present
		return endpoint, queryParameters{}, nil
	}

	if len(split) > 2 {
		return endpoint, nil, BadRequestError()
	}

	for query := range strings.SplitSeq(split[1], "&") {
		kvp := strings.Split(query, "=")
		if len(kvp) != 2 {
			return endpoint, nil, BadRequestError()
		}
		queryParams[kvp[0]] = kvp[1]
	}
	return endpoint, queryParams, nil
}
