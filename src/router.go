package routeit

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/sktylr/routeit/trie"
)

// Matches against strings of the form "/foo /bar # Comment", where "/foo" can
// have any number of path components, but must start with a leading slash and
// not have a trailing slash, and the same for "/bar". Any amount of whitespace
// between the key an value is allowed. The line may optionally end with a
// comment, specific using the "#" character. This can be prefixed optionally
// with any amount of whitespace, though does not have to be. We also allow for
// dynamic rules on both the key and value. These are denoted by ${<name>}
// where <name> is the name of the variable given to the substring so that it
// can be used in the template to rewrite to. Where a key path component is
// dynamic, the component must entirely be encapsulated by ${ }. This regex
// does not prohibit this behaviour, but the key will be incorrectly
// interpreted within the parser if this is the case.
var rewriteParseRe = regexp.MustCompile(`^(/(?:[\w.${}|-]+(?:/[\w.${}|-]+)*)?)\s+(/(?:[\w.${}-]+(?:/[\w.${}-]+)*)?)(?:\s*#.*)?$`)

// The [RouteRegistry] is used to associate routes with their corresponding
// handlers. Routing supports both static and dynamic routes. The keys of the
// [RouteRegistry] represent the route that the handler will be matched
// against. They can optionally include a leading slash and all trailing
// slashes will be stripped. Static routes can be supplied by providing the
// exact path the route should match against. Dynamic routes can be provided
// using a colon-notation. Optional prefixes and/or suffixes can be provided
// for dynamic routes, that will only match against the incoming URI if the
// corresponding path segments start or end with the given prefix or suffix.
// The extracted path parameter contains the entire matched path segment -
// including the prefix and/or suffix. The substring between the prefix and
// suffix must have non-zero length.
//
// Examples:
//   - "/foo/:bar" -> This will match against "/foo/<anything>" and name the
//     first matched parameter "bar".
//   - "/:foo/bar/:baz" -> This will match against "/<anything>/bar/<anything>".
//     The first matched parameter will be named "foo", while the second will
//     be named "baz".
//   - "/:foo|pref" -> This will match against "/pref<anything>".
//   - "/:foo||suf" -> This will match against "/<anything>suf".
//   - "/:foo|pref|suf" -> This will match against "/pref<anything>suf".
//
// Registering routes with dynamic components with the same name (such as
// "/:foo/bar/:foo") will cause the application to panic.
//
// Names parameters can be accessed using [Request.PathParam], providing the
// case-sensitive name of the parameter as provided in the route registration.
type RouteRegistry map[string]Handler

// The [matchedRoute] is a route that has been returned from the routing trie
// that has the relevant handler and the path parameters that have been
// extracted from the path.
type matchedRoute struct {
	handler *Handler
	params  pathParameters
}

type matchedRouteExtractor struct{}

type urlRewriteExtractor struct{}

type router struct {
	routes *trie.StringTrie[Handler, matchedRoute]
	// The global namespace that all registered routes are prefixed with.
	namespace string
	// The static directory for serving responses from disk.
	staticDir    string
	staticLoader *Handler
	rewrites     *trie.StringTrie[string, string]
}

func newRouter() *router {
	return &router{
		routes:   trie.NewStringTrie('/', &matchedRouteExtractor{}),
		rewrites: trie.NewStringTrie('/', &urlRewriteExtractor{}),
	}
}

// Registers the routes to the router. Uses the keys of the map as the path,
// and the value as the handler for the given path.
func (r *router) RegisterRoutes(rreg RouteRegistry) {
	// To save on depth of the trie, we don't need to use the global namespace
	// (if set) when registering routes. We must make sure that lookup accounts
	//  for this namespace however.
	for path, handler := range rreg {
		// When registering routes, we ignore **all* trailing slashes and remove
		// the leading slash. This is different to lookup, where we only ignore
		// the **last** trailing slash if present.
		path = r.trimRouteForInsert(path)
		r.routes.Insert(path, &handler)
	}
}

// Registers routes under a namespace (e.g. /api)
func (r *router) RegisterRoutesUnderNamespace(namespace string, rreg RouteRegistry) {
	namespace = r.trimRouteForInsert(namespace)
	for path, handler := range rreg {
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		r.routes.Insert(namespace+path, &handler)
	}
}

// Registers a global namespace to all routes
func (r *router) GlobalNamespace(namespace string) {
	r.namespace = r.trimRouteForInsert(namespace)
}

