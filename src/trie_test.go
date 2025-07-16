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
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				trie := newTrie[int]()
				for k, v := range tc.in {
					trie.Insert(k, &v)
				}

				val, found := trie.Find(tc.search)
				if found {
					t.Errorf(`Trie.Find(%#q), did not expect to find element`, tc.search)
				}
				if val != nil {
					t.Errorf(`Trie.Find(%#q) = %d, expected value to be nil`, tc.search, *val)
				}
			})
		}
	})

	t.Run("found", func(t *testing.T) {
		tests := []struct {
			name   string
			in     map[string]int
			search string
			want   int
		}{
			{
				"one element",
				map[string]int{"/foo": 13},
				"/foo",
				13,
			},
			{
				"multiple elements leaf",
				map[string]int{"/foo": 13, "/foo/bar": 42, "/foo/bar/baz": 15},
				"/foo/bar/baz",
				15,
			},
			{
				"multiple elements non-leaf",
				map[string]int{"/foo": 13, "/foo/bar": 42, "/foo/bar/baz": 15},
				"/foo/bar",
				42,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				trie := newTrie[int]()
				for k, v := range tc.in {
					trie.Insert(k, &v)
				}

				actual, found := trie.Find(tc.search)
				if !found {
					t.Errorf("Trie.Find(%#q) expected to find element", tc.search)
				}
				if *actual != tc.want {
					t.Errorf(`trie["%s"] = %d, wanted %d`, tc.search, actual, tc.want)
				}
			})
		}
	})
}
