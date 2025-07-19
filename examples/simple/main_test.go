package main

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/sktylr/routeit"
)

func TestHello(t *testing.T) {
	client := routeit.NewTestClient(GetServer())
	verify := func(t *testing.T, res *routeit.TestResponse) {
		res.AssertStatusCode(t, routeit.StatusOK)
		res.AssertHeaderMatches(t, "Content-Type", "application/json")
		res.AssertHeaderMatches(t, "Content-Length", "53")
	}

	t.Run("GET", func(t *testing.T) {
		res := client.Get("/hello")

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
		res := client.Head("/hello")

		verify(t, res)
		res.AssertBodyEmpty(t)
	})
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
	client := routeit.NewTestClient(GetServer())

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
	client := routeit.NewTestClient(GetServer())

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

func TestRoot(t *testing.T) {
	client := routeit.NewTestClient(GetServer())

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
	client := routeit.NewTestClient(GetServer())

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
	client := routeit.NewTestClient(GetServer())

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
	client := routeit.NewTestClient(GetServer())

	res := client.Options("*")

	res.AssertStatusCode(t, routeit.StatusNoContent)
	res.AssertBodyEmpty(t)
	res.AssertHeaderMatches(t, "Allow", "GET, HEAD, POST, PUT, OPTIONS")
}
