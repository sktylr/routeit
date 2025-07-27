package cmp

import "testing"

func TestExactOrWildcard_Matches(t *testing.T) {
	tests := []struct {
		name string
		m    *ExactOrWildcard
		in   string
		want bool
	}{
		{
			name: "exact match - success",
			m:    NewExactMatcher("hello"),
			in:   "hello",
			want: true,
		},
		{
			name: "exact match - fail",
			m:    NewExactMatcher("hello"),
			in:   "hello!",
			want: false,
		},
		{
			name: "wildcard match - prefix and suffix, valid",
			m:    NewWildcardMatcher("foo", "bar"),
			in:   "foobazbar",
			want: true,
		},
		{
			name: "wildcard match - prefix and suffix, too short",
			m:    NewWildcardMatcher("foo", "bar"),
			in:   "foobar",
			want: false,
		},
		{
			name: "wildcard match - prefix only, valid",
			m:    NewWildcardMatcher("start", ""),
			in:   "start123",
			want: true,
		},
		{
			name: "wildcard match - suffix only, valid",
			m:    NewWildcardMatcher("", "end"),
			in:   "123end",
			want: true,
		},
		{
			name: "wildcard match - empty pre/suf, non-empty input",
			m:    NewWildcardMatcher("", ""),
			in:   "anything",
			want: true,
		},
		{
			name: "wildcard match - empty pre/suf, empty input",
			m:    NewWildcardMatcher("", ""),
			in:   "",
			want: false,
		},
		{
			name: "wildcard match - mismatch prefix",
			m:    NewWildcardMatcher("foo", "bar"),
			in:   "bazbar",
			want: false,
		},
		{
			name: "wildcard match - mismatch suffix",
			m:    NewWildcardMatcher("foo", "bar"),
			in:   "foobaz",
			want: false,
		},
		{
			name: "wildcard match - prefix and suffix, dynamic matcher returns true",
			m:    NewDynamicWildcardMatcher("foo", "bar", func(seg string) bool { return seg == "baz" }),
			in:   "foobazbar",
			want: true,
		},
		{
			name: "wildcard match - prefix and suffix, dynamic matcher returns false",
			m:    NewDynamicWildcardMatcher("foo", "bar", func(seg string) bool { return seg != "baz" }),
			in:   "foobazbar",
			want: false,
		},
		{
			name: "wildcard match - prefix and suffix, too short, dynamic matcher not called",
			m:    NewDynamicWildcardMatcher("foo", "bar", func(seg string) bool { panic(seg) }),
			in:   "foobar",
			want: false,
		},
		{
			name: "wildcard match - prefix only, dynamic matcher returns true",
			m:    NewDynamicWildcardMatcher("start", "", func(seg string) bool { return seg == "123" }),
			in:   "start123",
			want: true,
		},
		{
			name: "wildcard match - prefix only, dynamic matcher returns false",
			m:    NewDynamicWildcardMatcher("start", "", func(seg string) bool { return seg != "123" }),
			in:   "start123",
			want: false,
		},
		{
			name: "wildcard match - suffix only, dynamic matcher returns true",
			m:    NewDynamicWildcardMatcher("", "end", func(seg string) bool { return seg == "123" }),
			in:   "123end",
			want: true,
		},
		{
			name: "wildcard match - suffix only, dynamic matcher returns false",
			m:    NewDynamicWildcardMatcher("", "end", func(seg string) bool { return seg != "123" }),
			in:   "123end",
			want: false,
		},
		{
			name: "wildcard match - empty pre/suf, non-empty input, dynamic matcher returns true",
			m:    NewDynamicWildcardMatcher("", "", func(seg string) bool { return seg == "anything" }),
			in:   "anything",
			want: true,
		},
		{
			name: "wildcard match - empty pre/suf, non-empty input, dynamic matcher returns false",
			m:    NewDynamicWildcardMatcher("", "", func(seg string) bool { return seg != "anything" }),
			in:   "anything",
			want: false,
		},
		{
			name: "wildcard match - empty pre/suf, empty input, dynamic matcher not called",
			m:    NewDynamicWildcardMatcher("", "", func(seg string) bool { panic(seg) }),
			in:   "",
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.m.Matches(tc.in)
			if got != tc.want {
				t.Errorf("[%+v].Matches(%q) = %t, want %t", tc.m, tc.in, got, tc.want)
			}
		})
	}
}
