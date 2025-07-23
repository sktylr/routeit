package main

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/sktylr/routeit"
)

var client = routeit.NewTestClient(GetServer())

func TestHello(t *testing.T) {
	tests := []struct {
		name   string
		accept string
	}{
		{
			name:   "multiple values with one valid",
			accept: "application/xml, application/json",
		},
		{
			name:   "with q parameter favoring unsupported",
			accept: "application/xml;q=0.9, application/json;q=0.1",
		},
		{
			name:   "with wildcard type and subtype",
			accept: "*/*",
		},
		{
			name:   "wildcard with quality on supported",
			accept: "*/*;q=0.1, application/json;q=1.0",
		},
		{
			name:   "valid with extra spaces",
			accept: " application/json ",
		},
		{
			name:   "duplicate valid entries",
			accept: "application/json,application/json",
		},
		{
			name:   "valid with charset parameter",
			accept: "application/json; charset=utf-8",
		},
		{
			name:   "valid followed by malformed",
			accept: "application/json, ?",
		},
	}
	verify := func(t *testing.T, res *routeit.TestResponse) {
		res.AssertStatusCode(t, routeit.StatusOK)
		res.AssertHeaderMatches(t, "Content-Type", "application/json")
		res.AssertHeaderMatches(t, "Content-Length", "53")
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("GET", func(t *testing.T) {
				res := client.Get("/hello", "Accept", tc.accept)

				verify(t, res)
				var body Example
				res.BodyToJson(t, &body)
				want := Example{
					Name: "John Doe",
					Nested: Nested{
						Age:    25,
						Height: 1.82,
					},
				}
				if !reflect.DeepEqual(body, want) {
					t.Errorf(`Json response = %#v, wanted %#v`, body, want)
				}
			})

			t.Run("HEAD", func(t *testing.T) {
				res := client.Head("/hello", "Accept", tc.accept)

				verify(t, res)
				res.AssertBodyEmpty(t)
			})
		})
	}
}

func TestEcho(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		wantBody string
	}{
		{
			"no query params",
			"",
			"Looks like you didn't want me to echo anything!\n",
		},
		{
			"simple query param",
			"?message=Hello",
			"Received message to echo: Hello\n",
		},
		{
			"escaped query param",
			"?message=%22Hello%20there%21%20This%20is%20escaped%22",
			"Received message to echo: \"Hello there! This is escaped\"\n",
		},
	}
	verify := func(t *testing.T, res *routeit.TestResponse, wantLen int) {
		res.AssertStatusCode(t, routeit.StatusOK)
		res.AssertHeaderMatches(t, "Content-Length", fmt.Sprintf("%d", wantLen))
		res.AssertHeaderMatches(t, "Content-Type", "text/plain")
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			endpoint := fmt.Sprintf("/echo%s", tc.query)

			t.Run("GET", func(t *testing.T) {
				res := client.Get(endpoint)

				verify(t, res, len(tc.wantBody))
				res.AssertBodyMatchesString(t, tc.wantBody)
			})

			t.Run("HEAD", func(t *testing.T) {
				res := client.Head(endpoint)

				verify(t, res, len(tc.wantBody))
				res.AssertBodyEmpty(t)
			})
		})
	}
}

func TestInternalServerError(t *testing.T) {
	tests := []string{
		"/error",
		"/crash",
		"/panic",
		"/bad-status",
	}
	verify := func(t *testing.T, res *routeit.TestResponse) {
		res.AssertStatusCode(t, routeit.StatusInternalServerError)
		res.AssertHeaderMatches(t, "Content-Length", "26")
		res.AssertHeaderMatches(t, "Content-Type", "text/plain")
	}

	for _, tc := range tests {
		t.Run(tc, func(t *testing.T) {
			t.Run("GET", func(t *testing.T) {
				res := client.Get(tc)

				verify(t, res)
				res.AssertBodyMatchesString(t, "500: Internal Server Error")
			})

			t.Run("HEAD", func(t *testing.T) {
				res := client.Head(tc)

				verify(t, res)
				res.AssertBodyEmpty(t)
			})
		})
	}
}

