package main

import (
	"crypto/tls"
	"reflect"
	"testing"

	"github.com/sktylr/routeit"
)

func TestServer(t *testing.T) {
	tests := []struct {
		name     string
		tlsState *tls.ConnectionState
		wantRes  InfoResponse
	}{
		{
			name: "valid",
			tlsState: &tls.ConnectionState{
				Version:     tls.VersionTLS13,
				ServerName:  "foo-bar-server",
				CipherSuite: tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
			},
			wantRes: InfoResponse{
				Version:     "TLS 1.3",
				ServerName:  "foo-bar-server",
				CipherSuite: "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA",
			},
		},
		{
			name: "fallback",
			tlsState: &tls.ConnectionState{
				Version:     1234,
				ServerName:  "foo-bar-server",
				CipherSuite: 5678,
			},
			wantRes: InfoResponse{
				Version:     "0x04D2",
				ServerName:  "foo-bar-server",
				CipherSuite: "0x162E",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := routeit.NewTestTlsClient(GetServer(), tc.tlsState)

			res := client.Get("/info")

			res.AssertStatusCode(t, routeit.StatusOK)
			var body InfoResponse
			res.BodyToJson(t, &body)
			if !reflect.DeepEqual(body, tc.wantRes) {
				t.Errorf(`mismatching body, got = %+v, want = %+v`, body, tc.wantRes)
			}
		})
	}
}
