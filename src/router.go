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
		// TODO: need to improve the string manipulation here - it looks expensive!
		// When registering routes, we ignore **all* trailing slashes and remove
		// the leading slash. This is different to lookup, where we only ignore
		// the **last** trailing slash if present.
		path = strings.TrimPrefix(path, "/")
		for strings.HasSuffix(path, "/") {
			path = strings.TrimSuffix(path, "/")
		}
		// For now we only support GET requests
		r.routes.insert(path, &route{Get: handler})
	}
}

func (r *router) registerRoutesUnderNamespace(namespace string, rreg RouteRegistry) {
	// TODO: need to improve the string manipulation here - it looks expensive!
	namespace = strings.TrimPrefix(namespace, "/")
	for strings.HasSuffix(namespace, "/") {
		namespace = strings.TrimSuffix(namespace, "/")
	}
	for path, handler := range rreg {
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		r.routes.insert(namespace+path, &route{Get: handler})
	}
}

// Registers a global namespace to all routes
func (r *router) globalNamespace(namespace string) {
	// TODO: need to improve the string manipulation here - it looks expensive!
	namespace = strings.TrimPrefix(namespace, "/")
	for strings.HasSuffix(namespace, "/") {
		namespace = strings.TrimSuffix(namespace, "/")
	}
	r.namespace = namespace

}

func (r *router) route(req *Request) (*route, bool) {
	// TODO: need to improve the string manipulation here - it looks expensive!
	sanitised := strings.TrimSuffix(strings.TrimPrefix(req.Url(), "/"), "/")
	if !strings.HasPrefix(sanitised, r.namespace) {
		// The route is not under the global namespace so we know it isn't valid
		return nil, false
	}

	trimmed := strings.TrimPrefix(sanitised, r.namespace+"/")
	route, found := r.routes.find(trimmed)
	if route != nil && found {
		return route, true
	}
	return nil, false
}

func (r *route) supportsMethod(m HttpMethod) bool {
	switch m {
	case GET, HEAD:
		return r.Get.fn != nil
	}
	return false
}

// Dispatches the request for the given route, choosing the handler based on
// the request's Http method.
func (r *route) dispatch(rw *ResponseWriter, req *Request) error {
	if !r.supportsMethod(req.Method()) {
		return MethodNotAllowedError()
	}
	if req.Method() == GET {
		return r.Get.fn(rw, req)
	}
	if req.Method() == HEAD {
		// The HEAD method is the same as GET, except it does not return a
		// response body, only headers. It is often used to determine how
		// large a resource is before committing to downloading it.
		err := r.Get.fn(rw, req)
		rw.bdy = []byte{}
		return err
	}
	return MethodNotAllowedError()
}
