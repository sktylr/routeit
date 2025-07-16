package main

import (
	"testing"

	"github.com/sktylr/routeit"
)

// TODO: routeit cType comparison is case sensitive!

func TestAbout(t *testing.T) {
	client := routeit.NewTestClient(GetServer())
	verifyMeta := func(t *testing.T, res *routeit.TestResponse) {
		t.Helper()
		res.AssertStatusCode(t, routeit.StatusOK)
		res.AssertHeaderMatches(t, "Content-Type", "text/html; charset=utf-8")
		res.AssertHeaderMatches(t, "Content-Length", "650")
	}

	t.Run("GET", func(t *testing.T) {
		res := client.Get("/static/about.html")

		verifyMeta(t, res)
		res.AssertBodyStartsWithString(t, "<!DOCTYPE html>")
		res.AssertBodyContainsString(t, "<title>About - My Server</title>")
		res.AssertBodyContainsString(t, "<h1>About This Server</h1>")
	})

	t.Run("HEAD", func(t *testing.T) {
		res := client.Head("/static/about.html")

		verifyMeta(t, res)
		res.AssertBodyMatchesString(t, "")
	})
}

func TestGopher(t *testing.T) {
	client := routeit.NewTestClient(GetServer())
	verifyMeta := func(t *testing.T, res *routeit.TestResponse) {
		t.Helper()
		res.AssertStatusCode(t, routeit.StatusOK)
		res.AssertHeaderMatches(t, "Content-Type", "image/png")
		res.AssertHeaderMatches(t, "Content-Length", "70548")
	}

	t.Run("GET", func(t *testing.T) {
		res := client.Get("/static/gopher.png")

		verifyMeta(t, res)
		res.AssertBodyStartsWithString(t, "\x89PNG")
	})

	t.Run("HEAD", func(t *testing.T) {
		res := client.Head("/static/gopher.png")

		verifyMeta(t, res)
		res.AssertBodyMatchesString(t, "")
	})
}

func TestHello(t *testing.T) {
	client := routeit.NewTestClient(GetServer())
	verifyMeta := func(t *testing.T, res *routeit.TestResponse) {
		t.Helper()
		res.AssertStatusCode(t, routeit.StatusOK)
		res.AssertHeaderMatches(t, "Content-Type", "text/plain; charset=utf-8")
		res.AssertHeaderMatches(t, "Content-Length", "13")
	}

	t.Run("GET", func(t *testing.T) {
		res := client.Get("/static/hello.txt")

		verifyMeta(t, res)
		res.AssertBodyMatchesString(t, "Hello World!\n")
	})

	t.Run("HEAD", func(t *testing.T) {
		res := client.Head("/static/hello.txt")

		verifyMeta(t, res)
		res.AssertBodyMatchesString(t, "")
	})
}

func TestIndex(t *testing.T) {
	client := routeit.NewTestClient(GetServer())
	verifyMeta := func(t *testing.T, res *routeit.TestResponse) {
		t.Helper()
		res.AssertStatusCode(t, routeit.StatusOK)
		res.AssertHeaderMatches(t, "Content-Type", "text/html; charset=utf-8")
		res.AssertHeaderMatches(t, "Content-Length", "579")
	}

	t.Run("GET", func(t *testing.T) {
		res := client.Get("/static/index.html")

		verifyMeta(t, res)
		res.AssertBodyStartsWithString(t, "<!DOCTYPE html>")
		res.AssertBodyContainsString(t, "<title>Home - My Server</title>")
		res.AssertBodyContainsString(t, "<h1>Welcome to the Home Page</h1>")
	})

	t.Run("HEAD", func(t *testing.T) {
		res := client.Head("/static/index.html")

		verifyMeta(t, res)
		res.AssertBodyMatchesString(t, "")
	})
}

func TestStyles(t *testing.T) {
	client := routeit.NewTestClient(GetServer())
	verifyMeta := func(t *testing.T, res *routeit.TestResponse) {
		t.Helper()
		res.AssertStatusCode(t, routeit.StatusOK)
		res.AssertHeaderMatches(t, "Content-Type", "text/css; charset=utf-8")
		res.AssertHeaderMatches(t, "Content-Length", "868")
	}

	t.Run("GET", func(t *testing.T) {
		res := client.Get("/static/styles.css")

		verifyMeta(t, res)
		res.AssertBodyContainsString(t, `font-family: "Segoe UI", Tahoma, Geneva, Verdana, sans-serif;`)
		res.AssertBodyContainsString(t, "box-shadow: 0 6px 16px rgba(0, 0, 0, 0.15);\n")
	})

	t.Run("HEAD", func(t *testing.T) {
		res := client.Head("/static/styles.css")

		verifyMeta(t, res)
		res.AssertBodyMatchesString(t, "")
	})
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
	verify := func(t *testing.T, res *routeit.TestResponse) {
		t.Helper()
		res.AssertStatusCode(t, routeit.StatusNotFound)
		res.AssertHeaderMatches(t, "Content-Type", "text/plain")
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("GET", func(t *testing.T) {
				res := client.Get(tc.path)
				verify(t, res)
				res.AssertBodyStartsWithString(t, "404: Not Found")
			})

			t.Run("HEAD", func(t *testing.T) {
				res := client.Head(tc.path)
				verify(t, res)
				res.AssertBodyMatchesString(t, "")
			})
		})
	}
}
