package routeit

import (
	"errors"
	"testing"
)

func TestRouteEmpty(t *testing.T) {
	router := newRouter()
	req := requestWithUrlAndMethod("/want", GET)

	verifyRouteNotFound(t, router, req)
}

func TestRouteOneRoute(t *testing.T) {
	router := newRouter()
	router.registerRoutes(RouteRegistry{
		"/want": Get(wantHandler),
	})
	req := requestWithUrlAndMethod("/want", GET)

	verifyRouteFound(t, router, req)
}

func TestRouteValidRoute(t *testing.T) {
	router := newRouter()
	router.registerRoutes(defaultRouteRegistry())
	req := requestWithUrlAndMethod("/want", GET)

	verifyRouteFound(t, router, req)
}

func TestRouteHandlesRepeatedSlashes(t *testing.T) {
	router := newRouter()
	router.registerRoutes(defaultRouteRegistry())
	req := requestWithUrlAndMethod("/some//route", GET)

	verifyRouteNotFound(t, router, req)
}

func TestRouteWithGlobalNamespaceFound(t *testing.T) {
	router := newRouter()
	router.registerRoutes(defaultRouteRegistry())
	router.globalNamespace("/api")
	req := requestWithUrlAndMethod("/api/want", GET)

	verifyRouteFound(t, router, req)
}

func TestRouteWithMultiTieredGlobalNamespace(t *testing.T) {
	router := newRouter()
	router.registerRoutes(defaultRouteRegistry())
	router.globalNamespace("/api/foo")
	req := requestWithUrlAndMethod("/api/foo/want", GET)

	verifyRouteFound(t, router, req)
}

func TestRouteWithGlobalNamespaceNotFound(t *testing.T) {
	router := newRouter()
	router.registerRoutes(defaultRouteRegistry())
	router.globalNamespace("/api")
	req := requestWithUrlAndMethod("/want", GET)

	verifyRouteNotFound(t, router, req)
}

func TestRouteLocalNamespaceFound(t *testing.T) {
	router := newRouter()
	router.registerRoutesUnderNamespace("/api", defaultRouteRegistry())
	req := requestWithUrlAndMethod("/api/want", GET)

	verifyRouteFound(t, router, req)
}

func TestRouteWithMultiTieredLocalNamespace(t *testing.T) {
	router := newRouter()
	router.registerRoutesUnderNamespace("/api/foo", defaultRouteRegistry())
	req := requestWithUrlAndMethod("/api/foo/want", GET)

	verifyRouteFound(t, router, req)
}

func TestRouteLocalNamespaceNotFound(t *testing.T) {
	router := newRouter()
	router.registerRoutesUnderNamespace("/api", defaultRouteRegistry())
	req := requestWithUrlAndMethod("/want", GET)

	verifyRouteNotFound(t, router, req)
}

func TestRouteGlobalAndLocalNamespaceFound(t *testing.T) {
	router := newRouter()
	router.registerRoutesUnderNamespace("/api", defaultRouteRegistry())
	router.globalNamespace("/foo")
	req := requestWithUrlAndMethod("/foo/api/want", GET)

	verifyRouteFound(t, router, req)
}

func TestRouteGlobalAndLocalNamespaceNotFound(t *testing.T) {
	router := newRouter()
	router.registerRoutesUnderNamespace("/api", defaultRouteRegistry())
	router.globalNamespace("/foo")

	verifyRouteNotFound(t, router, requestWithUrlAndMethod("/api/foo/want", GET))
	verifyRouteNotFound(t, router, requestWithUrlAndMethod("/api/want", GET))
	verifyRouteNotFound(t, router, requestWithUrlAndMethod("/foo/want", GET))
}

func TestRegistrationEnsuresLeadingSlash(t *testing.T) {
	router := newRouter()
	router.registerRoutes(RouteRegistry{
		"some/route":    Get(doNotWantHandler),
		"another/route": Get(doNotWantHandler),
		"want":          Get(wantHandler),
	})
	req := requestWithUrlAndMethod("/want", GET)

	verifyRouteFound(t, router, req)
}

