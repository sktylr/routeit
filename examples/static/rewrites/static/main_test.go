package main

import (
	"reflect"
	"testing"

	"github.com/sktylr/routeit"
)

func TestIndex(t *testing.T) {
	routes := []string{"/", "/statics/index.html"}
	verify := func(t *testing.T, res *routeit.TestResponse) {
		t.Helper()
		res.AssertStatusCode(t, routeit.StatusOK)
		res.RefuteHeaderPresent(t, "Allow")
		res.AssertHeaderMatches(t, "Content-Type", "text/html; charset=utf-8")
		res.AssertHeaderMatches(t, "Content-Length", "563")
	}
	client := routeit.NewTestClient(GetServer())

	for _, r := range routes {
		t.Run(r, func(t *testing.T) {
			t.Run("GET", func(t *testing.T) {
				res := client.Get(r)

				verify(t, res)
				res.AssertBodyStartsWithString(t, "<!DOCTYPE html>")
				res.AssertBodyContainsString(t, "<title>Home - URL Rewrites</title>")
				res.AssertBodyContainsString(t, "<h1>Welcome to my example URL rewritten website</h1>")
			})

			t.Run("HEAD", func(t *testing.T) {
				res := client.Head(r)

				verify(t, res)
				res.AssertBodyEmpty(t)
			})

			t.Run("OPTIONS", func(t *testing.T) {
				res := client.Options(r)

				res.AssertStatusCode(t, routeit.StatusNoContent)
				res.AssertBodyEmpty(t)
				res.AssertHeaderMatches(t, "Allow", "GET, HEAD, OPTIONS")
			})

			t.Run("POST", func(t *testing.T) {
				res := client.PostText(r, "This should fail")

				res.AssertStatusCode(t, routeit.StatusMethodNotAllowed)
				res.AssertBodyMatchesString(t, "405: Method Not Allowed")
				res.AssertHeaderMatches(t, "Allow", "GET, HEAD, OPTIONS")
			})
		})
	}
}

func TestAbout(t *testing.T) {
	routes := []string{"/about", "/statics/about.html"}
	verify := func(t *testing.T, res *routeit.TestResponse) {
		t.Helper()
		res.AssertStatusCode(t, routeit.StatusOK)
		res.RefuteHeaderPresent(t, "Allow")
		res.AssertHeaderMatches(t, "Content-Type", "text/html; charset=utf-8")
		res.AssertHeaderMatches(t, "Content-Length", "1650")
	}
	client := routeit.NewTestClient(GetServer())

	for _, r := range routes {
		t.Run(r, func(t *testing.T) {
			t.Run("GET", func(t *testing.T) {
				res := client.Get(r)

				verify(t, res)
				res.AssertBodyStartsWithString(t, "<!DOCTYPE html>")
				res.AssertBodyContainsString(t, "<title>About - URL Rewrites</title>")
				res.AssertBodyContainsString(t, "<h1>Learn more about URL rewrites</h1>")
			})

			t.Run("HEAD", func(t *testing.T) {
				res := client.Head(r)

				verify(t, res)
				res.AssertBodyEmpty(t)
			})

			t.Run("OPTIONS", func(t *testing.T) {
				res := client.Options(r)

				res.AssertStatusCode(t, routeit.StatusNoContent)
				res.AssertBodyEmpty(t)
				res.AssertHeaderMatches(t, "Allow", "GET, HEAD, OPTIONS")
			})

			t.Run("POST", func(t *testing.T) {
				res := client.PostText(r, "This should fail")

				res.AssertStatusCode(t, routeit.StatusMethodNotAllowed)
				res.AssertBodyMatchesString(t, "405: Method Not Allowed")
				res.AssertHeaderMatches(t, "Allow", "GET, HEAD, OPTIONS")
			})
		})
	}
}

