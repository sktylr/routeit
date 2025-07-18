package routeit

import (
	"errors"
	"testing"
)

type RouteTest struct {
	name           string
	gNamespace     string
	lNamespace     string
	reg            RouteRegistry
	staticDir      string
	path           string
	wantPathParams pathParameters
}

func TestRoute(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		tests := []RouteTest{
			{
				name: "one route no namespace",
				reg: RouteRegistry{
					"/want": Get(wantHandler),
				},
				path: "/want",
			},
			{
				name: "multiple, valid route no namespace",
				path: "/want",
			},
			{
				name: "registration forces leading slash",
				reg: RouteRegistry{
					"some/route":    Get(doNotWantHandler),
					"another/route": Get(doNotWantHandler),
					"want":          Get(wantHandler),
				},
				path: "/want",
			},
			{
				name: "lookup ensures leading slash",
				path: "want",
			},
			{
				name: "registration ignores trailing slash",
				reg: RouteRegistry{
					"/some/route/":    Get(doNotWantHandler),
					"/another/route/": Get(doNotWantHandler),
					"/want/":          Get(wantHandler),
				},
				path: "/want",
			},
			{
				name: "lookup ignores trailing slash",
				path: "/want/",
			},
			{
				name:       "simple global",
				gNamespace: "/api",
				path:       "/api/want",
			},
			{
				name:       "multi tiered global",
				gNamespace: "/api/foo",
				path:       "/api/foo/want",
			},
			{
				name:       "ensures leading slash on global namespace",
				gNamespace: "api",
				path:       "/api/want",
			},
			{
				name:       "ignores trailing slash on global namespace",
				gNamespace: "/api/",
				path:       "/api/want",
			},
			{
				name:       "ignores multiple trailing slashes on global namespace",
				gNamespace: "/api//",
				path:       "/api/want",
			},
			{
				name:       "ensures leading slash on paths with global namespace",
				gNamespace: "/api",
				reg: RouteRegistry{
					"some/route":    Get(doNotWantHandler),
					"another/route": Get(doNotWantHandler),
					"want":          Get(wantHandler),
				},
				path: "/api/want",
			},
			{
				name:       "simple local",
				lNamespace: "/api",
				path:       "/api/want",
			},
			{
				name:       "multi tiered local namespace",
				lNamespace: "/api/foo",
				path:       "/api/foo/want",
			},
			{
				name:       "global and local",
				gNamespace: "/api",
				lNamespace: "/foo",
				path:       "/api/foo/want",
			},
			{
				name:       "local registration ensures leading slash",
				lNamespace: "api",
				path:       "/api/want",
			},
			{
				name:       "local registration ensures leading slashes on paths",
				lNamespace: "/api",
				reg: RouteRegistry{
					"some/route":    Get(doNotWantHandler),
					"another/route": Get(doNotWantHandler),
					"want":          Get(wantHandler),
				},
				path: "/api/want",
			},
			{
				name:       "local registration ignores trailing slash",
				lNamespace: "/api/",
				path:       "/api/want",
			},
			{
				name:       "local registration ignores multiple trailing slashes",
				lNamespace: "/api//",
				path:       "/api/want",
			},
			{
				name: "dynamic lookup includes path params",
				reg: RouteRegistry{
					"/:funky/:dynamic/:route": Get(wantHandler),
				},
				path:           "/awesome/dynamic/router-magic",
				wantPathParams: pathParameters{"funky": "awesome", "dynamic": "dynamic", "route": "router-magic"},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				router := newRouter()
				reg := tc.reg
				if len(reg) == 0 {
					reg = RouteRegistry{
						"/some/route":    Get(doNotWantHandler),
						"/another/route": Get(doNotWantHandler),
						"/want":          Get(wantHandler),
					}
				}
				if tc.lNamespace == "" {
					router.RegisterRoutes(reg)
				} else {
					router.RegisterRoutesUnderNamespace(tc.lNamespace, reg)
				}
				router.GlobalNamespace(tc.gNamespace)
				req := requestWithUrlAndMethod(tc.path, GET)

				got, found := router.Route(req)
				if !found {
					t.Error("expected route to be found")
				}
				err := got.handle(&ResponseWriter{}, req)
				if err != nil {
					t.Errorf("did not expect handler to error: %s", err.Error())
				}
				params := req.uri.pathParams
				if len(params) != len(tc.wantPathParams) {
					t.Errorf(`Route() returned %d length params, wanted %d`, len(params), len(tc.wantPathParams))
				}
				for k, v := range tc.wantPathParams {
					if params[k] != v {
						t.Errorf(`pathParams[%#q] = %s, wanted %s`, k, params[k], v)
					}
				}
			})
		}
	})

	t.Run("not found", func(t *testing.T) {
		tests := []RouteTest{
			{
				name: "empty",
				reg:  RouteRegistry{},
				path: "/want",
			},
			{
				name: "repeated slashes",
				path: "/some//route",
			},
			{
				name:       "valid route in registry, but global namespace",
				gNamespace: "/api",
				path:       "/want",
			},
			{
				name:       "valid route in registry, but local namespace",
				lNamespace: "/api",
				path:       "/want",
			},
			{
				name:       "valid route in registry, but global and local namespace, just route",
				gNamespace: "/api",
				lNamespace: "/foo",
				path:       "/want",
			},
			{
				name:       "valid route in registry, but global and local namespace, global + route",
				gNamespace: "/api",
				lNamespace: "/foo",
				path:       "/api/want",
			},
			{
				name:       "valid route in registry, but global and local namespace, local + route",
				gNamespace: "/api",
				lNamespace: "/foo",
				path:       "/foo/want",
			},
			{
				name:       "valid route in registry, but global and local namespace, local + global + route",
				gNamespace: "/api",
				lNamespace: "/foo",
				path:       "/foo/api/want",
			},
			{
				name:      "static backtracking",
				staticDir: "static",
				path:      "static/../main.go",
			},
			{
				name:       "global namespace, static backtracking",
				gNamespace: "/api",
				staticDir:  "static",
				path:       "/api/static/../main.go",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				router := newRouter()
				req := requestWithUrlAndMethod(tc.path, GET)
				if tc.lNamespace == "" {
					router.RegisterRoutes(tc.reg)
				} else {
					router.RegisterRoutesUnderNamespace(tc.lNamespace, tc.reg)
				}
				router.GlobalNamespace(tc.gNamespace)
				router.NewStaticDir(tc.staticDir)

				_, found := router.Route(req)
				if found {
					t.Errorf("did not expect to find a route for [url=%s, method=%s]", req.Path(), req.Method().name)
				}
			})
		}
	})
}

