package util

import "testing"

func TestStripeDuplicates(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "handles empty list",
			in:   []string{},
			want: []string{},
		},
		{
			name: "handles singleton list",
			in:   []string{"foo"},
			want: []string{"foo"},
		},
		{
			name: "only considers direct equality",
			in:   []string{"foo", "Foo"},
			want: []string{"foo", "Foo"},
		},
		{
			name: "maintains insert order",
			in:   []string{"b", "a", "c", "f", "g", "d"},
			want: []string{"b", "a", "c", "f", "g", "d"},
		},
		{
			name: "maintains order of first occurrence when duplicates",
			in:   []string{"z", "a", "b", "z", "a", "c"},
			want: []string{"z", "a", "b", "c"},
		},
		{
			name: "removes single duplicates of same element",
			in:   []string{"foo", "foo"},
			want: []string{"foo"},
		},
		{
			name: "removes multiple duplicates of same element",
			in:   []string{"foo", "foo", "foo"},
			want: []string{"foo"},
		},
		{
			name: "removes multiple single duplicates of different elements",
			in:   []string{"bar", "foo", "foo", "bar"},
			want: []string{"bar", "foo"},
		},
		{
			name: "removes multiple duplicates of different elements",
			in:   []string{"foo", "foo", "bar", "foo", "bar", "bar"},
			want: []string{"foo", "bar"},
		},
		{
			name: "alternating duplicates",
			in:   []string{"foo", "bar", "foo", "bar", "baz"},
			want: []string{"foo", "bar", "baz"},
		},
		{
			name: "large unique list",
			in:   []string{"a", "b", "c", "d", "e", "f", "g", "h"},
			want: []string{"a", "b", "c", "d", "e", "f", "g", "h"},
		},
		{
			name: "all duplicates",
			in:   []string{"x", "x", "x", "x", "x"},
			want: []string{"x"},
		},
		{
			name: "interleaved duplicates",
			in:   []string{"foo", "bar", "foo", "baz", "foo"},
			want: []string{"foo", "bar", "baz"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out := StripDuplicates(tc.in)

			if len(out) != len(tc.want) {
				t.Fatalf(`length = %d, wanted %d`, len(out), len(tc.want))
			}
			for i, v := range out {
				if v != tc.want[i] {
					t.Errorf(`arr[%d] = %#q, wanted %#q`, i, v, tc.want[i])
				}
			}
		})
	}
}