func TestHostValidation(t *testing.T) {
	t.Run("rejected", func(t *testing.T) {
		hosts := []string{
			"sub.web.example.com",
			"127.0.1.1",
			"127.0.0.1",
			"[::2]",
			"example.com.web",
			"dev.localhost:3000A",
		}

		for _, host := range hosts {
			t.Run(host, func(t *testing.T) {
				res := client.Get("/hello", "Host", host)
				res.AssertStatusCode(t, routeit.StatusBadRequest)
				res.AssertBodyMatchesString(t, "400: Bad Request")
			})
		}
	})

	t.Run("accepted", func(t *testing.T) {
		hosts := []string{
			"localhost",
			"localhost:8080",
			"localhost:1234",
			"dev.localhost",
			"dev.localhost:3000",
			"[::1]",
			"[::1]:8000",
			"example.com",
			"web.example.com",
			"www.example.com",
			"example.com:443",
			"api.example.com:8080",
		}

		for _, host := range hosts {
			t.Run(host, func(t *testing.T) {
				res := client.Get("/hello", "Host", host)
				res.AssertStatusCode(t, routeit.StatusOK)
				res.AssertBodyContainsString(t, `"name":"John Doe"`)
			})
		}
	})
}

func TestRoot(t *testing.T) {
	t.Run("POST", func(t *testing.T) {
		t.Run("application/json", func(t *testing.T) {
			inBody := Example{
				Name: "Foo Bar",
				Nested: Nested{
					Age:    34,
					Height: 1.89,
				},
			}
			wantBody := Greeting{
				From: inBody,
				To: Example{
					Name: "Jane Doe",
					Nested: Nested{
						Age:    29,
						Height: 1.62,
					},
				},
			}

			res := client.PostJson("/", inBody)

			res.AssertStatusCode(t, routeit.StatusCreated)
			var actual Greeting
			res.BodyToJson(t, &actual)
			if !reflect.DeepEqual(actual, wantBody) {
				t.Errorf(`Json response = %#v, wanted %#v`, actual, wantBody)
			}
		})

		t.Run("unsupported media type", func(t *testing.T) {
			res := client.PostText("/", "this will not be supported")

			res.AssertStatusCode(t, routeit.StatusUnsupportedMediaType)
			res.AssertHeaderMatches(t, "Accept", "application/json")
		})
	})

	t.Run("GET - not allowed", func(t *testing.T) {
		res := client.Get("/")

		res.AssertStatusCode(t, routeit.StatusMethodNotAllowed)
		res.AssertHeaderMatches(t, "Allow", "POST, OPTIONS")
	})
}

func TestMulti(t *testing.T) {
	t.Run("GET", func(t *testing.T) {
		wantBody := Example{
			Name: "From GET",
			Nested: Nested{
				Age:    100,
				Height: 2.0,
			},
		}

		res := client.Get("/multi")

		res.AssertStatusCode(t, routeit.StatusOK)
		var body Example
		res.BodyToJson(t, &body)
		if !reflect.DeepEqual(body, wantBody) {
			t.Errorf(`Json response = %#v, wanted %#v`, body, wantBody)
		}
	})

	t.Run("POST", func(t *testing.T) {
		inBody := Nested{
			Age:    25,
			Height: 1.95,
		}
		wantBody := Example{
			Name:   "From POST",
			Nested: inBody,
		}

		res := client.PostJson("/multi", inBody)

		res.AssertStatusCode(t, routeit.StatusCreated)
		var body Example
		res.BodyToJson(t, &body)
		if !reflect.DeepEqual(body, wantBody) {
			t.Errorf(`Json response = %#v, wanted %#v`, body, wantBody)
		}
	})
}

func TestModify(t *testing.T) {
	t.Run("accepts PUT with text/plain", func(t *testing.T) {
		res := client.PutText("/modify", "Hello!")

		res.AssertStatusCode(t, routeit.StatusOK)
		res.AssertBodyMatchesString(t, "Hello!")
	})

	t.Run("rejects PUT with application/json", func(t *testing.T) {
		in := Nested{
			Age: 28,
		}
		res := client.PutJson("/modify", in)

		res.AssertStatusCode(t, routeit.StatusUnsupportedMediaType)
		res.AssertHeaderMatches(t, "Accept", "text/plain")
	})

	t.Run("rejects POST", func(t *testing.T) {
		res := client.PostText("/modify", "Hello!")

		res.AssertStatusCode(t, routeit.StatusMethodNotAllowed)
		res.AssertHeaderMatches(t, "Allow", "PUT, OPTIONS")
	})

	t.Run("OPTIONS", func(t *testing.T) {
		res := client.Options("/modify")

		res.AssertStatusCode(t, routeit.StatusNoContent)
		res.AssertHeaderMatches(t, "Allow", "PUT, OPTIONS")
		res.AssertBodyEmpty(t)
	})
}

