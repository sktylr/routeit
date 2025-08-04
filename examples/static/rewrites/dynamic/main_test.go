package main

import (
	"fmt"
	"testing"

	"github.com/sktylr/routeit"
)

var client = routeit.NewTestClient(GetServer())

func TestStaticHtml(t *testing.T) {
	routes := []struct {
		name       string
		paths      []string
		wantCl     string
		wantTitle  string
		wantHeader string
	}{
		{
			name:       "index",
			paths:      []string{"/", "/assets/index.html"},
			wantCl:     "496",
			wantTitle:  "Home",
			wantHeader: "Welcome to my example dynamic URL rewritten website",
		},
		{
			name:       "contact",
			paths:      []string{"/contact", "/assets/contact.html"},
			wantCl:     "1466",
			wantTitle:  "Contact",
			wantHeader: "Contact Us",
		},
		{
			name:       "about",
			paths:      []string{"/about", "/assets/about.html"},
			wantCl:     "634",
			wantTitle:  "About",
			wantHeader: "About This Demo",
		},
		{
			name:       "joke",
			paths:      []string{"/joke", "/assets/joke.html"},
			wantCl:     "952",
			wantTitle:  "Joke",
			wantHeader: "Need a Laugh?",
		},
	}

	for _, tc := range routes {
		verify := func(t *testing.T, res *routeit.TestResponse) {
			t.Helper()
			res.AssertStatusCode(t, routeit.StatusOK)
			res.RefuteHeaderPresent(t, "Allow")
			res.AssertHeaderMatchesString(t, "Content-Type", "text/html; charset=utf-8")
			res.AssertHeaderMatchesString(t, "Content-Length", tc.wantCl)
		}

		t.Run(tc.name, func(t *testing.T) {
			for _, p := range tc.paths {
				t.Run("GET", func(t *testing.T) {
					res := client.Get(p)

					verify(t, res)
					res.AssertBodyStartsWithString(t, "<!DOCTYPE html>")
					res.AssertBodyContainsString(t, fmt.Sprintf("<title>%s</title>", tc.wantTitle))
					res.AssertBodyContainsString(t, fmt.Sprintf("<h1>%s</h1>", tc.wantHeader))
				})

				t.Run("HEAD", func(t *testing.T) {
					res := client.Head(p)

					verify(t, res)
					res.AssertBodyEmpty(t)
				})

				t.Run("OPTIONS", func(t *testing.T) {
					res := client.Options(p)

					res.AssertStatusCode(t, routeit.StatusNoContent)
					res.AssertBodyEmpty(t)
					res.AssertHeaderMatches(t, "Allow", []string{"GET", "HEAD", "OPTIONS"})
				})

				t.Run("POST", func(t *testing.T) {
					res := client.PostText(p, "This should fail")

					res.AssertStatusCode(t, routeit.StatusMethodNotAllowed)
					res.AssertBodyMatchesString(t, "405: Method Not Allowed")
					res.AssertHeaderMatches(t, "Allow", []string{"GET", "HEAD", "OPTIONS"})
				})
			}
		})
	}
}

func TestFavicon(t *testing.T) {
	routes := []string{"/favicon.ico", "/assets/images/tnt.png"}
	verify := func(t *testing.T, res *routeit.TestResponse) {
		t.Helper()
		res.AssertStatusCode(t, routeit.StatusOK)
		res.RefuteHeaderPresent(t, "Allow")
		res.AssertHeaderMatchesString(t, "Content-Type", "image/png")
		res.AssertHeaderMatchesString(t, "Content-Length", "72276")
	}

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
				res.AssertHeaderMatches(t, "Allow", []string{"GET", "HEAD", "OPTIONS"})
			})

			t.Run("POST", func(t *testing.T) {
				res := client.PostText(r, "This should fail")

				res.AssertStatusCode(t, routeit.StatusMethodNotAllowed)
				res.AssertBodyMatchesString(t, "405: Method Not Allowed")
				res.AssertHeaderMatches(t, "Allow", []string{"GET", "HEAD", "OPTIONS"})
			})
		})
	}
}

func TestStyles(t *testing.T) {
	routes := []string{"/styles", "/assets/css/styles.css"}
	verify := func(t *testing.T, res *routeit.TestResponse) {
		t.Helper()
		res.AssertStatusCode(t, routeit.StatusOK)
		res.RefuteHeaderPresent(t, "Allow")
		res.AssertHeaderMatchesString(t, "Content-Type", "text/css; charset=utf-8")
		res.AssertHeaderMatchesString(t, "Content-Length", "503")
	}

	for _, r := range routes {
		t.Run(r, func(t *testing.T) {
			t.Run("GET", func(t *testing.T) {
				res := client.Get(r)

				verify(t, res)
				res.AssertBodyStartsWithString(t, "body {\n")
				res.AssertBodyContainsString(t, "nav a:hover {\n  text-decoration: underline;\n}")
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
				res.AssertHeaderMatches(t, "Allow", []string{"GET", "HEAD", "OPTIONS"})
			})

			t.Run("POST", func(t *testing.T) {
				res := client.PostText(r, "This should fail")

				res.AssertStatusCode(t, routeit.StatusMethodNotAllowed)
				res.AssertBodyMatchesString(t, "405: Method Not Allowed")
				res.AssertHeaderMatches(t, "Allow", []string{"GET", "HEAD", "OPTIONS"})
			})
		})
	}
}

func TestContactApi(t *testing.T) {
	in := ContactForm{
		Name:    "test",
		Email:   "test@test.com",
		Message: "From the test!",
	}

	t.Run("happy", func(t *testing.T) {
		res := client.PostJson("/api/contact", in)

		res.AssertStatusCode(t, routeit.StatusCreated)
		want := "Thanks for your message test!"
		res.AssertHeaderMatchesString(t, "Content-Length", fmt.Sprintf("%d", len(want)))
		res.AssertHeaderMatchesString(t, "Content-Type", "text/plain")
		res.AssertBodyMatchesString(t, want)
	})

	t.Run("bad", func(t *testing.T) {
		// POST ing against /contact is already tested in [TestStaticHtml]
		t.Run("not found", func(t *testing.T) {
			res := client.PostJson("/api/contact/me", in)

			res.AssertStatusCode(t, routeit.StatusNotFound)
		})

		t.Run("invalid method", func(t *testing.T) {
			res := client.Get("/api/contact")

			res.AssertStatusCode(t, routeit.StatusMethodNotAllowed)
			res.AssertHeaderMatches(t, "Allow", []string{"POST", "OPTIONS"})
		})

		t.Run("unsupported content type", func(t *testing.T) {
			res := client.PostText("/api/contact", "I am upset")

			res.AssertStatusCode(t, routeit.StatusUnsupportedMediaType)
			res.AssertHeaderMatchesString(t, "Accept", "application/json")
		})
	})
}
