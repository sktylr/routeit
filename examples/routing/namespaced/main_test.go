package main

import (
	"fmt"
	"testing"

	"github.com/sktylr/routeit"
)

var client = routeit.NewTestClient(GetServer())

func TestGetHelloNotFound(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			"not namespaced",
			"/hello",
		},
		{
			"only locally namespaced",
			"/namespace/hello",
		},
		{
			"incorrect ordering of namespaces",
			"/namespace/api/hello",
		},
		{
			"non-leaf route",
			"/api/namespace",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := client.Get(tc.path)

			res.AssertStatusCode(t, routeit.StatusNotFound)
			wantBody := fmt.Sprintf("404: Not Found. Invalid route: %s", tc.path)
			res.AssertBodyMatchesString(t, wantBody)
			res.AssertHeaderMatches(t, "Content-Length", fmt.Sprintf("%d", len(wantBody)))
			res.AssertHeaderMatches(t, "Content-Type", "text/plain")
		})
	}
}

func TestGetHelloSuccess(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			"top level namespace",
			"/api/hello",
		},
		{
			"local namespace",
			"/api/namespace/hello",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := client.Get(tc.path)

			res.AssertStatusCode(t, routeit.StatusOK)
			wantBody := fmt.Sprintf(`Hello from "%s"`, tc.path)
			res.AssertBodyMatchesString(t, wantBody)
			res.AssertHeaderMatches(t, "Content-Length", fmt.Sprintf("%d", len(wantBody)))
			res.AssertHeaderMatches(t, "Content-Type", "text/plain")
		})
	}
}
