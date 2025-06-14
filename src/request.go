package routeit

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
