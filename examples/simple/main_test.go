package main

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/sktylr/routeit"
)

func TestGetHello(t *testing.T) {
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

func TestHeadHello(t *testing.T) {
	client := routeit.NewTestClient(GetServer())

	res := client.Head("/hello")

	res.AssertBodyNilOrEmpty(t)
	res.AssertStatusCode(t, routeit.StatusOK)
	res.AssertHeaderMatches(t, "Content-Type", "application/json")
}

func TestGetEchoNoQueryParams(t *testing.T) {
	client := routeit.NewTestClient(GetServer())

	res := client.Get("/echo")

	res.AssertStatusCode(t, routeit.StatusOK)
	wantBody := "Looks like you didn't want me to echo anything!\n"
	res.AssertBodyMatchesString(t, wantBody)
	res.AssertHeaderMatches(t, "Content-Length", fmt.Sprintf("%d", len(wantBody)))
	res.AssertHeaderMatches(t, "Content-Type", "text/plain")
}

func TestHeadEchoNoQueryParams(t *testing.T) {
	client := routeit.NewTestClient(GetServer())

	res := client.Head("/echo")

	res.AssertStatusCode(t, routeit.StatusOK)
	res.AssertBodyNilOrEmpty(t)
	bodyLen := len("Looks like you didn't want me to echo anything!\n")
	res.AssertHeaderMatches(t, "Content-Length", fmt.Sprintf("%d", bodyLen))
	res.AssertHeaderMatches(t, "Content-Type", "text/plain")
}

func TestGetEchoWithQueryParam(t *testing.T) {
	client := routeit.NewTestClient(GetServer())

	res := client.Get("/echo?message=hello")

	res.AssertStatusCode(t, routeit.StatusOK)
	wantBody := "Received message to echo: hello\n"
	res.AssertBodyMatchesString(t, wantBody)
	res.AssertHeaderMatches(t, "Content-Length", fmt.Sprintf("%d", len(wantBody)))
	res.AssertHeaderMatches(t, "Content-Type", "text/plain")
}

func TestHeadEchoWithQueryParam(t *testing.T) {
	client := routeit.NewTestClient(GetServer())

	res := client.Head("/echo?message=hello")

	res.AssertStatusCode(t, routeit.StatusOK)
	bodyLen := len("Received message to echo: hello\n")
	res.AssertBodyNilOrEmpty(t)
	res.AssertHeaderMatches(t, "Content-Length", fmt.Sprintf("%d", bodyLen))
	res.AssertHeaderMatches(t, "Content-Type", "text/plain")
}

func TestGetInternalServerError(t *testing.T) {
	tests := []string{
		"/error",
		"/crash",
		"/panic",
	}

	client := routeit.NewTestClient(GetServer())

	for _, tc := range tests {
		t.Run(tc, func(t *testing.T) {
			res := client.Get(tc)

			res.AssertStatusCode(t, routeit.StatusInternalServerError)
			res.AssertBodyMatchesString(t, "500: Internal Server Error")
			res.AssertHeaderMatches(t, "Content-Type", "text/plain")
		})
	}
}

func TestHeadInternalServerError(t *testing.T) {
	tests := []string{
		"/error",
		"/crash",
		"/panic",
	}

	client := routeit.NewTestClient(GetServer())

	for _, tc := range tests {
		t.Run(tc, func(t *testing.T) {
			res := client.Head(tc)

			res.AssertBodyNilOrEmpty(t)
			res.AssertStatusCode(t, routeit.StatusInternalServerError)
			res.AssertHeaderMatches(t, "Content-Type", "text/plain")
			res.AssertHeaderMatches(t, "Content-Length", fmt.Sprintf("%d", len("500: Internal Server Error")))
		})
	}
}

func TestGetRootMethodNotAllowed(t *testing.T) {
	client := routeit.NewTestClient(GetServer())

	res := client.Get("/")

	res.AssertStatusCode(t, routeit.StatusMethodNotAllowed)
	res.AssertHeaderMatches(t, "Allow", "POST")
}

func TestPostRoot(t *testing.T) {
	client := routeit.NewTestClient(GetServer())
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
}

func TestPostRootUnsupportedMediaType(t *testing.T) {
	client := routeit.NewTestClient(GetServer())

	res := client.PostText("/", "this will not be supported")

	res.AssertStatusCode(t, routeit.StatusUnsupportedMediaType)
	res.AssertHeaderMatches(t, "Accept", "application/json")
}

func TestGetMulti(t *testing.T) {
	client := routeit.NewTestClient(GetServer())
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
}

func TestPostMulti(t *testing.T) {
	client := routeit.NewTestClient(GetServer())
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
}
