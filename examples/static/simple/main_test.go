package main

import (
	"fmt"
	"testing"

	"github.com/sktylr/routeit"
)

var client = routeit.NewTestClient(GetServer())

func TestServer(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		tests := []struct {
			endpoint     string
			wantCT       string
			wantCLen     int
			wantStart    string
			wantContains []string
		}{
			{
				endpoint:     "/static/about.html",
				wantCT:       "text/html; charset=utf-8",
				wantCLen:     650,
				wantStart:    "<!DOCTYPE html>",
				wantContains: []string{"<title>About - My Server</title>", "<h1>About This Server</h1>"},
			},
			{
				endpoint:  "/static/gopher.png",
				wantCT:    "image/png",
				wantCLen:  70548,
				wantStart: "\x89PNG",
			},
			{
				endpoint:  "/static/hello.txt",
				wantCT:    "text/plain; charset=utf-8",
				wantCLen:  13,
				wantStart: "Hello World!\n",
			},
			{
				endpoint:     "/static/index.html",
				wantCT:       "text/html; charset=utf-8",
				wantCLen:     579,
				wantStart:    "<!DOCTYPE html>",
				wantContains: []string{"<title>Home - My Server</title>", "<h1>Welcome to the Home Page</h1>"},
			},
			{
				endpoint:  "/static/styles.css",
				wantCT:    "text/css; charset=utf-8",
				wantCLen:  868,
				wantStart: "body {\n",
				wantContains: []string{
					`font-family: "Segoe UI", Tahoma, Geneva, Verdana, sans-serif;`,
					"box-shadow: 0 6px 16px rgba(0, 0, 0, 0.15);\n",
				},
			},
		}

		for _, tc := range tests {
			t.Run(tc.endpoint, func(t *testing.T) {
				verifyMeta := func(t *testing.T, res *routeit.TestResponse) {
					t.Helper()
					res.AssertStatusCode(t, routeit.StatusOK)
					res.AssertHeaderMatchesString(t, "Content-Type", tc.wantCT)
					res.AssertHeaderMatchesString(t, "Content-Length", fmt.Sprintf("%d", tc.wantCLen))
				}

				t.Run("GET", func(t *testing.T) {
					res := client.Get(tc.endpoint)

					verifyMeta(t, res)
					res.AssertBodyStartsWithString(t, tc.wantStart)
					for _, want := range tc.wantContains {
						res.AssertBodyContainsString(t, want)
					}
				})

				t.Run("HEAD", func(t *testing.T) {
					res := client.Head(tc.endpoint)

					verifyMeta(t, res)
					res.AssertBodyEmpty(t)
				})
			})
		}
	})

	t.Run("not found", func(t *testing.T) {
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
		verify := func(t *testing.T, res *routeit.TestResponse) {
			t.Helper()
			res.AssertStatusCode(t, routeit.StatusNotFound)
			res.AssertHeaderMatchesString(t, "Content-Type", "text/plain")
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
					res.AssertBodyEmpty(t)
				})
			})
		}
	})
}
