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
			{
				"dynamic with prefix, search without prefix",
				map[string]int{"/foo/:bar|baz": 42},
				"/foo/bar",
			},
			{
				"dynamic with suffix, search without suffix",
				map[string]int{"/foo/:bar||baz": 42},
				"/foo/bar",
			},
			{
				"dynamic with prefix and suffix, search with prefix, without suffix",
				map[string]int{"/foo/:bar|baz|qux": 42},
				"/foo/baza",
			},
			{
				"dynamic with prefix and suffix, search without prefix, with suffix",
				map[string]int{"/foo/:bar|baz|qux": 42},
				"/foo/aqux",
			},
			{
				// The reason we don't want this to match is because the name
				// part is supposed to represent a capture group - it should
				// not be empty.
				"dynamic with prefix and suffix, search is exactly prefix + suffix",
				map[string]int{"/foo/:bar|baz|qux": 42},
				"/foo/bazqux",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				trie := newTrie((*trieValue[int]).PathParams)
				for k, v := range tc.in {
					trie.Insert(k, &v)
				}

				val, params, found := trie.Find(tc.search)
				if found {
					t.Fatalf(`Trie.Find(%#q), did not expect to find element`, tc.search)
				}
				if val != nil {
					t.Errorf(`Trie.Find(%#q) = %d, expected value to be nil`, tc.search, *val)
				}
				if params != nil && len(*params) != 0 {
					t.Errorf(`Trie.Find(%#q) returned %d length params, expected none`, tc.search, len(*params))
				}
			})
		}
	})

	t.Run("found", func(t *testing.T) {
		tests := []struct {
			name       string
			in         map[string]int
			search     string
			wantParams pathParameters
		}{
			{
				name:   "one element",
				in:     map[string]int{"/foo": 42},
				search: "/foo",
			},
			{
				name:   "multiple elements leaf",
				in:     map[string]int{"/foo": 13, "/foo/bar": 15, "/foo/bar/baz": 42},
				search: "/foo/bar/baz",
			},
			{
				name:   "multiple elements non-leaf",
				in:     map[string]int{"/foo": 13, "/foo/bar": 42, "/foo/bar/baz": 15},
				search: "/foo/bar",
			},
			{
				name:       "dynamic leaf",
				in:         map[string]int{"/foo/:bar": 42},
				search:     "/foo/some-variable",
				wantParams: pathParameters{"bar": "some-variable"},
			},
			{
				name:       "dynamic valid non-leaf",
				in:         map[string]int{"/foo/:bar": 42, "/foo/:bar/:baz": 13},
				search:     "/foo/some-variable",
				wantParams: pathParameters{"bar": "some-variable"},
			},
			{
				name:   "prioritises exact match",
				in:     map[string]int{"/foo/bar": 13, "/foo/:var": 100, "/foo/baz": 42},
				search: "/foo/baz",
			},
			{
				name:       "handles complex dynamic matches",
				in:         map[string]int{"/foo/:bar": 42},
				search:     "/foo/this-is-a-really!long-matcher-05A6C58E-0FE4-4108-93E7-8DEAD94282F8",
				wantParams: pathParameters{"bar": "this-is-a-really!long-matcher-05A6C58E-0FE4-4108-93E7-8DEAD94282F8"},
			},
			{
				name:       "prioritises more specific dynamic matches",
				in:         map[string]int{"/foo/:bar": 42, "/:foo/bar": 13},
				search:     "/foo/bar",
				wantParams: pathParameters{"bar": "bar"},
			},
			{
				name:       "prioritises dynamic nodes with more static components",
				in:         map[string]int{"/foo/:bar/:baz": 13, "/foo/:bar/baz": 42},
				search:     "/foo/bar/baz",
				wantParams: pathParameters{"bar": "bar"},
			},
			{
				name:       "prioritises same dynamic matches, more prefixes",
				in:         map[string]int{"/foo/:bar|baz": 42, "/foo/:bar": 13},
				search:     "/foo/baza",
				wantParams: pathParameters{"bar": "baza"},
			},
			{
				name:       "prioritises same dynamic matches, more suffixes",
				in:         map[string]int{"/foo/:bar||baz": 42, "/foo/:bar": 13},
				search:     "/foo/abaz",
				wantParams: pathParameters{"bar": "abaz"},
			},
			{
				name:       "prioritises same dynamic matches, 1 suffix + prefix over 1 prefix",
				in:         map[string]int{"/foo/:bar|baz|bar": 42, "/foo/:bar|baz": 13},
				search:     "/foo/bazabar",
				wantParams: pathParameters{"bar": "bazabar"},
			},
			{
				name:       "prioritises same dynamic matches, 1 suffix + prefix over 1 suffix",
				in:         map[string]int{"/foo/:bar|baz|bar": 42, "/foo/:bar||bar": 13},
				search:     "/foo/bazabar",
				wantParams: pathParameters{"bar": "bazabar"},
			},
			{
				name:       "prioritises less dynamic matches over more dynamic matches with 1 suffix + prefix",
				in:         map[string]int{"/foo/:bar/qux": 42, "/foo/:bar|baz|bar/:qux": 13},
				search:     "/foo/bazabar/qux",
				wantParams: pathParameters{"bar": "bazabar"},
			},
			{
				name:       "prioritises more specific dynamic matches (1 prefix) for same count, different position",
				in:         map[string]int{"/foo/:bar|baz/qux": 42, "/foo/baza/:bar": 13},
				search:     "/foo/baza/qux",
				wantParams: pathParameters{"bar": "baza"},
			},
			{
				name:       "dynamic match with prefix",
				in:         map[string]int{"/foo/:bar|baz": 42},
				search:     "/foo/baz_search",
				wantParams: pathParameters{"bar": "baz_search"},
			},
			{
				name:       "dynamic match with suffix",
				in:         map[string]int{"/foo/:bar||baz": 42},
				search:     "/foo/search_baz",
				wantParams: pathParameters{"bar": "search_baz"},
			},
			{
				name:       "dynamic match with prefix and suffix",
				in:         map[string]int{"/foo/:bar|baz|qux": 42},
				search:     "/foo/bazaqux",
				wantParams: pathParameters{"bar": "bazaqux"},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				trie := newTrie((*trieValue[int]).PathParams)
				for k, v := range tc.in {
					trie.Insert(k, &v)
				}

				actual, params, found := trie.Find(tc.search)
				if !found {
					t.Errorf("Trie.Find(%#q) expected to find element", tc.search)
				}
				if *actual != 42 {
					t.Errorf(`trie["%s"] = %d, wanted 42`, tc.search, *actual)
				}
				if len(*params) != len(tc.wantParams) {
					t.Errorf(`Trie.Find(%#q) returned %d length params, wanted %d`, tc.search, len(*params), len(tc.wantParams))
				}
				for k, v := range tc.wantParams {
					if (*params)[k] != v {
						t.Errorf(`pathParams[%#q] = %s, wanted %s`, k, (*params)[k], v)
					}
				}
			})
		}
	})
}

