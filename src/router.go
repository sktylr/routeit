package routeit

type RouteRegistry map[string]Handler

type route struct {
	Get Handler
}

type router struct {
	routes *trie[route]
}

func newRouter() *router {
	return &router{routes: newTrie[route]()}
}

func (r *router) registerRoutes(rreg RouteRegistry) {
	for path, handler := range rreg {
		// For now we only support GET requests
		r.routes.insert(path, &route{Get: handler})
	}
}

func (r *router) route(req *Request) (Handler, bool) {
	handler, found := r.routes.find(req.url)
	if handler != nil && found {
		return handler.Get, true
	}
	return Handler{}, false
}
