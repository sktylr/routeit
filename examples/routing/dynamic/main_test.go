package main

import (
	"reflect"
	"testing"

	"github.com/sktylr/routeit"
)

func TestHello(t *testing.T) {
	client := routeit.NewTestClient(GetServer())

	t.Run("happy path", func(t *testing.T) {
		tests := []struct {
			name string
			path string
			want OutGreeting
		}{
			{
				"dynamic route",
				"/hello/username",
				OutGreeting{
					To:      "username",
					From:    "routeit dynamic route",
					Message: "Hello from the test!",
				},
			},
			{
				"static route",
				"/hello/bob",
				OutGreeting{
					To:      "bob",
					From:    "routeit static route",
					Message: "Hello from the test!",
				},
			},
			{
				"custom greeting",
				"/welcome/bob",
				OutGreeting{
					To:      "bob",
					From:    "routeit custom greeting route",
					Message: "welcome",
				},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				in := InGreeting{"Hello from the test!"}

				res := client.PostJson(tc.path, in)

				res.AssertStatusCode(t, routeit.StatusCreated)
				var out OutGreeting
				res.BodyToJson(t, &out)
				if !reflect.DeepEqual(out, tc.want) {
					t.Errorf(`body = %#q, wanted %#q`, out, tc.want)
				}
			})
		}
	})

	t.Run("not found", func(t *testing.T) {
		tests := []string{
			"/hello/",
			"/hello",
			"/hello/name/bob",
		}
		inBody := InGreeting{"Hello from the test!"}

		for _, tc := range tests {
			t.Run(tc, func(t *testing.T) {
				res := client.PostJson(tc, inBody)
				res.AssertStatusCode(t, routeit.StatusNotFound)
			})
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		tests := []string{
			"/hello/bob",
			"/greetings/bob",
			"/hello/username",
		}

		for _, tc := range tests {
			t.Run(tc, func(t *testing.T) {
				res := client.Get(tc)
				res.AssertStatusCode(t, routeit.StatusMethodNotAllowed)
				res.AssertHeaderMatches(t, "Allow", "POST, OPTIONS")
			})
		}
	})

	t.Run("bad request body content", func(t *testing.T) {
		tests := []string{
			"/hello/bob",
			"/hello/username",
			"/hello/foo",
		}

		for _, tc := range tests {
			t.Run(tc, func(t *testing.T) {
				res := client.PostText(tc, "this is not going to be accepted")
				res.AssertStatusCode(t, routeit.StatusUnsupportedMediaType)
				res.AssertHeaderMatches(t, "Accept", "application/json")
			})
		}
	})
}
