package main

import (
	"crypto/tls"
	"testing"

	"github.com/sktylr/routeit"
)

func TestServer(t *testing.T) {
	srv := GetServer()

	t.Run("http redirected", func(t *testing.T) {
		client := routeit.NewTestClient(srv)

		res := client.Get("/echo", "Host", "example.com")

		res.AssertStatusCode(t, routeit.StatusMovedPermanently)
		res.AssertHeaderMatchesString(t, "Location", "https://example.com:8443/echo")
	})

	t.Run("https uses HSTS", func(t *testing.T) {
		client := routeit.NewTestTlsClient(srv, &tls.ConnectionState{})

		res := client.Get("/echo?message=Hello")

		res.AssertStatusCode(t, routeit.StatusOK)
		res.AssertBodyMatchesString(t, "Hello")
		res.AssertHeaderMatchesString(t, "Strict-Transport-Security", "max-age=1; includeSubdomains")
	})
}
