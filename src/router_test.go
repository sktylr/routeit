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
	router.registerRoutes(RouteRegistry{
		"/some/route":    Get(doNotWantHandler),
		"/another/route": Get(doNotWantHandler),
		"/want":          Get(wantHandler),
	})
	req := requestWithUrlAndMethod("/want", GET)

	verifyRouteFound(t, router, req)
}

func TestRouteHandlesRepeatedSlashes(t *testing.T) {
	router := newRouter()
	router.registerRoutes(RouteRegistry{
		"/some/route":    Get(doNotWantHandler),
		"/another/route": Get(doNotWantHandler),
		"/want":          Get(wantHandler),
	})
	req := requestWithUrlAndMethod("/some//route", GET)

	verifyRouteNotFound(t, router, req)
}

func TestRouteWithGlobalNamespaceFound(t *testing.T) {
	router := newRouter()
	router.registerRoutes(RouteRegistry{
		"/some/route":    Get(doNotWantHandler),
		"/another/route": Get(doNotWantHandler),
		"/want":          Get(wantHandler),
	})
	router.globalNamespace("/api")
	req := requestWithUrlAndMethod("/api/want", GET)

	verifyRouteFound(t, router, req)
}

func TestRouteWithMultiTieredGlobalNamespace(t *testing.T) {
	router := newRouter()
	router.registerRoutes(RouteRegistry{
		"/some/route":    Get(doNotWantHandler),
		"/another/route": Get(doNotWantHandler),
		"/want":          Get(wantHandler),
	})
	router.globalNamespace("/api/foo")
	req := requestWithUrlAndMethod("/api/foo/want", GET)

	verifyRouteFound(t, router, req)
}

func TestRouteWithGlobalNamespaceNotFound(t *testing.T) {
	router := newRouter()
	router.registerRoutes(RouteRegistry{
		"/some/route":    Get(doNotWantHandler),
		"/another/route": Get(doNotWantHandler),
		"/want":          Get(wantHandler),
	})
	router.globalNamespace("/api")
	req := requestWithUrlAndMethod("/want", GET)

	verifyRouteNotFound(t, router, req)
}

func TestRouteLocalNamespaceFound(t *testing.T) {
	router := newRouter()
	router.registerRoutesUnderNamespace("/api", RouteRegistry{
		"/some/route":    Get(doNotWantHandler),
		"/another/route": Get(doNotWantHandler),
		"/want":          Get(wantHandler),
	})
	req := requestWithUrlAndMethod("/api/want", GET)

	verifyRouteFound(t, router, req)
}

func TestRouteWithMultiTieredLocalNamespace(t *testing.T) {
	router := newRouter()
	router.registerRoutesUnderNamespace("/api/foo", RouteRegistry{
		"/some/route":    Get(doNotWantHandler),
		"/another/route": Get(doNotWantHandler),
		"/want":          Get(wantHandler),
	})
	req := requestWithUrlAndMethod("/api/foo/want", GET)

	verifyRouteFound(t, router, req)
}

func TestRouteLocalNamespaceNotFound(t *testing.T) {
	router := newRouter()
	router.registerRoutesUnderNamespace("/api", RouteRegistry{
		"/some/route":    Get(doNotWantHandler),
		"/another/route": Get(doNotWantHandler),
		"/want":          Get(wantHandler),
	})
	req := requestWithUrlAndMethod("/want", GET)

	verifyRouteNotFound(t, router, req)
}

func TestRouteGlobalAndLocalNamespaceFound(t *testing.T) {
	router := newRouter()
	router.registerRoutesUnderNamespace("/api", RouteRegistry{
		"/some/route":    Get(doNotWantHandler),
		"/another/route": Get(doNotWantHandler),
		"/want":          Get(wantHandler),
	})
	router.globalNamespace("/foo")
	req := requestWithUrlAndMethod("/foo/api/want", GET)

	verifyRouteFound(t, router, req)
}

