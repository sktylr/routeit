package main

import (
	"testing"

	"github.com/sktylr/routeit"
)

func TestGetHelloUnauthorised(t *testing.T) {
	tests := []struct {
		name    string
		headers []string
	}{
		{
			"no header",
			[]string{},
		},
		{
			"invalid authorisation header",
			[]string{"Authorization", "Bearer 123"},
		},
	}
	client := routeit.NewTestClient(GetServer())

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := client.Get("/hello", tc.headers...)

			res.AssertStatusCode(t, routeit.StatusUnauthorized)
		})
	}
}

func TestGetHelloAuthorised(t *testing.T) {
	client := routeit.NewTestClient(GetServer())

	res := client.Get("/hello", "Authorization", "LET ME IN")

	res.AssertStatusCode(t, routeit.StatusOK)
	res.AssertBodyMatchesString(t, "Hello authorised user!")
}

func TestPostHelloAuthorised(t *testing.T) {
	client := routeit.NewTestClient(GetServer())

	res := client.PostText("/hello", "hello!", "Authorization", "LET ME IN")

	res.AssertStatusCode(t, routeit.StatusMethodNotAllowed)
}

func TestPostHelloUnauthorised(t *testing.T) {
	client := routeit.NewTestClient(GetServer())

	res := client.PostText("/hello", "hello!", "Authorization", "Bearer 123")

	// The authorisation is the first point of failure (as opposed to the
	// requested method, which is the second point). For improved security, the
	// first point of failure is returned to the client, ensuring that as
	// little as possible is revealed to the client.
	res.AssertStatusCode(t, routeit.StatusUnauthorized)
}

func TestGetNotFoundAuthorised(t *testing.T) {
	client := routeit.NewTestClient(GetServer())

	res := client.Get("/goodbye", "Authorization", "LET ME IN")

	res.AssertStatusCode(t, routeit.StatusNotFound)
}

func TestGetNotFoundUnauthorised(t *testing.T) {
	client := routeit.NewTestClient(GetServer())

	res := client.Get("/goodbye", "Authorization", "Bearer 123")

	// The authorisation is the first point of failure (as opposed to the
	// resource not being found, which is the the second failure point). For
	// improved security, the first point of failure is returned to the
	// client, ensuring that as little as possible is revealed to the client.
	res.AssertStatusCode(t, routeit.StatusUnauthorized)
}

func TestOptions(t *testing.T) {
	client := routeit.NewTestClient(GetServer())

	t.Run("unauthorised", func(t *testing.T) {
		res := client.Options("/hello")
		res.AssertStatusCode(t, routeit.StatusUnauthorized)
		res.RefuteHeaderPresent(t, "Allow")
	})

	t.Run("authorised not found", func(t *testing.T) {
		res := client.Options("/goodbye", "Authorization", "LET ME IN")
		res.AssertStatusCode(t, routeit.StatusNotFound)
		res.RefuteHeaderPresent(t, "Allow")
	})

	t.Run("authorised", func(t *testing.T) {
		res := client.Options("/hello", "Authorization", "LET ME IN")
		res.AssertStatusCode(t, routeit.StatusNoContent)
		res.AssertBodyMatchesString(t, "")
	})
}
