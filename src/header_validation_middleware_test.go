package routeit

import (
	"fmt"
	"testing"
)

func TestHeaderValidationMiddleware(t *testing.T) {
	tests := []struct {
		name              string
		disallow          []string
		headers           []string
		wantFailingHeader string
	}{
		{
			name: "no headers",
		},
		{
			name:              "repeated default header",
			headers:           []string{"Host", "localhost:8080", "Host", "localhost:8080"},
			wantFailingHeader: "Host",
		},
		{
			name:              "repeated default header, different case",
			headers:           []string{"host", "localhost:8080", "Host", "localhost:8080"},
			wantFailingHeader: "Host",
		},
		{
			name:     "non repeated custom header",
			disallow: []string{"X-My-Header"},
			headers:  []string{"X-My-Header", "foo"},
		},
		{
			name:              "repeated custom header",
			disallow:          []string{"X-My-Header"},
			headers:           []string{"X-My-Header", "localhost:8080", "X-My-Header", "localhost:8080"},
			wantFailingHeader: "X-My-Header",
		},
		{
			name:              "repeated custom header, different case",
			disallow:          []string{"X-My-Header"},
			headers:           []string{"X-My-Header", "localhost:8080", "x-my-header", "localhost:8080"},
			wantFailingHeader: "X-My-Header",
		},
		{
			name:              "repeated disallowed header with empty value",
			disallow:          []string{"X-Empty"},
			headers:           []string{"X-Empty", "", "X-Empty", ""},
			wantFailingHeader: "X-Empty",
		},
		{
			name:    "default header appears once",
			headers: []string{"User-Agent", "Go-http-client/1.1"},
		},
		{
			name:    "header not in disallow list appears multiple times, should pass",
			headers: []string{"X-Allowed", "val1", "X-Allowed", "val2"},
		},
		{
			name:              "overlapping disallow and default headers",
			disallow:          []string{"Host"},
			headers:           []string{"Host", "a", "Host", "b"},
			wantFailingHeader: "Host",
		},
		{
			name:              "mixed-case disallowed header, repeated with various casing",
			disallow:          []string{"X-Custom-Header"},
			headers:           []string{"x-custom-header", "abc", "X-CUSTOM-HEADER", "def"},
			wantFailingHeader: "X-Custom-Header",
		},
		{
			name:              "case mismatch in disallow list, but should still be enforced",
			disallow:          []string{"x-case-Mix"},
			headers:           []string{"X-CASE-MIX", "123", "x-case-mix", "456"},
			wantFailingHeader: "x-case-Mix",
		},
		{
			name:    "non-disallowed header repeated with different casing",
			headers: []string{"X-Whatever", "a", "x-whatever", "b"},
		},
		{
			name:              "multiple disallowed headers, one repeated",
			disallow:          []string{"A", "B", "C"},
			headers:           []string{"A", "1", "B", "2", "B", "3", "C", "4"},
			wantFailingHeader: "B",
		},
		{
			name:              "multiple repeated disallowed headers, first caught",
			disallow:          []string{"X-1", "X-2"},
			headers:           []string{"X-1", "a", "X-1", "b", "X-2", "x", "X-2", "y"},
			wantFailingHeader: "X-1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mware := headerValidationMiddleware(tc.disallow)
			req := NewTestRequest(t, "/path", GET, TestRequestOptions{
				Headers: tc.headers,
			})
			wantSuccess := tc.wantFailingHeader == ""

			_, proceeded, err := TestMiddleware(mware, req)
			if proceeded != wantSuccess {
				t.Errorf(`proceeded = %t, want success = %t`, proceeded, wantSuccess)
			}
			if (err == nil) != wantSuccess {
				t.Errorf(`error = %v, want success = %t`, err, wantSuccess)
			}
			if err != nil {
				he, ok := err.(*HttpError)
				if !ok {
					t.Fatalf(`expected to return HttpError, returned %T`, err)
				}
				wantErr := fmt.Sprintf("Header %#q cannot appear more than once", tc.wantFailingHeader)
				if he.message != wantErr {
					t.Errorf(`error.message = %#q, wanted %#q`, he.message, wantErr)
				}
			}
		})
	}
}
