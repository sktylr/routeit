package routeit

import "testing"

func TestQueryParams(t *testing.T) {
	tests := []struct {
		name        string
		rawQuery    string
		key         string
		wantFound   bool
		wantOnly    string
		wantOnlyOk  bool
		wantOnlyErr bool
		wantFirst   string
		wantFirstOk bool
		wantLast    string
		wantLastOk  bool
	}{
		{
			name:        "no param",
			key:         "x",
			wantFound:   false,
			wantOnlyOk:  false,
			wantFirstOk: false,
			wantLastOk:  false,
		},
		{
			name:        "single param",
			rawQuery:    "x=hello",
			key:         "x",
			wantFound:   true,
			wantOnly:    "hello",
			wantOnlyOk:  true,
			wantFirst:   "hello",
			wantFirstOk: true,
			wantLast:    "hello",
			wantLastOk:  true,
		},
		{
			name:        "multiple param values",
			rawQuery:    "x=a&x=b&x=c",
			key:         "x",
			wantFound:   true,
			wantOnlyErr: true,
			wantOnlyOk:  true,
			wantFirst:   "a",
			wantFirstOk: true,
			wantLast:    "c",
			wantLastOk:  true,
		},
		{
			name:        "empty value",
			rawQuery:    "x=",
			key:         "x",
			wantFound:   true,
			wantOnly:    "",
			wantOnlyOk:  true,
			wantFirst:   "",
			wantFirstOk: true,
			wantLast:    "",
			wantLastOk:  true,
		},
		{
			name:        "empty and non-empty values",
			rawQuery:    "x=&x=val",
			key:         "x",
			wantFound:   true,
			wantOnlyErr: true,
			wantOnlyOk:  true,
			wantFirst:   "",
			wantFirstOk: true,
			wantLast:    "val",
			wantLastOk:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			queries := newQueryParams()
			err := parseQueryParams(tc.rawQuery, queries)
			if err != nil {
				t.Fatalf("parseUri(%q) failed: %v", tc.rawQuery, err)
			}

			t.Run("All", func(t *testing.T) {
				vals, found := queries.All(tc.key)
				if found != tc.wantFound {
					t.Errorf("All(%q): found = %v, want %v", tc.key, found, tc.wantFound)
				}
				if found && len(vals) == 0 {
					t.Errorf("All(%q): returned empty slice unexpectedly", tc.key)
				}
			})

			t.Run("Only", func(t *testing.T) {
				got, found, err := queries.Only(tc.key)
				if tc.wantOnlyErr && err == nil {
					t.Errorf("Only(%q): expected error, got nil", tc.key)
				}
				if !tc.wantOnlyErr && err != nil {
					t.Errorf("Only(%q): unexpected error: %v", tc.key, err)
				}
				if found != tc.wantOnlyOk {
					t.Errorf("Only(%q): found = %v, want %v", tc.key, found, tc.wantOnlyOk)
				}
				if got != tc.wantOnly {
					t.Errorf("Only(%q): got = %q, want %q", tc.key, got, tc.wantOnly)
				}
			})

			t.Run("First", func(t *testing.T) {
				got, found := queries.First(tc.key)
				if found != tc.wantFirstOk {
					t.Errorf("Fist(%q): found = %v, want %v", tc.key, found, tc.wantFirstOk)
				}
				if got != tc.wantFirst {
					t.Errorf("First(%q): got = %q, want %q", tc.key, got, tc.wantFirst)
				}
			})

			t.Run("Last", func(t *testing.T) {
				got, found := queries.Last(tc.key)
				if found != tc.wantLastOk {
					t.Errorf("Last(%q): found = %v, want %v", tc.key, found, tc.wantLastOk)
				}
				if got != tc.wantLast {
					t.Errorf("Last(%q): got = %q, want %q", tc.key, got, tc.wantLast)
				}
			})
		})
	}
}