func TestGlobalOptions(t *testing.T) {
	res := client.Options("*")

	res.AssertStatusCode(t, routeit.StatusNoContent)
	res.AssertBodyEmpty(t)
	res.AssertHeaderMatches(t, "Allow", "GET, HEAD, POST, PUT, DELETE, PATCH, OPTIONS")
}

func TestDelete(t *testing.T) {
	t.Run("DELETE", func(t *testing.T) {
		res := client.Delete("/delete")

		res.AssertStatusCode(t, routeit.StatusNoContent)
		res.AssertBodyEmpty(t)
	})

	t.Run("OPTIONS", func(t *testing.T) {
		res := client.Options("/delete")

		res.AssertStatusCode(t, routeit.StatusNoContent)
		res.AssertBodyEmpty(t)
		res.AssertHeaderMatches(t, "Allow", "DELETE, OPTIONS")
	})

	t.Run("GET", func(t *testing.T) {
		res := client.Get("/delete")

		res.AssertStatusCode(t, routeit.StatusMethodNotAllowed)
		res.AssertHeaderMatches(t, "Allow", "DELETE, OPTIONS")
	})
}

func TestUpdate(t *testing.T) {
	t.Run("PATCH", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			res := client.PatchText("/update?conflict=false", "")
			res.AssertStatusCode(t, routeit.StatusOK)
			res.AssertBodyMatchesString(t, "Resource updated successfully\n")
		})

		t.Run("failure", func(t *testing.T) {
			tests := []struct {
				q    string
				want routeit.HttpStatus
			}{
				{
					"",
					routeit.StatusUnprocessableContent,
				},
				{
					"conflict=true",
					routeit.StatusConflict,
				},
			}

			for _, tc := range tests {
				t.Run(fmt.Sprintf("query = %#q", tc.q), func(t *testing.T) {
					res := client.PatchText("/update?"+tc.q, "")
					res.AssertStatusCode(t, tc.want)
				})
			}
		})
	})

	t.Run("OPTIONS", func(t *testing.T) {
		res := client.Options("/update")

		res.AssertStatusCode(t, routeit.StatusNoContent)
		res.AssertBodyEmpty(t)
		res.AssertHeaderMatches(t, "Allow", "PATCH, OPTIONS")
	})

	t.Run("GET", func(t *testing.T) {
		res := client.Get("/update")

		res.AssertStatusCode(t, routeit.StatusMethodNotAllowed)
		res.AssertHeaderMatches(t, "Allow", "PATCH, OPTIONS")
	})
}

func TestURIValidation(t *testing.T) {
	t.Run("strips single trailing slash", func(t *testing.T) {
		res := client.Get("/hello/")
		res.AssertStatusCode(t, routeit.StatusOK)
		res.AssertBodyContainsString(t, `"name":"John Doe",`)
	})

	t.Run("rejects URIs longer than 8KiB", func(t *testing.T) {
		uri := "/" + strings.Repeat("a", 8192)
		res := client.Get(uri)
		res.AssertStatusCode(t, routeit.StatusURITooLong)
	})

	t.Run("accepts URIs at 8KiB", func(t *testing.T) {
		uri := "/" + strings.Repeat("a", 8191)
		res := client.Get(uri)
		res.AssertStatusCode(t, routeit.StatusNotFound)
	})
}

func TestClientAcceptHeaderValidation(t *testing.T) {
	tests := []struct {
		name   string
		accept string
	}{
		{
			name:   "explicitly empty",
			accept: "",
		},
		{
			name:   "valid part but invalid subtype",
			accept: "application/graphql",
		},
		{
			name:   "valid subtype but invalid part",
			accept: "foo/json",
		},
		{
			name:   "wildcard subtype but invalid part",
			accept: "text/*",
		},
		{
			name:   "multiple values with none valid",
			accept: "application/xml, text/html",
		},
		{
			name:   "with q=0 for supported type",
			accept: "application/json;q=0, application/xml",
		},
		{
			name:   "wildcard with low quality for all",
			accept: "*/*;q=0",
		},
		{
			name:   "empty string with spaces",
			accept: "   ",
		},
		{
			name:   "only commas",
			accept: ",,,",
		},
		{
			name:   "only * part, no subtype",
			accept: "*",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := client.Get("/hello", "Accept", tc.accept)

			res.AssertStatusCode(t, routeit.StatusNotAcceptable)
			res.AssertBodyEmpty(t)
			res.RefuteHeaderPresent(t, "Content-Type")
			res.RefuteHeaderPresent(t, "Content-Length")
		})
	}
}
