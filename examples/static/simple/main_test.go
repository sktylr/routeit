package main

import (
	"testing"

	"github.com/sktylr/routeit"
)

func TestGetAbout(t *testing.T) {
	client := routeit.NewTestClient(GetServer())

	res := client.Get("/static/about.html")

	res.AssertStatusCode(t, routeit.StatusOK)
	res.AssertHeaderMatches(t, "Content-Type", "text/html; charset=utf-8")
	res.AssertBodyStartsWithString(t, "<!DOCTYPE html>")
	res.AssertBodyContainsString(t, "<title>About - My Server</title>")
	res.AssertBodyContainsString(t, "<h1>About This Server</h1>")
}

func TestGetGopher(t *testing.T) {
	client := routeit.NewTestClient(GetServer())

	res := client.Get("/static/gopher.png")

	res.AssertStatusCode(t, routeit.StatusOK)
	res.AssertHeaderMatches(t, "Content-Type", "image/png")
	res.AssertBodyStartsWithString(t, "\x89PNG")
}

func TestGetHello(t *testing.T) {
	client := routeit.NewTestClient(GetServer())

	res := client.Get("/static/hello.txt")

	res.AssertStatusCode(t, routeit.StatusOK)
	res.AssertHeaderMatches(t, "Content-Type", "text/plain; charset=utf-8")
	res.AssertBodyMatchesString(t, "Hello World!\n")
}

func TestGetIndex(t *testing.T) {
	client := routeit.NewTestClient(GetServer())

	res := client.Get("/static/index.html")

	res.AssertStatusCode(t, routeit.StatusOK)
	res.AssertHeaderMatches(t, "Content-Type", "text/html; charset=utf-8")
	res.AssertBodyStartsWithString(t, "<!DOCTYPE html>")
	res.AssertBodyContainsString(t, "<title>Home - My Server</title>")
	res.AssertBodyContainsString(t, "<h1>Welcome to the Home Page</h1>")
}

func TestGetStyles(t *testing.T) {
	client := routeit.NewTestClient(GetServer())

	res := client.Get("/static/styles.css")

	res.AssertStatusCode(t, routeit.StatusOK)
	res.AssertHeaderMatches(t, "Content-Type", "text/css; charset=utf-8")
	res.AssertBodyContainsString(t, `font-family: "Segoe UI", Tahoma, Geneva, Verdana, sans-serif;`)
	res.AssertBodyContainsString(t, "box-shadow: 0 6px 16px rgba(0, 0, 0, 0.15);\n")
}

func TestGetNotFound(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"outside static - about.html", "/about.html"},
		{"outside static - gopher.png", "/gopher.png"},
		{"outside static - hello.txt", "/hello.txt"},
		{"outside static - index.html", "/index.html"},
		{"outside static - permission-denied.txt", "/permission-denied.txt"},
		{"outside static - styles.css", "/styles.css"},
		{"back-tracking", "/static/../main.go"},
		{"not found file", "/static/not-found.txt"},
		{"missing extension", "/static/about"},
		{"incorrect extension", "/static/about.txt"},
	}
	client := routeit.NewTestClient(GetServer())

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := client.Get(tc.path)

			res.AssertStatusCode(t, routeit.StatusNotFound)
			res.AssertHeaderMatches(t, "Content-Type", "text/plain")
			res.AssertBodyStartsWithString(t, "404: Not Found")
		})
	}
}
