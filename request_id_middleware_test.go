package routeit

import "testing"

func TestRequestIdMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		header     string
		provider   RequestIdProvider
		wantId     string
		wantHeader string
	}{
		{
			name:       "no header provided, uses default",
			provider:   func(r *Request) string { return "id" },
			wantId:     "id",
			wantHeader: "X-Request-Id",
		},
		{
			name:       "uses custom header",
			header:     "X-My-Request-Id",
			provider:   func(r *Request) string { return "id" },
			wantId:     "id",
			wantHeader: "X-My-Request-Id",
		},
		{
			name:   "uses provider correctly",
			header: "With-Provider",
			provider: func(r *Request) string {
				return r.Path()
			},
			wantId:     "/foo",
			wantHeader: "With-Provider",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mware := requestIdMiddleware(tc.provider, tc.header)
			req := NewTestRequest(t, "/foo", GET, TestRequestOptions{})

			res, proceeded, err := TestMiddleware(mware, req)

			if err != nil {
				t.Errorf(`err = %v, wanted nil`, err)
			}
			if !proceeded {
				t.Error("did not proceed")
			}
			if req.req.id != tc.wantId {
				t.Errorf(`id = %+q, wanted %+q`, req.req.id, tc.wantId)
			}
			res.AssertHeaderMatchesString(t, tc.wantHeader, tc.wantId)
		})
	}
}
