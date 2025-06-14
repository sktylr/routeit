package routeit

type router struct {
	rreg RouteRegistry
}

func (r *router) registerRoutes(rreg RouteRegistry) {
	r.rreg = rreg
}

func (r *router) route(req *Request) (Handler, bool) {
	hndl, found := r.rreg[req.url]
	if !found {
		hdnl := Handler{}
		return hdnl, found
	}

	if hndl.mthd != req.mthd {
		return hndl, false
	}

	return hndl, found
}

type RouteRegistry map[string]Handler
