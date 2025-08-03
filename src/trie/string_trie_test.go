package trie

import (
	"regexp"
	"testing"
)

type extracted struct {
	val   *int
	parts []string
	re    *regexp.Regexp
}

type extractor struct{}

func (e *extractor) NewFromStatic(val *int) *extracted {
	return &extracted{val: val}
}

func (e *extractor) NewFromDynamic(val *int, parts []string, re *regexp.Regexp) *extracted {
	return &extracted{val: val, parts: parts, re: re}
}

func TestTrieLookup(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		tests := []struct {
			name   string
			in     map[string]int
			search []string
		}{
			{
				"empty",
				map[string]int{},
				[]string{"foo"},
			},
			{
				"multiple populated",
				map[string]int{"/foo": 13, "/foo/bar": 0, "/foo/baz": 9, "/foo/bar/qux": 17},
				[]string{"foo", "bar", "baz"},
			},
			{
				"present but non value",
				map[string]int{"/foo/bar/baz": 42, "/foo/baz": 19, "/foo/bar/qux": 13},
				[]string{"foo", "bar"},
			},
			{
				"dynamic present but non value",
				map[string]int{"/foo/:bar/baz": 42},
				[]string{"foo", "bar"},
			},
			{
				"dynamic with prefix, search without prefix",
				map[string]int{"/foo/:bar|baz": 42},
				[]string{"foo", "bar"},
			},
			{
				"dynamic with suffix, search without suffix",
				map[string]int{"/foo/:bar||baz": 42},
				[]string{"foo", "bar"},
			},
			{
				"dynamic with prefix and suffix, search with prefix, without suffix",
				map[string]int{"/foo/:bar|baz|qux": 42},
				[]string{"foo", "baza"},
			},
			{
				"dynamic with prefix and suffix, search without prefix, with suffix",
				map[string]int{"/foo/:bar|baz|qux": 42},
				[]string{"foo", "aqux"},
			},
			{
				// The reason we don't want this to match is because the name
				// part is supposed to represent a capture group - it should
				// not be empty.
				"dynamic with prefix and suffix, search is exactly prefix + suffix",
				map[string]int{"/foo/:bar|baz|qux": 42},
				[]string{"foo", "bazqux"},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				trie := NewStringTrie('/', &extractor{})
				for k, v := range tc.in {
					trie.Insert(k, &v)
				}

				val, found := trie.Find(tc.search)
				if found {
					t.Fatalf(`Trie.Find(%+v), did not expect to find element`, tc.search)
				}
				if val != nil {
					t.Errorf(`Trie.Find(%#q) = %+v, expected value to be nil`, tc.search, *val)
				}
			})
		}
	})

	t.Run("found", func(t *testing.T) {
		tests := []struct {
			name        string
			in          map[string]int
			search      []string
			wantDynamic bool
		}{
			{
				name:   "one element",
				in:     map[string]int{"/foo": 42},
				search: []string{"foo"},
			},
			{
				name:   "multiple elements leaf",
				in:     map[string]int{"/foo": 13, "/foo/bar": 15, "/foo/bar/baz": 42},
				search: []string{"foo", "bar", "baz"},
			},
			{
				name:   "multiple elements non-leaf",
				in:     map[string]int{"/foo": 13, "/foo/bar": 42, "/foo/bar/baz": 15},
				search: []string{"foo", "bar"},
			},
			{
				name:        "dynamic leaf",
				in:          map[string]int{"/foo/:bar": 42},
				search:      []string{"foo", "some-variable"},
				wantDynamic: true,
			},
			{
				name:        "dynamic valid non-leaf",
				in:          map[string]int{"/foo/:bar": 42, "/foo/:bar/:baz": 13},
				search:      []string{"foo", "some-variable"},
				wantDynamic: true,
			},
			{
				name:   "prioritises exact match",
				in:     map[string]int{"/foo/bar": 13, "/foo/:var": 100, "/foo/baz": 42},
				search: []string{"foo", "baz"},
			},
			{
				name:        "handles complex dynamic matches",
				in:          map[string]int{"/foo/:bar": 42},
				search:      []string{"foo", "this-is-a-really!long-matcher-05A6C58E-0FE4-4108-93E7-8DEAD94282F8"},
				wantDynamic: true,
			},
			{
				name:        "prioritises more specific dynamic matches",
				in:          map[string]int{"/foo/:bar": 42, "/:foo/bar": 13},
				search:      []string{"foo", "bar"},
				wantDynamic: true,
			},
			{
				name:        "prioritises dynamic nodes with more static components",
				in:          map[string]int{"/foo/:bar/:baz": 13, "/foo/:bar/baz": 42},
				search:      []string{"foo", "bar", "baz"},
				wantDynamic: true,
			},
			{
				name:        "prioritises same dynamic matches, more prefixes",
				in:          map[string]int{"/foo/:bar|baz": 42, "/foo/:bar": 13},
				search:      []string{"foo", "baza"},
				wantDynamic: true,
			},
			{
				name:        "prioritises same dynamic matches, more suffixes",
				in:          map[string]int{"/foo/:bar||baz": 42, "/foo/:bar": 13},
				search:      []string{"foo", "abaz"},
				wantDynamic: true,
			},
			{
				name:        "prioritises same dynamic matches, 1 suffix + prefix over 1 prefix",
				in:          map[string]int{"/foo/:bar|baz|bar": 42, "/foo/:bar|baz": 13},
				search:      []string{"foo", "bazabar"},
				wantDynamic: true,
			},
			{
				name:        "prioritises same dynamic matches, 1 suffix + prefix over 1 suffix",
				in:          map[string]int{"/foo/:bar|baz|bar": 42, "/foo/:bar||bar": 13},
				search:      []string{"foo", "bazabar"},
				wantDynamic: true,
			},
			{
				name:        "prioritises less dynamic matches over more dynamic matches with 1 suffix + prefix",
				in:          map[string]int{"/foo/:bar/qux": 42, "/foo/:bar|baz|bar/:qux": 13},
				search:      []string{"foo", "bazabar", "qux"},
				wantDynamic: true,
			},
			{
				name:        "prioritises more specific dynamic matches (1 prefix) for same count, different position",
				in:          map[string]int{"/foo/:bar|baz/qux": 42, "/foo/baza/:bar": 13},
				search:      []string{"foo", "baza", "qux"},
				wantDynamic: true,
			},
			{
				name:        "dynamic match with prefix",
				in:          map[string]int{"/foo/:bar|baz": 42},
				search:      []string{"foo", "baz_search"},
				wantDynamic: true,
			},
			{
				name:        "dynamic match with suffix",
				in:          map[string]int{"/foo/:bar||baz": 42},
				search:      []string{"foo", "search_baz"},
				wantDynamic: true,
			},
			{
				name:        "dynamic match with prefix and suffix",
				in:          map[string]int{"/foo/:bar|baz|qux": 42},
				search:      []string{"foo", "bazaqux"},
				wantDynamic: true,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				trie := NewStringTrie('/', &extractor{})
				for k, v := range tc.in {
					trie.Insert(k, &v)
				}

				actual, found := trie.Find(tc.search)
				if !found {
					t.Errorf("Trie.Find(%+v) expected to find element", tc.search)
				}
				if *actual.val != 42 {
					t.Errorf(`trie["%s"] = %+v, wanted 42`, tc.search, *actual)
				}
				if tc.wantDynamic != (actual.re != nil && len(actual.parts) > 0) {
					t.Errorf("wanted dynamic, got static %+v", actual)
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
			start *StringTrie[int, extracted]
		}{
			{
				"duplicate names",
				"/:foo/bar/:foo",
				NewStringTrie('/', &extractor{}),
			},
			{
				"conflicting dynamic",
				"/:foo/bar/:bar",
				func() *StringTrie[int, extracted] {
					trie := NewStringTrie('/', &extractor{})
					v := 17
					trie.Insert("/:foo/bar/:baz", &v)
					return trie
				}(),
			},
			{
				"dynamic too many separators",
				"/:foo|pre|suf|",
				NewStringTrie('/', &extractor{}),
			},
			{
				"dynamic too many sections",
				"/:foo|pre|suf|extra",
				NewStringTrie('/', &extractor{}),
			},
			{
				"dynamic match no name",
				"/:",
				NewStringTrie('/', &extractor{}),
			},
			{
				"dynamic match no name - prefix pipe",
				"/:|",
				NewStringTrie('/', &extractor{}),
			},
			{
				"dynamic match no name - prefix and suffix pipe",
				"/:||",
				NewStringTrie('/', &extractor{}),
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
