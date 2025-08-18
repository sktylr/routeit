package errors

import (
	"errors"
	"testing"

	"github.com/sktylr/routeit"
	"github.com/sktylr/routeit/examples/todo/db"
)

func TestErrorMapper(t *testing.T) {
	tests := []struct {
		name string
		in   error
		want *routeit.HttpError
	}{
		{
			name: "no match",
			in:   errors.New("my error"),
		},
		{
			name: "duplicate key",
			in:   db.ErrDuplicateKey,
			want: routeit.ErrBadRequest(),
		},
		{
			name: "item not found",
			in:   db.ErrItemNotFound,
			want: routeit.ErrNotFound(),
		},
		{
			name: "list not found",
			in:   db.ErrListNotFound,
			want: routeit.ErrNotFound(),
		},
		{
			name: "database issue",
			in:   db.ErrDatabaseIssue,
			want: routeit.ErrServiceUnavailable(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out := ErrorMapper(tc.in)
			if !errors.Is(out, tc.want) {
				t.Errorf(`ErrorMapper(%v) = %v, wanted %v`, tc.in, out, tc.want)
			}
		})
	}
}
