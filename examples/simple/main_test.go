package main

import (
	"reflect"
	"testing"

	"github.com/sktylr/routeit"
)

func TestHello(t *testing.T) {
	client := routeit.NewTestClient(GetServer())

	res := client.Get("/hello")

	res.AssertStatusCode(t, routeit.StatusOK)
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
}

func TestEchoNoQueryParams(t *testing.T) {
	client := routeit.NewTestClient(GetServer())

	res := client.Get("/echo")

	res.AssertStatusCode(t, routeit.StatusOK)
	res.AssertBodyMatchesString(t, "Looks like you didn't want me to echo anything!\n")
}

func TestEchoWithQueryParam(t *testing.T) {
	client := routeit.NewTestClient(GetServer())

	res := client.Get("/echo?message=hello")

	res.AssertStatusCode(t, routeit.StatusOK)
	res.AssertBodyMatchesString(t, "Received message to echo: hello\n")
}

func TestError(t *testing.T) {
	client := routeit.NewTestClient(GetServer())

	res := client.Get("/error")

	assertInternalServerError(t, res)
}

func TestCrash(t *testing.T) {
	client := routeit.NewTestClient(GetServer())

	res := client.Get("/crash")

	assertInternalServerError(t, res)
}

func TestPanic(t *testing.T) {
	client := routeit.NewTestClient(GetServer())

	res := client.Get("/panic")

	assertInternalServerError(t, res)
}

func assertInternalServerError(t *testing.T, res *routeit.TestResponse) {
	t.Helper()
	res.AssertStatusCode(t, routeit.StatusInternalServerError)
	res.AssertBodyMatchesString(t, "500: Internal Server Error")
}
