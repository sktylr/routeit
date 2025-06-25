package routeit

import "strings"

type RouteRegistry map[string]Handler

type route struct {
	Get Handler
}

type router struct {
	routes    *trie[route]
	namespace string
}

func newRouter() *router {
	return &router{routes: newTrie[route]()}
}

func (r *router) registerRoutes(rreg RouteRegistry) {
	// To save on depth of the trie, we don't need to use the global namespace
	// (if set) when registering routes. We must make sure that lookup accounts
	//  for this namespace however.
	for path, handler := range rreg {
		// For now we only support GET requests
		r.routes.insert(path, &route{Get: handler})
	}
}

// Registers a global namespace to all routes
func (r *router) globalNamespace(namespace string) {
	if !strings.HasPrefix(namespace, "/") {
		r.namespace = "/" + namespace
	} else {
		r.namespace = namespace
	}
}

func (r *router) route(req *Request) (Handler, bool) {
	url := req.Url()
	if !strings.HasPrefix(url, r.namespace) {
		// The route is not under the global namespace so we know it isn't valid
		return Handler{}, false
	}

	trimmed := strings.TrimPrefix(url, r.namespace)
	handler, found := r.routes.find(trimmed)
	if handler != nil && found {
		return handler.Get, true
	}
	return Handler{}, false
}
