package cmp

import "testing"

func TestExactOrWildcard_Matches(t *testing.T) {
	tests := []struct {
		name string
		m    ExactOrWildcard
		in   string
		want bool
	}{
		{
			name: "exact match - success",
			m:    ExactOrWildcard{exact: "hello"},
			in:   "hello",
			want: true,
		},
		{
			name: "exact match - fail",
			m:    ExactOrWildcard{exact: "hello"},
			in:   "hello!",
			want: false,
		},
		{
			name: "wildcard match - prefix and suffix, valid",
			m:    ExactOrWildcard{wc: &wildcard{prefix: "foo", suffix: "bar", minLen: 6}},
			in:   "foobazbar",
			want: true,
		},
		{
			name: "wildcard match - prefix and suffix, too short",
			m:    ExactOrWildcard{wc: &wildcard{prefix: "foo", suffix: "bar", minLen: 6}},
			in:   "foobar",
			want: false,
		},
		{
			name: "wildcard match - prefix only, valid",
			m:    ExactOrWildcard{wc: &wildcard{prefix: "start", suffix: "", minLen: 5}},
			in:   "start123",
			want: true,
		},
		{
			name: "wildcard match - suffix only, valid",
			m:    ExactOrWildcard{wc: &wildcard{prefix: "", suffix: "end", minLen: 3}},
			in:   "123end",
			want: true,
		},
		{
			name: "wildcard match - empty pre/suf, non-empty input",
			m:    ExactOrWildcard{wc: &wildcard{prefix: "", suffix: "", minLen: 0}},
			in:   "anything",
			want: true,
		},
		{
			name: "wildcard match - empty pre/suf, empty input",
			m:    ExactOrWildcard{wc: &wildcard{prefix: "", suffix: "", minLen: 0}},
			in:   "",
			want: false,
		},
		{
			name: "wildcard match - mismatch prefix",
			m:    ExactOrWildcard{wc: &wildcard{prefix: "foo", suffix: "bar", minLen: 6}},
			in:   "bazbar",
			want: false,
		},
		{
			name: "wildcard match - mismatch suffix",
			m:    ExactOrWildcard{wc: &wildcard{prefix: "foo", suffix: "bar", minLen: 6}},
			in:   "foobaz",
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
