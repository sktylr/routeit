package routeit

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
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
			{
				name: "complex dynamic matches",
				reg: RouteRegistry{
					"/foo/:bar": Get(wantHandler),
				},
				path:           "/foo/this-is-a-really!long-matcher-05A6C58E-0FE4-4108-93E7-8DEAD94282F8",
				wantPathParams: pathParameters{"bar": "this-is-a-really!long-matcher-05A6C58E-0FE4-4108-93E7-8DEAD94282F8"},
			},

			{
				name: "prioritises same dynamic matches, more prefixes",
				reg: RouteRegistry{
					"/foo/:bar|baz": Get(wantHandler),
					"/foo/:bar":     Get(doNotWantHandler),
				},
				path:           "/foo/baza",
				wantPathParams: pathParameters{"bar": "baza"},
			},
			{
				name: "prioritises same dynamic matches, more suffixes",
				reg: RouteRegistry{
					"/foo/:bar||baz": Get(wantHandler),
					"/foo/:bar":      Get(doNotWantHandler),
				},
				path:           "/foo/abaz",
				wantPathParams: pathParameters{"bar": "abaz"},
			},
			{
				name: "prioritises same dynamic matches, 1 suffix + prefix over 1 prefix",
				reg: RouteRegistry{
					"/foo/:bar|baz|bar": Get(wantHandler),
					"/foo/:bar|baz":     Get(doNotWantHandler),
				},
				path:           "/foo/bazabar",
				wantPathParams: pathParameters{"bar": "bazabar"},
			},
			{
				name: "prioritises same dynamic matches, 1 suffix + prefix over 1 suffix",
				reg: RouteRegistry{
					"/foo/:bar|baz|bar": Get(wantHandler),
					"/foo/:bar||bar":    Get(doNotWantHandler),
				},
				path:           "/foo/bazabar",
				wantPathParams: pathParameters{"bar": "bazabar"},
			},
			{
				name: "prioritises less dynamic matches over more dynamic matches with 1 suffix + prefix",
				reg: RouteRegistry{
					"/foo/:bar/qux":          Get(wantHandler),
					"/foo/:bar|baz|bar/:qux": Get(doNotWantHandler),
				},
				path:           "/foo/bazabar/qux",
				wantPathParams: pathParameters{"bar": "bazabar"},
			},
			{
				name: "prioritises more specific dynamic matches (1 prefix) for same count, different position",
				reg: RouteRegistry{
					"/foo/:bar|baz/qux": Get(wantHandler),
					"/foo/baza/:bar":    Get(doNotWantHandler),
				},
				path:           "/foo/baza/qux",
				wantPathParams: pathParameters{"bar": "baza"},
			},
			{
				name: "dynamic match with prefix",
				reg: RouteRegistry{
					"/foo/:bar|baz": Get(wantHandler),
				},
				path:           "/foo/baz_search",
				wantPathParams: pathParameters{"bar": "baz_search"},
			},
			{
				name: "dynamic match with suffix",
				reg: RouteRegistry{
					"/foo/:bar||baz": Get(wantHandler),
				},
				path:           "/foo/search_baz",
				wantPathParams: pathParameters{"bar": "search_baz"},
			},
			{
				name: "dynamic match with prefix and suffix",
				reg: RouteRegistry{
					"/foo/:bar|baz|qux": Get(wantHandler),
				},
				path:           "/foo/bazaqux",
				wantPathParams: pathParameters{"bar": "bazaqux"},
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
					t.Fatal("expected route to be found")
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

	t.Run("static found", func(t *testing.T) {
		tests := []RouteTest{
			{
				name: "simple",
				path: "/static",
			},
			{
				name:       "with global namespace",
				gNamespace: "/api",
				path:       "/api/static",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				router := newRouter()
				router.NewStaticDir("static")
				router.GlobalNamespace(tc.gNamespace)
				req := requestWithUrlAndMethod(tc.path, GET)

				// We don't want to actually load the file from disk in the test, so we
				// only assert on the presence of the routing. This works for these tests
				// since we have no other handlers registered, meaning that if we found
				// this one it must be the static loader.
				_, found := router.Route(req)
				if !found {
					t.Error("expected to find static router")
				}
			})
		}
	})
}

