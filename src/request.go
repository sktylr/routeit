package routeit

type Request struct {
	mthd       HttpMethod
	url        string
	pathParams pathParameters
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