// Sets the static directory that files are loaded from. Panics whenever the
// directory is not a subdirectory of the project root but does not require the
// directory to exist when setting it.
func (r *router) NewStaticDir(s string) {
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

// Adds a new URL rewrite rule to the router. Ignores comments and empty lines
// but will panic if the input is malformed, such as an incorrect number of
// values provided, no leading slashes, invalid URI syntax. Additionally, will
// panic if the new rule conflicts with an existing rule. Expects to receive a
// single line containing exactly 1 rewrite rule (or an empty line or comment).
func (r *router) NewRewrite(raw string) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return
	}

	matches := rewriteParseRe.FindStringSubmatch(trimmed)
	if len(matches) < 3 {
		panic(fmt.Errorf("invalid configuration line - not enough entries %#q", raw))
	}

	rawKey, value := matches[1], matches[2]
	if rawKey == value {
		return
	}

	// Rewrite the key from the regex for using ${} to signify variables to the
	// trie form using :
	var kb strings.Builder
	for i, seg := range strings.Split(rawKey, "/") {
		if i == 0 && seg == "" {
			continue
		}
		kb.WriteRune('/')
		if !(strings.HasPrefix(seg, "${") && strings.HasSuffix(seg, "}")) {
			kb.WriteString(seg)
		} else {
			kb.WriteRune(':')
			kb.WriteString(seg[2 : len(seg)-1])
		}
	}
	key := kb.String()

	r.rewrites.Insert(key, &value)
}

// Routes a request to the corresponding handler. A handler may support multiple
// methods, or may not support the method of the request at all, however the
// handler found is known to be the correct handler for the given request URI.
func (r *router) Route(req *Request) (*Handler, bool) {
	if req.Method() == OPTIONS && req.Path() == "*" {
		// The server can respond to OPTIONS * requests, which ask the entire
		// server for options. At this point, the parsed request is guaranteed
		// to only have a path of "*" if the request method is OPTIONS, so we
		// are safe to return the options handler early.
		return globalOptionsHandler(), true
	}

	// TODO: need to improve the string manipulation here - it looks expensive!
	sanitised := strings.TrimPrefix(req.Path(), "/")
	if !strings.HasPrefix(sanitised, r.namespace) {
		// The route is not under the global namespace so we know it isn't valid
		return nil, false
	}

	trimmed := strings.TrimPrefix(sanitised, r.namespace+"/")

	if r.staticDir != "" && strings.HasPrefix(trimmed, r.staticDir) {
		if strings.Contains(trimmed, "..") {
			// We want to prohibit back-tracking, even if it is technically safe
			// (e.g. /foo/bar/../bar/image.png is safe since it can be simplified
			// to /foo/bar/image.png but we don't want to allow back-tracking of
			// any sort)
			return nil, false
		}
		return r.staticLoader, true
	}

	route, found := r.routes.Find(trimmed)
	if route != nil && found {
		req.uri.pathParams = route.params
		return route.handler, true
	}

	return nil, false
}

// Passes the incoming URL through the router's rewrites.
func (r *router) Rewrite(url string) (string, bool) {
	// For static rewrites, the `static` variable is the actual rewrite. For
	// dynamic rewrites, it is the _template_ of the rewrite (i.e. the value
	// used in the config entry)
	rewritten, found := r.rewrites.Find(url)
	if !found {
		return url, false
	}
	return *rewritten, true
}

func (mre *matchedRouteExtractor) NewFromStatic(val *Handler) *matchedRoute {
	return &matchedRoute{handler: val}
}

func (mre *matchedRouteExtractor) NewFromDynamic(val *Handler, path string, re *regexp.Regexp) *matchedRoute {
	params := pathParameters{}
	names := re.SubexpNames()
	matches := re.FindStringSubmatch(path)

	if matches == nil {
		// Indicates that something has gone wrong with the regex or searching.
		return mre.NewFromStatic(val)
	}

	for i, name := range names {
		if i == 0 || name == "" {
			continue
		}
		params[name] = matches[i]
	}

	return &matchedRoute{handler: val, params: params}
}

func (ure *urlRewriteExtractor) NewFromStatic(val *string) *string {
	return val
}

func (ure *urlRewriteExtractor) NewFromDynamic(val *string, path string, re *regexp.Regexp) *string {
	match := re.FindStringSubmatchIndex(path)
	result := string(re.ExpandString(nil, *val, path, match))
	return &result
}

// Removes a single leading slash and any trailing slashes that a route has.
// This method should be used to prepare a route for insertion into a Trie.
func (r *router) trimRouteForInsert(s string) string {
	s = strings.TrimPrefix(s, "/")
	for strings.HasSuffix(s, "/") {
		s = strings.TrimSuffix(s, "/")
	}
	return s
}

// This handler responds to global `OPTIONS *` requests that are asking the
// server for information about the whole server. In this simple solution, we
// just respond with the supported methods for the server through the Allow
// header
func globalOptionsHandler() *Handler {
	return &Handler{options: func(rw *ResponseWriter, req *Request) error {
		// TODO: for testability, this constructs the list deterministically, since map looping is designed to be non-deterministic in go. This should be made more robust
		allowed := []string{
			GET.name,
			HEAD.name,
			POST.name,
			PUT.name,
			DELETE.name,
			PATCH.name,
			OPTIONS.name,
		}
		rw.Header("Allow", strings.Join(allowed, ", "))
		return nil
	}}
}
