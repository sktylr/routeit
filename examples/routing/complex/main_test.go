package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/sktylr/routeit"
)

func TestServer(t *testing.T) {
	tests := []struct {
		path          string
		wantPathParam string
		wantHandler   string
	}{
		{
			path:        "/path",
			wantHandler: "/:path",
		},
		{
			path:        "/prefix-only",
			wantHandler: "/:path|prefix",
		},
		{
			path:        "/only-suffix",
			wantHandler: "/:path||suffix",
		},
		{
			path:        "/prefix-suffix",
			wantHandler: "/:path|prefix|suffix",
		},
		{
			// Although this starts with /prefix, it must be followed by at
			// least 1 alphanumeric character (or - or _) to be considered a
			// match against the /:path|prefix route
			path:        "/prefix",
			wantHandler: "/:path",
		},
		{
			// Although this ends with "suffix", it must be preceded by at
			// least 1 alphanumeric character (or - or _) to be considered a
			// match against the /:path||suffix route
			path:        "/suffix",
			wantHandler: "/:path",
		},
		{
			// This path component matches against the /:path endpoint. It
			// contains the %2F sequence, which is the url encoding for "/".
			// routeit will decode the path **after** splitting it up into path
			// segments, so this request will still reach the correct /:path
			// endpoint. However, when the server extracts the path parameter,
			// it will be safely decoded, meaning the slash will be present
			// instead of %2F. This slash is **not** a control character used
			// in routing due to it being escaped in the request.
			path:          `/hello%2Fthere`,
			wantPathParam: "hello/there",
			wantHandler:   "/:path",
		},
	}
	client := routeit.NewTestClient(GetServer())

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			wantPathParam := tc.wantPathParam
			if wantPathParam == "" {
				wantPathParam = tc.path[1:]
			}
			wantBody := HelloResponse{
				IncomingUrl:  "/" + wantPathParam,
				HandlerRoute: tc.wantHandler,
				// This demonstrates that the whole path component is extracted,
				// regardless of any prefixes or suffixes present.
				PathParam: wantPathParam,
			}
			verify := func(t *testing.T, res *routeit.TestResponse) {
				res.AssertStatusCode(t, routeit.StatusOK)
				res.AssertHeaderMatchesString(t, "Content-Type", "application/json")
				b, _ := json.Marshal(wantBody)
				res.AssertHeaderMatchesString(t, "Content-Length", fmt.Sprintf("%d", len(b)))
			}

			t.Run("GET", func(t *testing.T) {
				res := client.Get(tc.path)

				verify(t, res)
				var body HelloResponse
				res.BodyToJson(t, &body)
				if !reflect.DeepEqual(body, wantBody) {
					t.Errorf(`body = %#q, wanted %#q`, body, wantBody)
				}
			})

			t.Run("HEAD", func(t *testing.T) {
				res := client.Head(tc.path)

				verify(t, res)
				res.AssertBodyEmpty(t)
			})

			t.Run("POST", func(t *testing.T) {
				res := client.PostText(tc.path, "This will not be accepted")

				res.AssertStatusCode(t, routeit.StatusMethodNotAllowed)
				res.AssertHeaderMatches(t, "Allow", []string{"GET", "HEAD", "OPTIONS"})
			})

			t.Run("OPTIONS", func(t *testing.T) {
				res := client.Options(tc.path)

				res.AssertStatusCode(t, routeit.StatusNoContent)
				res.AssertHeaderMatches(t, "Allow", []string{"GET", "HEAD", "OPTIONS"})
				res.AssertBodyEmpty(t)
			})
		})
	}
}
