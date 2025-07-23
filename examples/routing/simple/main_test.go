package main

import (
	"fmt"
	"testing"

	"github.com/sktylr/routeit"
)

var client = routeit.NewTestClient(GetServer())

func TestFoundRoute(t *testing.T) {
	tests := []struct {
		path    string
		message string
	}{
		{"/", "the root"},
		{"/a", "/a"},
		{"/a/heavily/nested", "/a/heavily/nested"},
		{"/a/heavily/nested/route", "/a/heavily/nested/route"},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			res := client.Get(tc.path)

			res.AssertStatusCode(t, routeit.StatusCreated)
			wantBody := fmt.Sprintf("Hello from %s!", tc.message)
			res.AssertBodyMatchesString(t, wantBody)
			res.AssertHeaderMatches(t, "Content-Length", fmt.Sprintf("%d", len(wantBody)))
			res.AssertHeaderMatches(t, "Content-Type", "text/plain")
		})
	}
}

func TestNotFoundRoute(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"non-leaf", "/a/heavily"},
		{"generic not found", "/foo"},
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
