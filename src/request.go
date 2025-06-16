package routeit

import (
	"bytes"
	"io"
	"strings"
)

type Request struct {
	mthd HttpMethod
	url  string
	// TODO: need to change these
	queries    queryParameters
	pathParams pathParameters
	headers    headers
	// TODO: consider byte slice here
	body string
}

func (req *Request) Method() HttpMethod {
	return req.mthd
}

func (req *Request) Url() string {
	return req.url
}

type HttpMethod struct {
	name string
}

var (
	GET = HttpMethod{"GET"}
)

var methodMap = map[string]HttpMethod{
	"GET": GET,
}

func (req *Request) PathParam(param string) (string, bool) {
	val, found := req.pathParams[param]
	return val, found
}

type pathParameters map[string]string
type queryParameters map[string]string

func (req *Request) Header(key string) (string, bool) {
	val, found := req.headers[key]
	return val, found
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
func requestFromRaw(raw []byte) (*Request, error) {
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

	reqHdrs := headersFromRaw(hdrsRaw)

	// TODO: in future, we should verify that this is an allowed host
	_, hasHost := reqHdrs["Host"]
	if !hasHost {
		// The Host header is required as part of HTTP/1.1
		return nil, BadRequestError()
	}

	cLen := reqHdrs.contentLength()
	var body string
	// TODO: prevent parsing if method is GET
	// TODO: make the buffer size also depend on the server max allowed request
	if cLen > 0 {
		reader := bytes.NewReader(bdyRaw)
		buf := make([]byte, cLen)
		_, err := io.ReadFull(reader, buf)
		if err != nil {
			return nil, err
		}
		body = string(buf)
	} else {
		body = ""
	}
	req := Request{mthd: ptcl.mthd, url: endpt, queries: queryParams, pathParams: pathParameters{}, headers: reqHdrs, body: body}
	return &req, nil
}

type protocolLine struct {
	mthd  HttpMethod
	path  string
	prtcl string
}

func parseProtocolLine(raw []byte) (protocolLine, *httpError) {
	split := bytes.Split(raw, []byte(" "))
	if len(split) != 3 {
		return protocolLine{}, BadRequestError()
	}

	mthdRaw, path, prtcl := split[0], string(split[1]), string(split[2])
	mthd, found := methodMap[string(mthdRaw)]
	if !found {
		return protocolLine{}, NotImplementedError()
	}
	if prtcl != "HTTP/1.1" {
		return protocolLine{}, BadRequestError()
	}

	return protocolLine{mthd, path, prtcl}, nil
}

func parseQuery(raw string) (string, queryParameters, *httpError) {
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

	for _, query := range strings.Split(split[1], "&") {
		kvp := strings.Split(query, "=")
		if len(kvp) != 2 {
			return endpoint, nil, BadRequestError()
		}
		queryParams[kvp[0]] = kvp[1]
	}
	return endpoint, queryParams, nil
}