func TestNewStaticDir(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "root directory", in: "~/foo/bar",
		},
		{
			name: "containing variables", in: "$HOME/bar",
		},
		{
			name: "backtracking outside", in: "static/../..",
		},
		{
			name: "project root", in: "static/..",
		},
		{
			// The leading slash should be understood to mean ./ since we don't
			// allow access outside of the project root.
			name: "leading slash", in: "/static", want: "static",
		},
		{
			name: "useless backtrack", in: "static/../static/../static", want: "static",
		},
		{
			name: "cyclic", in: "static/foo/../../static/../static/foo/../foo", want: "static/foo",
		},
		{
			// The `~` character is only expanded to the system root within the
			// shell (it is a shorthand), but it is a perfectly legal, albeit
			// confusing, directory name. This should not be expanded to the
			// computer root.
			name: "containing root shorthand", in: "static/~/foo", want: "static/~/foo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := newRouter()
			wantPanic := tc.want == ""

			defer func() {
				if r := recover(); r == nil && wantPanic {
					t.Errorf("router invalid static dir, expected panic but got none - static = %+v", router.staticDir)
				}
			}()

			router.NewStaticDir(tc.in)

			want := strings.Split(tc.want, "/")
			if !wantPanic && !reflect.DeepEqual(router.staticDir, want) {
				t.Errorf(`router.static = %+v, wanted %#q`, router.staticDir, tc.want)
			}
		})
	}
}

