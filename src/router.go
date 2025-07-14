package routeit

import (
	"fmt"
	"path"
	"regexp"
	"strings"
)

type RouteRegistry map[string]Handler

type router struct {
	routes *trie[Handler]
	// The global namespace that all registered routes are prefixed with.
	namespace string
	// The static directory for serving responses from disk.
	staticDir    string
	staticLoader *Handler
}

func newRouter() *router {
	return &router{routes: newTrie[Handler]()}
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
		r.routes.insert(path, &handler)
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
		r.routes.insert(namespace+path, &handler)
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

// Sets the static directory that files are loaded from. Panics whenever the
// directory is not a subdirectory of the project root but does not require the
// directory to exist when setting it.
func (r *router) newStaticDir(s string) {
	if s == "" {
		return
	}
	cleaned := path.Clean(s)
	// This is only run ~once per program execution so we don't need to fuss
	// too much about optimising by pre-compiling etc.
	re := regexp.MustCompile(`(^~|\.{2}|\$)`)
	if re.MatchString(cleaned) || cleaned == "." {
		panic(fmt.Sprintf("invalid static assets directory [%s] - must not be outside project root", s))
	}
	cleaned = strings.TrimPrefix(cleaned, "/")
	r.staticDir = cleaned
	r.staticLoader = staticLoader(r.namespace)
}

// Routes a request to the corresponding handler. A handler may support multiple
// methods, or may not support the method of the request at all, however the
// handler found is known to be the correct handler for the given request URI.
func (r *router) route(req *Request) (*Handler, bool) {
	// TODO: need to improve the string manipulation here - it looks expensive!
	sanitised := strings.TrimSuffix(strings.TrimPrefix(req.Url(), "/"), "/")
	if !strings.HasPrefix(sanitised, r.namespace) {
		// The route is not under the global namespace so we know it isn't valid
		return nil, false
	}

	trimmed := strings.TrimPrefix(sanitised, r.namespace+"/")

	if r.staticDir != "" && strings.HasPrefix(trimmed, r.staticDir) {
		if strings.Contains(trimmed, "..") {
			// We want to prohibit back-tracking, even if it is technically safe
			// (e.g. /foo/bar/../bar/image.png)
			return nil, false
		}
		return r.staticLoader, true
	}

	route, found := r.routes.find(trimmed)
	if route != nil && found {
		return route, true
	}
	return nil, false
}