func TestLookupEnsuresLeadingSlash(t *testing.T) {
	router := newRouter()
	router.registerRoutes(defaultRouteRegistry())
	req := requestWithUrlAndMethod("want", GET)

	verifyRouteFound(t, router, req)
}

func TestRegistrationIgnoresTrailingSlash(t *testing.T) {
	router := newRouter()
	router.registerRoutes(RouteRegistry{
		"/some/route/":    Get(doNotWantHandler),
		"/another/route/": Get(doNotWantHandler),
		"/want/":          Get(wantHandler),
	})
	req := requestWithUrlAndMethod("/want", GET)

	verifyRouteFound(t, router, req)
}

func TestLookupIgnoresTrailingSlash(t *testing.T) {
	router := newRouter()
	router.registerRoutes(defaultRouteRegistry())
	req := requestWithUrlAndMethod("/want/", GET)

	verifyRouteFound(t, router, req)
}

func TestGlobalNamespaceEnsuresLeadingSlashOnNamespace(t *testing.T) {
	router := newRouter()
	router.registerRoutes(defaultRouteRegistry())
	router.globalNamespace("api")
	req := requestWithUrlAndMethod("/api/want", GET)

	verifyRouteFound(t, router, req)
}

func TestGlobalNamespaceEnsuresLeadingSlashOnPaths(t *testing.T) {
	router := newRouter()
	router.registerRoutes(RouteRegistry{
		"some/route":    Get(doNotWantHandler),
		"another/route": Get(doNotWantHandler),
		"want":          Get(wantHandler),
	})
	router.globalNamespace("/api")
	req := requestWithUrlAndMethod("/api/want", GET)

	verifyRouteFound(t, router, req)
}

func TestLocalNamespaceEnsuresLeadingSlashOnNamespace(t *testing.T) {
	router := newRouter()
	router.registerRoutesUnderNamespace("api", defaultRouteRegistry())
	req := requestWithUrlAndMethod("/api/want", GET)

	verifyRouteFound(t, router, req)
}

func TestLocalNamespaceEnsuresLeadingSlashOnPaths(t *testing.T) {
	router := newRouter()
	router.registerRoutesUnderNamespace("/api", RouteRegistry{
		"some/route":    Get(doNotWantHandler),
		"another/route": Get(doNotWantHandler),
		"want":          Get(wantHandler),
	})
	req := requestWithUrlAndMethod("/api/want", GET)

	verifyRouteFound(t, router, req)
}

func TestGlobalNamespaceIgnoresTrailingSlash(t *testing.T) {
	router := newRouter()
	router.registerRoutes(defaultRouteRegistry())
	router.globalNamespace("/api/")
	req := requestWithUrlAndMethod("/api/want", GET)

	verifyRouteFound(t, router, req)
}

func TestLocalNamespaceIgnoresTrailingSlash(t *testing.T) {
	router := newRouter()
	router.registerRoutesUnderNamespace("/api/", defaultRouteRegistry())
	req := requestWithUrlAndMethod("/api/want", GET)

	verifyRouteFound(t, router, req)
}

func TestGlobalNamespaceIgnoresTrailingMultipleSlashes(t *testing.T) {
	router := newRouter()
	router.registerRoutes(defaultRouteRegistry())
	router.globalNamespace("/api//")
	req := requestWithUrlAndMethod("/api/want", GET)

	verifyRouteFound(t, router, req)
}

func TestLocalNamespaceIgnoresTrailingMultipleSlashes(t *testing.T) {
	router := newRouter()
	router.registerRoutesUnderNamespace("/api//", defaultRouteRegistry())
	req := requestWithUrlAndMethod("/api/want", GET)

	verifyRouteFound(t, router, req)
}