func TestRouteGlobalAndLocalNamespaceNotFound(t *testing.T) {
	router := newRouter()
	router.registerRoutesUnderNamespace("/api", RouteRegistry{
		"/some/route":    Get(doNotWantHandler),
		"/another/route": Get(doNotWantHandler),
		"/want":          Get(wantHandler),
	})
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
	router.registerRoutes(RouteRegistry{
		"/some/route":    Get(doNotWantHandler),
		"/another/route": Get(doNotWantHandler),
		"/want":          Get(wantHandler),
	})
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
	router.registerRoutes(RouteRegistry{
		"/some/route":    Get(doNotWantHandler),
		"/another/route": Get(doNotWantHandler),
		"/want":          Get(wantHandler),
	})
	req := requestWithUrlAndMethod("/want/", GET)

	verifyRouteFound(t, router, req)
}

func TestGlobalNamespaceEnsuresLeadingSlash(t *testing.T) {
	router := newRouter()
	router.registerRoutes(RouteRegistry{
		"/some/route":    Get(doNotWantHandler),
		"/another/route": Get(doNotWantHandler),
		"/want":          Get(wantHandler),
	})
	router.globalNamespace("api")
	req := requestWithUrlAndMethod("/api/want", GET)

	verifyRouteFound(t, router, req)
}

func TestLocalNamespaceEnsuresLeadingSlash(t *testing.T) {
	router := newRouter()
	router.registerRoutesUnderNamespace("api", RouteRegistry{
		"/some/route":    Get(doNotWantHandler),
		"/another/route": Get(doNotWantHandler),
		"/want":          Get(wantHandler),
	})
	req := requestWithUrlAndMethod("/api/want", GET)

	verifyRouteFound(t, router, req)
}

func TestGlobalNamespaceIgnoresTrailingSlash(t *testing.T) {
	router := newRouter()
	router.registerRoutes(RouteRegistry{
		"/some/route":    Get(doNotWantHandler),
		"/another/route": Get(doNotWantHandler),
		"/want":          Get(wantHandler),
	})
	router.globalNamespace("/api/")
	req := requestWithUrlAndMethod("/api/want", GET)

	verifyRouteFound(t, router, req)
}

func TestLocalNamespaceIgnoresTrailingSlash(t *testing.T) {
	router := newRouter()
	router.registerRoutesUnderNamespace("/api/", RouteRegistry{
		"/some/route":    Get(doNotWantHandler),
		"/another/route": Get(doNotWantHandler),
		"/want":          Get(wantHandler),
	})
	req := requestWithUrlAndMethod("/api/want", GET)

	verifyRouteFound(t, router, req)
}

func TestGlobalNamespaceIgnoresTrailingMultipleSlashes(t *testing.T) {
	router := newRouter()
	router.registerRoutes(RouteRegistry{
		"/some/route":    Get(doNotWantHandler),
		"/another/route": Get(doNotWantHandler),
		"/want":          Get(wantHandler),
	})
	router.globalNamespace("/api//")
	req := requestWithUrlAndMethod("/api/want", GET)

	verifyRouteFound(t, router, req)
}

func TestLocalNamespaceIgnoresTrailingMultipleSlashes(t *testing.T) {
	router := newRouter()
	router.registerRoutesUnderNamespace("/api//", RouteRegistry{
		"/some/route":    Get(doNotWantHandler),
		"/another/route": Get(doNotWantHandler),
		"/want":          Get(wantHandler),
	})
	req := requestWithUrlAndMethod("/api/want", GET)

	verifyRouteFound(t, router, req)
}

func wantHandler(rw *ResponseWriter, req *Request) error {
	return nil
}

func doNotWantHandler(rw *ResponseWriter, req *Request) error {
	return errors.New("did not want this handler")
}

func requestWithUrlAndMethod(url string, method HttpMethod) *Request {
	return &Request{url: url, mthd: method}
}

func verifyRouteFound(t *testing.T, router *router, req *Request) {
	t.Helper()
	got, found := router.route(req)
	if !found {
		t.Error("expected route to be found")
	}
	err := got.fn(&ResponseWriter{}, req)
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
