package trie

import "testing"

func TestContains(t *testing.T) {
	tests := []struct {
		name string
		base []string
		in   string
		want bool
	}{
		{
			name: "empty trie",
			base: []string{},
			in:   "foo",
			want: false,
		},
		{
			name: "all whitespace inputs",
			base: []string{" ", "   ", "\t\t", "\t \t "},
			in:   " ",
			want: false,
		},
		{
			name: "all whitespace search",
			base: []string{"foo", "bar"},
			in:   "\t ",
			want: false,
		},
		{
			name: "single element contains",
			base: []string{"foo"},
			in:   "foo",
			want: true,
		},
		{
			name: "single element doesn't contain",
			base: []string{"foo"},
			in:   "bar",
			want: false,
		},
		{
			name: "single element contains (case insensitive)",
			base: []string{"foo"},
			in:   "FoO",
			want: true,
		},
		{
			name: "single element contains (whitespace on insert)",
			base: []string{" \tf   o\t\to "},
			in:   "foo",
			want: true,
		},
		{
			name: "single element contains (whitespace on lookup)",
			base: []string{"foo"},
			in:   "\t\t\t\t\t f  o\t o ",
			want: true,
		},
		{
			name: "invalid character",
			base: []string{"\n"},
			in:   "\n",
			want: false,
		},
		{
			name: "multiple elements contains one",
			base: []string{"foo", "bar", "baz"},
			in:   "bar",
			want: true,
		},
		{
			name: "multiple elements contains none",
			base: []string{"foo", "bar", "baz"},
			in:   "qux",
			want: false,
		},
		{
			name: "prefix only (should fail)",
			base: []string{"foo"},
			in:   "fo",
			want: false,
		},
		{
			name: "superstring only (should fail)",
			base: []string{"foo"},
			in:   "fooo",
			want: false,
		},
		{
			name: "one matches, others are near-misses",
			base: []string{"foo", "foobar", "barfoo"},
			in:   "foo",
			want: true,
		},
		{
			name: "case and spacing across multiple entries",
			base: []string{" FooBar ", "QUX", " bAz "},
			in:   "baz",
			want: true,
		},
		{
			name: "partial match from multiple (should fail)",
			base: []string{"foobar", "barfoo"},
			in:   "foo",
			want: false,
		},
		{
			name: "empty string search (should fail)",
			base: []string{"foo", "bar"},
			in:   "",
			want: false,
		},
		{
			name: "empty string inserted (should not match anything)",
			base: []string{""},
			in:   "foo",
			want: false,
		},
		{
			name: "special chars: match allowed single char",
			base: []string{"!"},
			in:   "!",
			want: true,
		},
		{
			name: "special chars: no match for disallowed char",
			base: []string{"!"},
			in:   "@",
			want: false,
		},
		{
			name: "special chars: mixed allowed symbols",
			base: []string{"x-token^id", "x-header~debug"},
			in:   "x-token^id",
			want: true,
		},
		{
			name: "special chars: case insensitive with symbols",
			base: []string{"X-Custom_Header"},
			in:   "x-custom_header",
			want: true,
		},
		{
			name: "special chars: partial match should fail",
			base: []string{"x-meta+tag"},
			in:   "x-meta",
			want: false,
		},
		{
			name: "special chars: superstring match should fail",
			base: []string{"x-meta+tag"},
			in:   "x-meta+tag+extra",
			want: false,
		},
		{
			name: "special chars: edge case lowest symbol",
			base: []string{"!edge"},
			in:   "!edge",
			want: true,
		},
		{
			name: "special chars: edge case highest symbol",
			base: []string{"x-header~end"},
			in:   "x-header~end",
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			trie := NewRuneTrie()
			for _, s := range tc.base {
				trie.Insert(s)
			}

			got := trie.Contains(tc.in)

			if got != tc.want {
				t.Errorf(`Contains(%#q) = %t, wanted %t`, tc.in, got, tc.want)
			}
		})
	}
}