func TestTarget(t *testing.T) {
	routes := []string{"/favicon.ico", "/statics/images/target.png", "/target.png"}
	verify := func(t *testing.T, res *routeit.TestResponse) {
		t.Helper()
		res.AssertStatusCode(t, routeit.StatusOK)
		res.RefuteHeaderPresent(t, "Allow")
		res.AssertHeaderMatches(t, "Content-Type", "image/png")
		res.AssertHeaderMatches(t, "Content-Length", "14513")
	}
	client := routeit.NewTestClient(GetServer())

	for _, r := range routes {
		t.Run(r, func(t *testing.T) {
			t.Run("GET", func(t *testing.T) {
				res := client.Get(r)

				verify(t, res)
				res.AssertBodyStartsWithString(t, "\x89PNG")
			})

			t.Run("HEAD", func(t *testing.T) {
				res := client.Head(r)

				verify(t, res)
				res.AssertBodyEmpty(t)
			})

			t.Run("OPTIONS", func(t *testing.T) {
				res := client.Options(r)

				res.AssertStatusCode(t, routeit.StatusNoContent)
				res.AssertBodyEmpty(t)
				res.AssertHeaderMatches(t, "Allow", "GET, HEAD, OPTIONS")
			})

			t.Run("POST", func(t *testing.T) {
				res := client.PostText(r, "This should fail")

				res.AssertStatusCode(t, routeit.StatusMethodNotAllowed)
				res.AssertBodyMatchesString(t, "405: Method Not Allowed")
				res.AssertHeaderMatches(t, "Allow", "GET, HEAD, OPTIONS")
			})
		})
	}
}

func TestStyles(t *testing.T) {
	routes := []string{"/styles.css", "/statics/css/styles.css"}
	verify := func(t *testing.T, res *routeit.TestResponse) {
		t.Helper()
		res.AssertStatusCode(t, routeit.StatusOK)
		res.RefuteHeaderPresent(t, "Allow")
		res.AssertHeaderMatches(t, "Content-Type", "text/css; charset=utf-8")
		res.AssertHeaderMatches(t, "Content-Length", "1113")
	}
	client := routeit.NewTestClient(GetServer())

	for _, r := range routes {
		t.Run(r, func(t *testing.T) {
			t.Run("GET", func(t *testing.T) {
				res := client.Get(r)

				verify(t, res)
				res.AssertBodyStartsWithString(t, "* {")
				res.AssertBodyContainsString(t, "transition: background 0.2s ease;")
				res.AssertBodyContainsString(t, "#response-output {")
			})

			t.Run("HEAD", func(t *testing.T) {
				res := client.Head(r)

				verify(t, res)
				res.AssertBodyEmpty(t)
			})

			t.Run("OPTIONS", func(t *testing.T) {
				res := client.Options(r)

				res.AssertStatusCode(t, routeit.StatusNoContent)
				res.AssertBodyEmpty(t)
				res.AssertHeaderMatches(t, "Allow", "GET, HEAD, OPTIONS")
			})

			t.Run("POST", func(t *testing.T) {
				res := client.PostText(r, "This should fail")

				res.AssertStatusCode(t, routeit.StatusMethodNotAllowed)
				res.AssertBodyMatchesString(t, "405: Method Not Allowed")
				res.AssertHeaderMatches(t, "Allow", "GET, HEAD, OPTIONS")
			})
		})
	}
}

func TestHello(t *testing.T) {
	routes := []string{"/rewrite/me/please", "/hello/there"}
	verify := func(t *testing.T, res *routeit.TestResponse) {
		t.Helper()
		res.AssertStatusCode(t, routeit.StatusOK)
		res.RefuteHeaderPresent(t, "Allow")
		res.AssertHeaderMatches(t, "Content-Type", "application/json")
		res.AssertHeaderMatches(t, "Content-Length", "180")
	}
	client := routeit.NewTestClient(GetServer())

	for _, r := range routes {
		t.Run(r, func(t *testing.T) {
			t.Run("GET", func(t *testing.T) {
				res := client.Get(r)
				wantBody := Hello{
					Name:    "/hello/there endpoint",
					Message: `This single handler responds to two edge request paths: "/hello/there" and "/rewrite/me/please", which is handled via URL rewrites.`,
				}

				verify(t, res)
				var body Hello
				res.BodyToJson(t, &body)
				if !reflect.DeepEqual(body, wantBody) {
					t.Errorf(`Body() = %#q, wanted %#q`, body, wantBody)
				}
			})

			t.Run("HEAD", func(t *testing.T) {
				res := client.Head(r)

				verify(t, res)
				res.AssertBodyEmpty(t)
			})

			t.Run("OPTIONS", func(t *testing.T) {
				res := client.Options(r)

				res.AssertStatusCode(t, routeit.StatusNoContent)
				res.AssertBodyEmpty(t)
				res.AssertHeaderMatches(t, "Allow", "GET, HEAD, OPTIONS")
			})

			t.Run("POST", func(t *testing.T) {
				res := client.PostText(r, "This should fail")

				res.AssertStatusCode(t, routeit.StatusMethodNotAllowed)
				res.AssertBodyMatchesString(t, "405: Method Not Allowed")
				res.AssertHeaderMatches(t, "Allow", "GET, HEAD, OPTIONS")
			})
		})
	}
}
