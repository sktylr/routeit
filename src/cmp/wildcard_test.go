package cmp

import "testing"

func TestMatches(t *testing.T) {
	tests := []struct {
		name string
		pre  string
		suf  string
		in   string
		want bool
	}{
		{
			name: "match with prefix and suffix",
			pre:  "foo",
			suf:  "bar",
			in:   "foobar",
			want: false,
		},
		{
			name: "match with prefix and suffix (longer)",
			pre:  "foo",
			suf:  "bar",
			in:   "foobazbar",
			want: true,
		},
		{
			name: "match with only prefix",
			pre:  "start",
			suf:  "",
			in:   "start123",
			want: true,
		},
		{
			name: "match with only prefix but too short",
			pre:  "start",
			suf:  "",
			in:   "start",
			want: false,
		},
		{
			name: "match with only suffix",
			pre:  "",
			suf:  "end",
			in:   "123end",
			want: true,
		},
		{
			name: "match with only suffix but too short",
			pre:  "",
			suf:  "end",
			in:   "end",
			want: false,
		},
		{
			name: "match with no prefix or suffix and non-empty input",
			pre:  "",
			suf:  "",
			in:   "anything",
			want: true,
		},
		{
			name: "match with no prefix or suffix and empty input",
			pre:  "",
			suf:  "",
			in:   "",
			want: false,
		},
		{
			name: "non-matching prefix",
			pre:  "foo",
			suf:  "",
			in:   "barfoo",
			want: false,
		},
		{
			name: "non-matching suffix",
			pre:  "",
			suf:  "bar",
			in:   "barfoo",
			want: false,
		},
		{
			name: "non-matching prefix and suffix",
			pre:  "foo",
			suf:  "bar",
			in:   "buzz",
			want: false,
		},
		{
			name: "no match overlapping prefix and suffix",
			pre:  "pref",
			suf:  "eference",
			in:   "preference",
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wc := newWildcard(tc.pre, tc.suf)

			got := wc.Matches(tc.in)

			if got != tc.want {
				t.Errorf(`Wildcard[%+v].Matches(%#q) = %t, wanted %t`, wc, tc.in, got, tc.want)
			}
		})
	}
}
