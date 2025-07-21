package main

import (
	"fmt"
	"testing"

	"github.com/sktylr/routeit"
)

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
			paths:      []string{"/", "/static/html/index.html"},
			wantCl:     "698",
			wantTitle:  "Home - URL Rewrites",
			wantHeader: "Welcome to the URL Rewrite Demo",
		},
		{
			name:       "about",
			paths:      []string{"/about", "/static/html/about.html"},
			wantCl:     "459",
			wantTitle:  "About - URL Rewrites",
			wantHeader: "About URL Rewriting",
		},
		{
			name:       "faq",
			paths:      []string{"/faq", "/static/html/faq.html"},
			wantCl:     "761",
			wantTitle:  "FAQ - Rewrites",
			wantHeader: "Frequently Asked Questions",
		},
		{
			name:       "gallery",
			paths:      []string{"/gallery", "/static/html/gallery.html"},
			wantCl:     "675",
			wantTitle:  "Gallery - Visual Rewrites",
			wantHeader: "Visual Rewrite Gallery",
		},
		{
			name:       "story",
			paths:      []string{"/story", "/static/html/story.html"},
			wantCl:     "801",
			wantTitle:  "Story - The Life of a URL",
			wantHeader: "The Tale of a Tiny URL",
		},
		{
			name:       "transform",
			paths:      []string{"/transform", "/static/html/transform.html"},
			wantCl:     "433",
			wantTitle:  "Transform Example",
			wantHeader: "Transform Example",
		},
	}
	client := routeit.NewTestClient(GetServer())

	for _, tc := range routes {
		verify := func(t *testing.T, res *routeit.TestResponse) {
			t.Helper()
			res.AssertStatusCode(t, routeit.StatusOK)
			res.RefuteHeaderPresent(t, "Allow")
			res.AssertHeaderMatches(t, "Content-Type", "text/html; charset=utf-8")
			res.AssertHeaderMatches(t, "Content-Length", tc.wantCl)
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
					res.AssertHeaderMatches(t, "Allow", "GET, HEAD, OPTIONS")
				})

				t.Run("POST", func(t *testing.T) {
					res := client.PostText(p, "This should fail")

					res.AssertStatusCode(t, routeit.StatusMethodNotAllowed)
					res.AssertBodyMatchesString(t, "405: Method Not Allowed")
					res.AssertHeaderMatches(t, "Allow", "GET, HEAD, OPTIONS")
				})
			}
		})
	}
}

func TestStyles(t *testing.T) {
	tests := []struct {
		route       string
		wantCl      string
		partialBody []string
	}{
		{
			route:  "styles.css",
			wantCl: "623",
			partialBody: []string{
				"body {\n  font-family: sans-serif;\n  background: #f9f9f9;",
				"h1 {\n  color: #0066cc;\n}",
				"a:hover {\n  text-decoration: underline;\n}",
			},
		},
		{
			route:       "about.css",
			wantCl:      "27",
			partialBody: []string{"h1 {\n  color: darkgreen;\n}"},
		},
		{
			route:       "gallery.css",
			wantCl:      "47",
			partialBody: []string{"img {\n  max-width: 20em;\n  max-height: 20em;\n}"},
		},
	}
	client := routeit.NewTestClient(GetServer())

	for _, tc := range tests {
		routes := []string{fmt.Sprintf("/%s", tc.route), fmt.Sprintf("/static/css/%s", tc.route)}
		verify := func(t *testing.T, res *routeit.TestResponse) {
			t.Helper()
			res.AssertStatusCode(t, routeit.StatusOK)
			res.RefuteHeaderPresent(t, "Allow")
			res.AssertHeaderMatches(t, "Content-Type", "text/css; charset=utf-8")
			res.AssertHeaderMatches(t, "Content-Length", tc.wantCl)
		}

		for _, r := range routes {
			name := fmt.Sprintf("%s: %s", tc.route, r)
			t.Run(name, func(t *testing.T) {
				t.Run("GET", func(t *testing.T) {
					res := client.Get(r)

					verify(t, res)
					for _, wantB := range tc.partialBody {
						res.AssertBodyContainsString(t, wantB)
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
}

func TestImages(t *testing.T) {
	tests := []struct {
		route  string
		wantCl string
		wantCt string
		start  string
	}{
		{
			route:  "map_flow.png",
			wantCl: "853010",
			wantCt: "image/png",
			start:  "\x89PNG",
		},
		{
			route:  "rewrite_about.png",
			wantCl: "957",
			wantCt: "image/png",
			start:  "\x89PNG",
		},
		{
			route:  "route_contact.jpg",
			wantCl: "4068",
			wantCt: "image/jpeg",
			start:  "\xff\xd8\xff",
		},
		{
			route:  "transform_api.png",
			wantCl: "954",
			wantCt: "image/png",
			start:  "\x89PNG",
		},
	}
	client := routeit.NewTestClient(GetServer())

	for _, tc := range tests {
		routes := []string{fmt.Sprintf("/%s", tc.route), fmt.Sprintf("/static/images/%s", tc.route)}
		verify := func(t *testing.T, res *routeit.TestResponse) {
			t.Helper()
			res.AssertStatusCode(t, routeit.StatusOK)
			res.RefuteHeaderPresent(t, "Allow")
			res.AssertHeaderMatches(t, "Content-Type", tc.wantCt)
			res.AssertHeaderMatches(t, "Content-Length", tc.wantCl)
		}

		for _, r := range routes {
			name := fmt.Sprintf("%s: %s", tc.route, r)
			t.Run(name, func(t *testing.T) {
				t.Run("GET", func(t *testing.T) {
					res := client.Get(r)

					verify(t, res)
					res.AssertBodyStartsWithString(t, tc.start)
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
}
