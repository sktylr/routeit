package main

import (
	"fmt"
	"testing"

	"github.com/sktylr/routeit"
)

func TestFrontend(t *testing.T) {
	tests := []struct {
		endpoints    []string
		wantCT       string
		wantCLen     int
		wantStart    string
		wantContains []string
	}{
		{
			endpoints:    []string{"/static/html/login.html", "/login"},
			wantCT:       "text/html; charset=utf-8",
			wantCLen:     1922,
			wantStart:    "<!DOCTYPE html>",
			wantContains: []string{"<title>Login</title>", "Don't have an account?\n      <a href=\"/register\">Register here</a>"},
		},
		{
			endpoints:    []string{"/static/html/register.html", "/register"},
			wantCT:       "text/html; charset=utf-8",
			wantCLen:     2202,
			wantStart:    "<!DOCTYPE html>",
			wantContains: []string{"<title>Register</title>", "Already have an account?\n      <a href=\"/login\">Login here</a>"},
		},
		{
			endpoints:    []string{"/static/html/index.html", "/"},
			wantCT:       "text/html; charset=utf-8",
			wantCLen:     5721,
			wantStart:    "<!DOCTYPE html>",
			wantContains: []string{"<title>My TODO Lists</title>", `<button id="prevPage" disabled>Previous</button>`},
		},
		{
			endpoints:    []string{"/static/html/list.html", "/lists/foo"},
			wantCT:       "text/html; charset=utf-8",
			wantCLen:     6343,
			wantStart:    "<!DOCTYPE html>",
			wantContains: []string{"<title>TODO List</title>", `<button id="prevPage" disabled>Previous</button>`},
		},
		{
			endpoints: []string{"/static/styles/base.css", "/css/base.css"},
			wantCT:    "text/css; charset=utf-8",
			wantCLen:  993,
			wantStart: "* {\n  box-sizing: border-box;",
			wantContains: []string{
				`font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;`,
				"a {\n  color: #007bff;\n  text-decoration: none;\n}",
			},
		},
		{
			endpoints: []string{"/static/styles/auth.css", "/css/auth.css"},
			wantCT:    "text/css; charset=utf-8",
			wantCLen:  1095,
			wantStart: "form {\n  background-color: white;",
			wantContains: []string{
				"input:focus {\n  border-color: #007bff;\n  outline: none;\n  background: #f0f8ff;\n}",
			},
		},
		{
			endpoints: []string{"/static/styles/index.css", "/css/index.css"},
			wantCT:    "text/css; charset=utf-8",
			wantCLen:  3314,
			wantStart: ".container {\n  max-width: 1200px;",
			wantContains: []string{
				".list-card h2 {\n  margin-bottom: 0.5rem;\n  font-size: 1.25rem;\n}",
			},
		},
		{
			endpoints: []string{"/static/styles/list.css", "/css/list.css"},
			wantCT:    "text/css; charset=utf-8",
			wantCLen:  3377,
			wantStart: ".list-details {\n  margin-bottom: 2rem;\n  text-align: center;\n}",
			wantContains: []string{
				".list-details .meta {\n  font-size: 0.85rem;\n  color: #777;\n}",
			},
		},
		{
			endpoints:    []string{"/auth.js", "/static/js/auth.js"},
			wantCT:       "text/javascript; charset=utf-8",
			wantCLen:     2903,
			wantStart:    `import { API_BASE } from "./api.js";`,
			wantContains: []string{"export async function login(email, password) {"},
		},
		{
			endpoints:    []string{"/api.js", "/static/js/api.js"},
			wantCT:       "text/javascript; charset=utf-8",
			wantCLen:     5064,
			wantStart:    `import { refreshToken } from "./auth.js"`,
			wantContains: []string{"export async function getLists("},
		},
	}
	client := routeit.NewTestClient(GetFrontendServer())

	for _, tc := range tests {
		for _, ep := range tc.endpoints {
			t.Run(ep, func(t *testing.T) {
				verifyMeta := func(t *testing.T, res *routeit.TestResponse) {
					t.Helper()
					res.AssertStatusCode(t, routeit.StatusOK)
					res.AssertHeaderMatchesString(t, "Content-Type", tc.wantCT)
					res.AssertHeaderMatchesString(t, "Content-Length", fmt.Sprintf("%d", tc.wantCLen))
				}

				t.Run("GET", func(t *testing.T) {
					res := client.Get(ep)

					verifyMeta(t, res)
					res.AssertBodyStartsWithString(t, tc.wantStart)
					for _, want := range tc.wantContains {
						res.AssertBodyContainsString(t, want)
					}
				})

				t.Run("HEAD", func(t *testing.T) {
					res := client.Head(ep)

					verifyMeta(t, res)
					res.AssertBodyEmpty(t)
				})
			})
		}
	}
}
