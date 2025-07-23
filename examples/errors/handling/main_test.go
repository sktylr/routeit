package main

import (
	"reflect"
	"testing"
	"time"

	"github.com/sktylr/routeit"
)

var client = routeit.NewTestClient(GetServer())

func TestInternalServerError(t *testing.T) {
	tests := []struct {
		route   string
		wantMsg string
	}{
		{
			route:   "/crash",
			wantMsg: "uh oh we crashed",
		},
		{
			route:   "/panic",
			wantMsg: "oops",
		},
		{
			route:   "/custom-error",
			wantMsg: "this custom error will be mapped to a 500: Internal Server Error",
		},
		{
			route: "/manual-status",
		},
	}

	for _, tc := range tests {
		t.Run(tc.route, func(t *testing.T) {
			res := client.Get(tc.route)

			wantMsg := "An internal error has occurred. We are aware and are investigating. Please try again later or reach out support if it persists."
			if tc.wantMsg != "" {
				wantMsg += " " + tc.wantMsg
			}
			verifyErrorDetailResponse(t, res, routeit.StatusInternalServerError, "internal_server_error", wantMsg)
		})
	}
}

func TestNotFound(t *testing.T) {
	routes := []string{
		"/", "/not-found", "/no-auth-1", "/crash/twice",
	}

	for _, r := range routes {
		wantMsg := "No matching route found for " + r
		t.Run(r, func(t *testing.T) {
			res := client.Get(r)
			verifyErrorDetailResponse(t, res, routeit.StatusNotFound, "not_found", wantMsg)
		})
	}
}

func TestUnauthorised(t *testing.T) {
	tests := []struct {
		name    string
		perform func() *routeit.TestResponse
	}{
		{
			name: "GET",
			perform: func() *routeit.TestResponse {
				return client.Get("/no-auth")
			},
		},
		{
			name: "POST",
			perform: func() *routeit.TestResponse {
				return client.PostText("/no-auth", "no auth :(")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.perform()
			verifyErrorDetailResponse(t, res, routeit.StatusUnauthorized, "unauthorised", "Provide a valid access token")
		})
	}
}

func TestBadRequest(t *testing.T) {
	res := client.Get("/bad-request")

	res.AssertStatusCode(t, routeit.StatusBadRequest)
	res.AssertBodyMatchesString(t, "400: Bad Request")
}

func TestSlow(t *testing.T) {
	res := client.WithTestConfig(routeit.TestConfig{WriteDeadline: 10 * time.Millisecond}).Get("/slow")

	verifyErrorDetailResponse(
		t,
		res,
		routeit.StatusServiceUnavailable,
		"service_unavailable",
		"Our service is currently experiencing issues and is unavailable. Please try again in a few minutes.",
	)
}

func verifyErrorDetailResponse(t *testing.T, res *routeit.TestResponse, wantStatus routeit.HttpStatus, wantCode string, wantMsg string) {
	t.Helper()
	want := ErrorResponse{
		Error: ErrorDetail{
			Message: wantMsg,
			Code:    wantCode,
		},
	}
	res.AssertStatusCode(t, wantStatus)
	var body ErrorResponse
	res.BodyToJson(t, &body)
	if !reflect.DeepEqual(body, want) {
		t.Errorf("body = %+s, wanted %+s", body, want)
	}
}