func TestStaticRoutingFound(t *testing.T) {
	router := newRouter()
	router.NewStaticDir("static")
	req := requestWithUrlAndMethod("/static", GET)

	// We don't want to actually load the file from disk in the test, so we
	// only assert on the presence of the routing. This works for these tests
	// since we have no other handlers registered, meaning that if we found
	// this one it must be the static loader.
	_, found := router.Route(req)
	if !found {
		t.Error("expected to find static router")
	}
}

func TestStaticRoutingGlobalNamespaceFound(t *testing.T) {
	router := newRouter()
	router.GlobalNamespace("/api")
	router.NewStaticDir("static")
	req := requestWithUrlAndMethod("/api/static", GET)

	// We don't want to actually load the file from disk in the test, so we
	// only assert on the presence of the routing. This works for these tests
	// since we have no other handlers registered, meaning that if we found
	// this one it must be the static loader.
	_, found := router.Route(req)
	if !found {
		t.Error("expected to find static router")
	}
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

			router.NewStaticDir(tc.in)
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
			router.NewStaticDir(tc.in)
			if router.staticDir != tc.want {
				t.Errorf(`router.static = %q, wanted %#q`, router.staticDir, tc.want)
			}
		})
	}
}

func TestRewrite(t *testing.T) {
	tests := []struct {
		name    string
		base    map[string]string
		in      string
		want    string
		rewrite bool
	}{
		{
			"empty",
			map[string]string{},
			"/foo/bar",
			"/foo/bar",
			false,
		},
		{
			"1 element no match",
			map[string]string{"/foo": "/bar"},
			"/foo/bar",
			"/foo/bar",
			false,
		},
		{
			"1 element match",
			map[string]string{"/foo/bar": "/baz"},
			"/foo/bar",
			"/baz",
			true,
		},
	}
	router := newRouter()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router.rewrites = tc.base

			actual, rewrite := router.Rewrite(tc.in)
			if rewrite != tc.rewrite {
				t.Errorf(`Rewrite(%#q) didRewrite? = %t, wanted %t`, tc.in, rewrite, tc.rewrite)
			}
			if actual != tc.want {
				t.Errorf(`Rewrite(%#q) rewritten = %s, wanted %s`, tc.in, actual, tc.want)
			}
		})
	}
}

