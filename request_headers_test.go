package routeit

import "testing"

func TestHeadersFromRaw(t *testing.T) {
	t.Run("valid inputs", func(t *testing.T) {
		tests := []struct {
			name      string
			raw       [][]byte
			want      map[string][]string
			wantIndex int
		}{
			{
				name: "multi header",
				raw:  [][]byte{[]byte("Host: localhost"), []byte("Content-Type: application/json"), {}},
				want: map[string][]string{
					"Host":         {"localhost"},
					"Content-Type": {"application/json"},
				},
				wantIndex: 2,
			},
			{
				name:      "excessive leading and trailing whitespace",
				raw:       [][]byte{[]byte("X-My-Header:    value    "), {}},
				want:      map[string][]string{"X-My-Header": {"value"}},
				wantIndex: 1,
			},
			{
				name:      "exits after empty lines",
				raw:       [][]byte{[]byte(""), []byte("Host: localhost")},
				wantIndex: 0,
			},
			{
				name:      `allows multiple ":" characters`,
				raw:       [][]byte{[]byte("Host: localhost:433"), {}},
				want:      map[string][]string{"Host": {"localhost:433"}},
				wantIndex: 1,
			},
			{
				name: "stores multiple header entries with the same key",
				raw:  [][]byte{[]byte("Accept: application/json"), []byte("Accept: application/javascript"), []byte("Host: localhost"), []byte("Accept: text/html"), {}},
				want: map[string][]string{
					"Accept": {"application/json", "application/javascript", "text/html"},
					"Host":   {"localhost"},
				},
				wantIndex: 4,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				h, i, err := headersFromRaw(tc.raw)

				if err != nil {
					t.Errorf("expected error to be nil: %v", err)
				}
				if len(h.headers) != len(tc.want) {
					t.Errorf(`headers from raw len(h) = %d, want %d`, len(h.headers), len(tc.want))
				}
				for k, vals := range tc.want {
					verifyPresentAndMatches(t, h.headers, k, vals)
				}
				if i != tc.wantIndex {
					t.Errorf(`last valid header index = %d, wanted %d`, i, tc.wantIndex)
				}
			})
		}
	})

	t.Run("errors", func(t *testing.T) {
		tests := []struct {
			name string
			raw  [][]byte
		}{
			{
				name: "does not allow leading whitespace in keys",
				raw:  [][]byte{[]byte(" My-Header: Value")},
			},
			{
				name: "does not allow trailing whitespace in keys",
				raw:  [][]byte{[]byte("My-Header : Value")},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				h, i, err := headersFromRaw(tc.raw)

				if err == nil {
					t.Fatal("expected error but got nil")
				}
				if h != nil {
					t.Fatal("expected headers to be nil")
				}
				if err.Error() != "400: Bad Request" {
					t.Errorf(`error = %q, wanted "400: Bad Request"`, err.Error())
				}
				if i != -1 {
					t.Errorf(`last valid header index = %d, wanted -1`, i)
				}
			})
		}
	})
}