func TestRewritePath(t *testing.T) {
	tests := []struct {
		name            string
		base            map[string]string
		in              string
		wantRewritten   []string
		wantQueryParams queryParameters
		rewrite         bool
	}{
		{
			name:          "empty",
			in:            "/foo/bar",
			wantRewritten: []string{"foo", "bar"},
			rewrite:       false,
		},
		{
			name:          "1 element no match",
			base:          map[string]string{"/foo": "/bar"},
			in:            "/foo/bar",
			wantRewritten: []string{"foo", "bar"},
			rewrite:       false,
		},
		{
			name:          "1 element match",
			base:          map[string]string{"/foo/bar": "/baz"},
			in:            "/foo/bar",
			wantRewritten: []string{"baz"},
			rewrite:       true,
		},
		{
			name:          "1 dynamic",
			base:          map[string]string{"/foo/${bar}": "/baz/${bar}"},
			in:            "/foo/qux",
			wantRewritten: []string{"baz", "qux"},
			rewrite:       true,
		},
		{
			name:          "2 dynamics (same count), prioritises later dynamics",
			base:          map[string]string{"/foo/${bar}": "/baz/${bar}", "/${foo}/bar": "/${foo}/qux"},
			in:            "/foo/bar",
			wantRewritten: []string{"baz", "bar"},
			rewrite:       true,
		},
		{
			name:          "2 dynamics, different count, prioritises shorter count",
			base:          map[string]string{"/foo/${bar}/baz": "/pick/me", "/${foo}/bar/${baz}": "/not/me"},
			in:            "/foo/bar/baz",
			wantRewritten: []string{"pick", "me"},
			rewrite:       true,
		},
		{
			name:          "prioritises static rewrites",
			base:          map[string]string{"/foo/bar": "/baz", "/foo/${bar}": "/${bar}"},
			in:            "/foo/bar",
			wantRewritten: []string{"baz"},
			rewrite:       true,
		},
		{
			name:          "dynamic does not match (too short)",
			base:          map[string]string{"/foo/${bar}": "/baz/${bar}"},
			in:            "/foo",
			wantRewritten: []string{"foo"},
			rewrite:       false,
		},
		{
			name:          "dynamic does not match (too long)",
			base:          map[string]string{"/foo/${bar}": "/baz/${bar}"},
			in:            "/foo/bar/baz",
			wantRewritten: []string{"foo", "bar", "baz"},
			rewrite:       false,
		},
		{
			name:            "dynamic path to query param",
			base:            map[string]string{"/foo/${bar}": "/baz?id=${bar}"},
			in:              "/foo/123",
			wantRewritten:   []string{"baz"},
			wantQueryParams: queryParameters{"id": {"123"}},
			rewrite:         true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := newRouter()
			for k, v := range tc.base {
				router.NewRewrite(fmt.Sprintf("%s %s", k, v))
			}
			uri, err := parseUri(tc.in)
			if err != nil {
				t.Fatalf("error while parsing input uri: %v", err)
			}

			err = router.RewriteUri(uri)

			if err != nil {
				t.Fatalf("unexpected error during rewrite: %v", err)
			}
			actual := uri.rewrittenPath
			if !uri.rewritten {
				actual = uri.edgePath
			}
			if uri.rewritten != tc.rewrite {
				t.Errorf("RewritePath(%q) didRewrite? = %t, wanted %t", tc.in, uri.rewritten, tc.rewrite)
			}
			if !reflect.DeepEqual(actual, tc.wantRewritten) {
				t.Errorf("RewritePath(%q) rewritten = %+v, wanted %+v", tc.in, actual, tc.wantRewritten)
			}
			if tc.wantQueryParams != nil && !reflect.DeepEqual(uri.queryParams.q, tc.wantQueryParams) {
				t.Errorf("RewritePath(%q) query params = %#v, wanted %#v", tc.in, uri.queryParams.q, tc.wantQueryParams)
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
				name: "only one entry",
				raw:  "/foo",
			},
			{
				name: "missing leading slash",
				raw:  "foo /bar",
			},
			{
				name: "too many entries",
				raw:  "/foo /bar /baz",
			},
			{
				name:   "conflicting duplication",
				raw:    "/foo /bar",
				before: map[string]string{"/foo": "/baz"},
			},
			{
				name:   "conflicting dynamic duplication to static",
				raw:    "/${foo} /bar",
				before: map[string]string{"/${foo}": "/baz"},
			},
			{
				name:   "conflicting dynamic duplication to dynamic",
				raw:    "/${foo} /bar/${foo}",
				before: map[string]string{"/${foo}": "/bar/qux/${foo}"},
			},
			{
				name:   "conflicting dynamic duplication to complex dynamic",
				raw:    "/${foo} /bar/${foo}.png",
				before: map[string]string{"/${foo}": "/bar/${foo}"},
			},
			{
				name: "interior comment",
				raw:  "/foo # Comment /bar",
			},
			{
				name: "comment in key",
				raw:  "/foo#Comment /bar",
			},
			{
				name: "empty path component on key",
				raw:  "// /bar",
			},
			{
				name: "empty path component on value",
				raw:  "/foo/bar //",
			},
			{
				name: "trailing slash on key",
				raw:  "/foo/ /bar",
			},
			{
				name: "trailing slash on value",
				raw:  "/foo /bar/",
			},
			{
				name: "trailing slash on value with whitespaced comment",
				raw:  "/foo /bar/ # Comment",
			},
			{
				name: "trailing slash on value with non-whitespaced comment",
				raw:  "/foo /bar/# Comment",
			},
			{
				name: "dynamic key reuses variable names",
				raw:  "/foo/${bar}/baz/${bar} /qux/${bar}/quux/${bar}",
			},
			{
				name: "two variables same path component - key",
				raw:  "/foo/${bar}${baz} /qux",
			},
			{
				name: "trailing brace (key)",
				raw:  "/foo/${bar}} /baz",
			},
			{
				name: "two braced strings same path component - key",
				raw:  "/foo/${bar}{baz} /foo",
			},
			{
				name: "too many pipes",
				raw:  "/foo/${bar|||} /foo",
			},
			{
				name: "non-alphanumeric prefix",
				raw:  "/foo/${bar|\n} /foo",
			},
			{
				name: "non-alphanumeric suffix",
				raw:  "/foo/${bar||\n} /foo",
			},
			{
				name: "pipe in dynamic value",
				raw:  "/foo/${bar} /baz/${bar|prefix}",
			},
			{
				name:   "conflicting dynamic duplication with prefix",
				raw:    "/${foo|pref} /bar/${foo}.png",
				before: map[string]string{"/${foo|pref}": "/baz/${foo}"},
			},
			{
				name:   "conflicting dynamic duplication with suffix",
				raw:    "/${foo||suf} /bar/${foo}.png",
				before: map[string]string{"/${bar||suf}": "/baz/${bar}"},
			},
			{
				name:   "conflicting dynamic duplication with prefix and suffix",
				raw:    "/${foo|pref|suf} /bar/${foo}.png",
				before: map[string]string{"/${bar|pref|suf}": "/baz/${bar}"},
			},
			{
				name: "query parameter not on last path component",
				raw:  "/${foo} /bar?foo=${foo}/baz",
			},
			{
				name: "multiple query strings on last path component",
				raw:  "/${foo}/${bar} /baz?foo=${foo}?bar=${bar}",
			},
			{
				name: "query parameter in key",
				raw:  "/${foo}?bar=${baz} /bar/${foo}/${baz}",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				router := newRouter()
				for k, v := range tc.before {
					router.NewRewrite(fmt.Sprintf("%s %s", k, v))
				}

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
			name     string
			existing []string
			raw      string
			want     map[string][]string
		}{
			{
				name: "empty line",
				raw:  "",
			},
			{
				name: "comment line",
				raw:  "# This is a comment",
			},
			{
				name: "comment with leading whitespace",
				raw:  "      # Comment",
			},
			{
				name: "simple rewrite",
				raw:  "/foo /bar",
				want: map[string][]string{"foo": {"bar"}},
			},
			{
				name: "skips equivalences",
				raw:  "/foo /foo",
			},
			{
				name: "comment after value",
				raw:  "/foo/bar /baz # The comment",
				want: map[string][]string{"foo/bar": {"baz"}},
			},
			{
				name: "comment immediately after value (no whitespace)",
				raw:  "/foo/bar /baz# The comment",
				want: map[string][]string{"foo/bar": {"baz"}},
			},
			{
				name: "root in key",
				raw:  "/ /bar",
				want: map[string][]string{"": {"bar"}},
			},
			{
				name: "root in value",
				raw:  "/foo /",
				want: map[string][]string{"foo": {""}},
			},
			{
				name: "dynamic rewrite (key and value)",
				raw:  "/foo/${bar} /baz/${bar}",
				want: map[string][]string{"foo/bar": {"baz", "bar"}, "foo/foo": {"baz", "foo"}, "foo/123": {"baz", "123"}},
			},
			{
				name: "dynamic rewrites (key only)",
				raw:  "/foo/${bar} /baz/bar",
				want: map[string][]string{"foo/bar": {"baz", "bar"}, "foo/qux": {"baz", "bar"}, "foo/ABCD": {"baz", "bar"}},
			},
			{
				name: "dynamic rewrites (value only)",
				raw:  "/foo/bar /baz/${bar}",
				want: map[string][]string{"foo/bar": {"baz", "${bar}"}},
			},
			{
				name: "dynamic rewrites (value only) with repeated variable in value",
				raw:  "/foo/bar /baz/${bar}/${bar}",
				want: map[string][]string{"foo/bar": {"baz", "${bar}", "${bar}"}},
			},
			{
				name: "poorly formed dynamic key acts as static",
				raw:  "/foo/${bar /baz/${bar}",
				want: map[string][]string{"foo/${bar": {"baz", "${bar}"}},
			},
			{
				name:     "non conflicting duplicates allowed",
				existing: []string{"/${page} /assets/${page}.html"},
				// This is not conflicting since it contains no dynamic values
				// in the key, whereas the existing entry does. In lookup,
				// this entry will be prioritised if the input is exactly
				// /favicon.ico
				raw:  "/favicon.ico /assets/images/foo.png",
				want: map[string][]string{"favicon.ico": {"assets", "images", "foo.png"}, "contact": {"assets", "contact.html"}},
			},
			{
				name: "dynamic path with prefix",
				raw:  "/foo/${bar|prefix} /baz/${bar}",
				want: map[string][]string{"foo/prefixed": {"baz", "prefixed"}},
			},
			{
				name: "dynamic path with suffix",
				raw:  "/foo/${bar||suffix} /baz/${bar}",
				want: map[string][]string{"foo/my-suffix": {"baz", "my-suffix"}},
			},
			{
				name: "dynamic path with prefix and suffix",
				raw:  "/foo/${bar|prefix|suffix} /baz/${bar}",
				want: map[string][]string{"foo/prefix-suffix": {"baz", "prefix-suffix"}},
			},
			{
				name:     "colliding but 1 with prefix",
				existing: []string{"/${foo} /bar"},
				raw:      "/${foo|pref} /bar/baz",
				want:     map[string][]string{"pref-foo": {"bar", "baz"}, "pre": {"bar"}},
			},
			{
				name:     "colliding but 1 with suffix",
				existing: []string{"/${foo} /bar"},
				raw:      "/${foo||suf} /bar/baz",
				want:     map[string][]string{"foo-suf": {"bar", "baz"}, "suf": {"bar"}},
			},
			{
				// This is technically fine but a discouraged practice as it
				// can lead to ambiguous collisions.
				name:     "colliding but 1 with prefix, other with suffix",
				existing: []string{"/${foo|pref} /bar"},
				raw:      "/${foo||suf} /bar/baz",
				want:     map[string][]string{"foo-suf": {"bar", "baz"}, "pref-foo": {"bar"}},
			},
			{
				name:     "colliding but 1 with prefix, other with both",
				existing: []string{"/${foo|pref} /bar"},
				raw:      "/${foo|pref|suf} /bar/baz",
				want:     map[string][]string{"pref-suf": {"bar", "baz"}, "pref-foo": {"bar"}},
			},
			{
				name:     "colliding but 1 with suffix, other with both",
				existing: []string{"/${foo||suf} /bar"},
				raw:      "/${foo|pref|suf} /bar/baz",
				want:     map[string][]string{"pref-suf": {"bar", "baz"}, "pre-suf": {"bar"}},
			},
			{
				name: "query parameter on single path component value",
				raw:  "/foo/${bar} /baz?foo=${bar}",
				want: map[string][]string{"foo/baz": {"baz?foo=baz"}},
			},
			{
				name: "multiple query parameters on single path component value",
				raw:  "/foo/${bar} /baz?foo=${bar}&qux=foo-${bar}",
				want: map[string][]string{"foo/baz": {"baz?foo=baz&qux=foo-baz"}},
			},
			{
				name: "query parameter on multi path component value",
				raw:  "/foo/${bar} /bar/baz?foo=${bar}",
				want: map[string][]string{"foo/baz": {"bar", "baz?foo=baz"}},
			},
			{
				name: "multiple query parameters on multi path component value",
				raw:  "/foo/${bar} /bar/baz?foo=${bar}&qux=foo-${bar}",
				want: map[string][]string{"foo/baz": {"bar", "baz?foo=baz&qux=foo-baz"}},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				router := newRouter()
				for _, e := range tc.existing {
					router.NewRewrite(e)
				}

				router.NewRewrite(tc.raw)

				for k, v := range tc.want {
					key := strings.Split(k, "/")
					rewritten, exists := router.rewrites.Find(key)
					if !exists {
						t.Fatalf("rewrites[%#q] not found, expected to find", k)
					}
					if !reflect.DeepEqual(*rewritten, v) {
						t.Errorf("rewrites[%#q] = %+v, wanted %+v", k, *rewritten, v)
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
	uri, err := parseUri(url)
	if err != nil {
		panic(fmt.Errorf("failed to parse uri [%s] for test: %v", url, err))
	}
	return &Request{uri: *uri, mthd: method}
}