func TestNewRewrite(t *testing.T) {
	t.Run("panics", func(t *testing.T) {
		tests := []struct {
			name   string
			raw    string
			before map[string]string
		}{
			{
				"only one entry",
				"/foo",
				map[string]string{},
			},
			{
				"missing leading slash",
				"foo /bar",
				map[string]string{},
			},
			{
				"too many entries",
				"/foo /bar /baz",
				map[string]string{},
			},
			{
				"conflicting duplication",
				"/foo /bar",
				map[string]string{"/foo": "/baz"},
			},
			{
				"interior comment",
				"/foo # Comment /bar",
				map[string]string{},
			},
			{
				"comment in key",
				"/foo#Comment /bar",
				map[string]string{},
			},
			{
				"empty path component on key",
				"// /bar",
				map[string]string{},
			},
			{
				"empty path component on value",
				"/foo/bar //",
				map[string]string{},
			},
			{
				"trailing slash on key",
				"/foo/ /bar",
				map[string]string{},
			},
			{
				"trailing slash on value",
				"/foo /bar/",
				map[string]string{},
			},
			{
				"trailing slash on value with whitespaced comment",
				"/foo /bar/ # Comment",
				map[string]string{},
			},
			{
				"trailing slash on value with non-whitespaced comment",
				"/foo /bar/# Comment",
				map[string]string{},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				router := newRouter()
				router.rewrites = tc.before

				defer func() {
					if r := recover(); r == nil {
						t.Error("expected panic, found none")
					}
				}()

				router.NewRewrite(tc.raw)
			})
		}
	})

	t.Run("happy", func(t *testing.T) {
		tests := []struct {
			name string
			raw  string
			want map[string]string
		}{
			{
				"empty line",
				"",
				map[string]string{},
			},
			{
				"comment line",
				"# This is a comment",
				map[string]string{},
			},
			{
				"comment with leading whitespace",
				"      # Comment",
				map[string]string{},
			},
			{
				"simple rewrite",
				"/foo /bar",
				map[string]string{"/foo": "/bar"},
			},
			{
				"skips equivalences",
				"/foo /foo",
				map[string]string{},
			},
			{
				"comment after value",
				"/foo/bar /baz # The comment",
				map[string]string{"/foo/bar": "/baz"},
			},
			{
				"comment immediately after value (no whitespace)",
				"/foo/bar /baz# The comment",
				map[string]string{"/foo/bar": "/baz"},
			},
			{
				"root in key",
				"/ /bar",
				map[string]string{"/": "/bar"},
			},
			{
				"root in value",
				"/foo /",
				map[string]string{"/foo": "/"},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				router := newRouter()

				router.NewRewrite(tc.raw)

				if len(router.rewrites) != len(tc.want) {
					t.Errorf(`len(rewrites) = %d, wanted %d`, len(router.rewrites), len(tc.want))
				}
				for k, v := range tc.want {
					actual, exists := router.rewrites[k]
					if !exists {
						t.Errorf("rewrites[%#q] not found, expected to find", k)
					}
					if actual != v {
						t.Errorf("rewrites[%#q] = %#q, wanted %#q", k, actual, v)
					}
				}
			})
		}
	})
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
	return &Request{uri: uri{edgePath: url}, mthd: method}
}