func TestStaticRoutingFound(t *testing.T) {
	router := newRouter()
	router.newStaticDir("static")
	req := requestWithUrlAndMethod("/static", GET)

	// We don't want to actually load the file from disk in the test, so we
	// only assert on the presence of the routing. This works for these tests
	// since we have no other handlers registered, meaning that if we found
	// this one it must be the static loader.
	_, found := router.route(req)
	if !found {
		t.Error("expected to find static router")
	}
}

func TestStaticRoutingGlobalNamespaceFound(t *testing.T) {
	router := newRouter()
	router.globalNamespace("/api")
	router.newStaticDir("static")
	req := requestWithUrlAndMethod("/api/static", GET)

	// We don't want to actually load the file from disk in the test, so we
	// only assert on the presence of the routing. This works for these tests
	// since we have no other handlers registered, meaning that if we found
	// this one it must be the static loader.
	_, found := router.route(req)
	if !found {
		t.Error("expected to find static router")
	}
}

func TestStaticRoutingNotFoundBacktracking(t *testing.T) {
	router := newRouter()
	router.newStaticDir("static")
	req := requestWithUrlAndMethod("/static/../main.go", GET)

	verifyRouteNotFound(t, router, req)
}

func TestStaticRoutingGlobalNamespaceNotFoundBacktracking(t *testing.T) {
	router := newRouter()
	router.globalNamespace("/api")
	router.newStaticDir("static")
	req := requestWithUrlAndMethod("/api/static/../main.go", GET)

	verifyRouteNotFound(t, router, req)
}

func TestStaticDirEnforcesSubdirectory(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{
			"root directory", "~/foo/bar",
		},
		{
			"containing variables", "$HOME/bar",
		},
		{
			"backtracking outside", "static/../..",
		},
		{
			"project root", "static/..",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := newRouter()

			defer func() {
				if r := recover(); r == nil {
					t.Errorf("router invalid static dir, expected panic but got none - static = %#q", router.staticDir)
				}
			}()

			router.newStaticDir(tc.in)
		})
	}
}

func TestStaticDirSimplifiesExpressions(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			// The leading slash should be understood to mean ./ since we don't
			// allow access outside of the project root.
			"leading slash", "/static", "static",
		},
		{
			"useless backtrack", "static/../static/../static", "static",
		},
		{
			"cyclic", "static/foo/../../static/../static/foo/../foo", "static/foo",
		},
		{
			// The `~` character is only expanded to the system root within the
			// shell (it is a shorthand), but it is a perfectly legal, albeit
			// confusing, directory name. This should not be expanded to the
			// computer root.
			"containing root shorthand", "static/~/foo", "static/~/foo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := newRouter()
			router.newStaticDir(tc.in)
			if router.staticDir != tc.want {
				t.Errorf(`router.static = %q, wanted %#q`, router.staticDir, tc.want)
			}
		})
	}
}

func defaultRouteRegistry() RouteRegistry {
	return RouteRegistry{
		"/some/route":    Get(doNotWantHandler),
		"/another/route": Get(doNotWantHandler),
		"/want":          Get(wantHandler),
	}
}

// For handlers we do want, return nil which will satisfy the tests.
func wantHandler(rw *ResponseWriter, req *Request) error {
	return nil
}

// For any handler we don't want, return an error which will fail an assertion
// in the test.
func doNotWantHandler(rw *ResponseWriter, req *Request) error {
	return errors.New("did not want this handler")
}

func requestWithUrlAndMethod(url string, method HttpMethod) *Request {
	return &Request{uri: uri{url: url}, mthd: method}
}

func verifyRouteFound(t *testing.T, router *router, req *Request) {
	t.Helper()
	got, found := router.route(req)
	if !found {
		t.Error("expected route to be found")
	}
	err := got.handle(&ResponseWriter{}, req)
	if err != nil {
		t.Errorf("did not expect handler to error: %s", err.Error())
	}
}

func verifyRouteNotFound(t *testing.T, router *router, req *Request) {
	_, found := router.route(req)
	if found {
		t.Errorf("did not expect to find a route for [url=%s, method=%s]", req.Url(), req.Method().name)
	}
}
