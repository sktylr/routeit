package main

import (
	"crypto/tls"
	"strconv"
	"testing"

	"github.com/sktylr/routeit"
)

func TestServer(t *testing.T) {
	srv := GetServer()
	tests := []struct {
		name     string
		client   routeit.TestClient
		wantBody string
	}{
		{
			name:     "http",
			client:   routeit.NewTestClient(srv),
			wantBody: "Hello world!",
		},
		{
			name:     "https",
			client:   routeit.NewTestTlsClient(srv, &tls.ConnectionState{}),
			wantBody: "Hello world! Thanks for being secure!",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			verify := func(t *testing.T, res *routeit.TestResponse) {
				res.AssertHeaderMatchesString(t, "Content-Type", "text/plain")
				res.AssertHeaderMatchesString(t, "Content-Length", strconv.Itoa(len(tc.wantBody)))
			}

			t.Run("GET", func(t *testing.T) {
				res := tc.client.Get("/hello")
				verify(t, res)
				res.AssertBodyMatchesString(t, tc.wantBody)
			})

			t.Run("HEAD", func(t *testing.T) {
				res := tc.client.Head("/hello")
				verify(t, res)
				res.AssertBodyEmpty(t)
			})
		})
	}
}
