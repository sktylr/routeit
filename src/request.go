package routeit

import (
	"bytes"
	"fmt"
	"strings"
)

type Request struct {
	mthd       HttpMethod
	url        string
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

func parseMethod(mthdRaw string) (HttpMethod, bool) {
	mthd, found := methodMap[mthdRaw]
	return mthd, found
}

func (req *Request) PathParam(param string) (string, bool) {
	val, found := req.pathParams[param]
	return val, found
}

type pathParameters map[string]string

func (req *Request) Header(key string) (string, bool) {
	val, found := req.headers[key]
	return val, found
}

func requestFromRaw(raw []byte) (*Request, error) {
	lines := bytes.Split(raw, []byte("\n"))
	ptcl := bytes.SplitN(bytes.TrimSpace(lines[0]), []byte(" "), 3)
	if len(ptcl) != 3 {
		fmt.Print("Unexpected HTTP protocol line!\n")
		return nil, fmt.Errorf("unexpected http protocol line: %s", ptcl)
	}

	ver := string(ptcl[2])
	if ver != "HTTP/1.1" {
		// TODO: should make this a response object instead
		fmt.Print("Unsupported HTTP version!\n")
		return nil, fmt.Errorf("unsupported http version: %s", ver)
	}

	mthd, found := parseMethod(string(ptcl[0]))
	if !found {
		fmt.Print("Unsupported HTTP Method!\n")
		return nil, fmt.Errorf("unsupported http method: %s", ptcl[0])
	}

	path := string(ptcl[1])
	pathParams := pathParameters{}
	foo := strings.Split(path, "?")
	endpt := foo[0]
	if len(foo) > 1 {
		if len(foo) > 2 {
			// TODO: handle this better
			fmt.Print("Unexpected number of query options!\n")
		}

		queries := foo[1]
		for _, query := range strings.Split(queries, "&") {
			kvp := strings.SplitN(query, "=", 2)
			if len(kvp) != 2 {
				fmt.Print("Query string malformed!\n")
				continue
			}
			pathParams[kvp[0]] = kvp[1]
		}
	}

	reqHdrs := headers{}
	var end int
	for i, line := range lines {
		// ?????
		end = i
		if i == 0 {
			continue
		}
		sline := strings.TrimSpace(string(line))
		if sline == "" {
			// Blank line between headers and body
			break
		}

		kvp := strings.SplitN(sline, ": ", 2)
		if len(kvp) != 2 {
			fmt.Printf("Malformed header: [%s]\n", sline)
			continue
		}
		reqHdrs[kvp[0]] = kvp[1]
	}
	var sb strings.Builder
	for _, line := range lines[end:] {
		sb.Write(bytes.TrimSpace(line))
	}
	req := Request{mthd: mthd, url: endpt, pathParams: pathParams, headers: reqHdrs, body: sb.String()}
	return &req, nil
}