func TestTrieInsertion(t *testing.T) {
	t.Run("bad inputs", func(t *testing.T) {
		tests := []struct {
			name  string
			in    string
			start *slashTrie[int, pathParameters]
		}{
			{
				"duplicate names",
				"/:foo/bar/:foo",
				newTrie((*trieValue[int]).PathParams),
			},
			{
				"conflicting dynamic",
				"/:foo/bar/:bar",
				func() *slashTrie[int, pathParameters] {
					trie := newTrie((*trieValue[int]).PathParams)
					v := 17
					trie.Insert("/:foo/bar/:baz", &v)
					return trie
				}(),
			},
			{
				"dynamic too many separators",
				"/:foo|pre|suf|",
				newTrie((*trieValue[int]).PathParams),
			},
			{
				"dynamic too many sections",
				"/:foo|pre|suf|extra",
				newTrie((*trieValue[int]).PathParams),
			},
			{
				"dynamic match no name",
				"/:",
				newTrie((*trieValue[int]).PathParams),
			},
			{
				"dynamic match no name - prefix pipe",
				"/:|",
				newTrie((*trieValue[int]).PathParams),
			},
			{
				"dynamic match no name - prefix and suffix pipe",
				"/:||",
				newTrie((*trieValue[int]).PathParams),
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				trie := tc.start

				defer func() {
					if r := recover(); r == nil {
						t.Error("trie invalid insertion, expected panic")
					}
				}()

				val := 42
				trie.Insert(tc.in, &val)
			})
		}
	})
}
