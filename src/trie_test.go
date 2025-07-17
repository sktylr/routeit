package routeit

import "testing"

func TestTrieLookup(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		tests := []struct {
			name   string
			in     map[string]int
			search string
		}{
			{
				"empty",
				map[string]int{},
				"/foo",
			},
			{
				"multiple populated",
				map[string]int{"/foo": 13, "/foo/bar": 0, "/foo/baz": 9, "/foo/bar/qux": 17},
				"/foo/bar/baz",
			},
			{
				"present but non value",
				map[string]int{"/foo/bar/baz": 42, "/foo/baz": 19, "/foo/bar/qux": 13},
				"/foo/bar",
			},
			{
				"dynamic present but non value",
				map[string]int{"/foo/:bar/baz": 42},
				"/foo/bar",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				trie := newTrie[int]()
				for k, v := range tc.in {
					trie.Insert(k, &v)
				}

				val, params, found := trie.Find(tc.search)
				if found {
					t.Errorf(`Trie.Find(%#q), did not expect to find element`, tc.search)
				}
				if val != nil {
					t.Errorf(`Trie.Find(%#q) = %d, expected value to be nil`, tc.search, *val)
				}
				if len(params) != 0 {
					t.Errorf(`Trie.Find(%#q) returned %d length params, expected none`, tc.search, len(params))
				}
			})
		}
	})

	t.Run("found", func(t *testing.T) {
		tests := []struct {
			name       string
			in         map[string]int
			search     string
			want       int
			wantParams pathParameters
		}{
			{
				"one element",
				map[string]int{"/foo": 13},
				"/foo",
				13,
				pathParameters{},
			},
			{
				"multiple elements leaf",
				map[string]int{"/foo": 13, "/foo/bar": 42, "/foo/bar/baz": 15},
				"/foo/bar/baz",
				15,
				pathParameters{},
			},
			{
				"multiple elements non-leaf",
				map[string]int{"/foo": 13, "/foo/bar": 42, "/foo/bar/baz": 15},
				"/foo/bar",
				42,
				pathParameters{},
			},
			{
				"dynamic leaf",
				map[string]int{"/foo/:bar": 14},
				"/foo/some-variable",
				14,
				pathParameters{"bar": "some-variable"},
			},
			{
				"dynamic valid non-leaf",
				map[string]int{"/foo/:bar": 15, "/foo/:bar/:baz": 13},
				"/foo/some-variable",
				15,
				pathParameters{"bar": "some-variable"},
			},
			{
				"prioritises exact match",
				map[string]int{"/foo/bar": 13, "/foo/:var": 100, "/foo/baz": 42},
				"/foo/baz",
				42,
				pathParameters{},
			},
			{
				"handles complex dynamic matches",
				map[string]int{"/foo/:bar": 15},
				"/foo/this-is-a-really!long-matcher-05A6C58E-0FE4-4108-93E7-8DEAD94282F8",
				15,
				pathParameters{"bar": "this-is-a-really!long-matcher-05A6C58E-0FE4-4108-93E7-8DEAD94282F8"},
			},
			{
				"prioritises more specific dynamic matches",
				map[string]int{"/foo/:bar": 17, "/:foo/bar": 13},
				"/foo/bar",
				17,
				pathParameters{"bar": "bar"},
			},
			{
				"prioritises dynamic nodes with more static components",
				map[string]int{"/foo/:bar/:baz": 42, "/foo/:bar/baz": 13},
				"/foo/bar/baz",
				13,
				pathParameters{"bar": "bar"},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				trie := newTrie[int]()
				for k, v := range tc.in {
					trie.Insert(k, &v)
				}

				actual, params, found := trie.Find(tc.search)
				if !found {
					t.Errorf("Trie.Find(%#q) expected to find element", tc.search)
				}
				if *actual != tc.want {
					t.Errorf(`trie["%s"] = %d, wanted %d`, tc.search, *actual, tc.want)
				}
				if len(params) != len(tc.wantParams) {
					t.Errorf(`Trie.Find(%#q) returned %d length params, wanted %d`, tc.search, len(params), len(tc.wantParams))
				}
				for k, v := range tc.wantParams {
					if params[k] != v {
						t.Errorf(`pathParams[%#q] = %s, wanted %s`, k, params[k], v)
					}
				}
			})
		}
	})
}
