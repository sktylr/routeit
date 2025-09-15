package routeit

import (
	"crypto/tls"
	"testing"
	"time"
)

func TestUpgradeToHttpsMiddleware(t *testing.T) {
	tests := []struct {
		name        string
		opts        TestRequestOptions
		validateRes func(t *testing.T, res *TestResponse)
		wantProceed bool
	}{
		{
			name: "https uses HSTS",
			opts: TestRequestOptions{TlsConnectionState: &tls.ConnectionState{}},
			validateRes: func(t *testing.T, res *TestResponse) {
				val := "max-age=31536000; includeSubdomains"
				res.AssertHeaderMatchesString(t, "Strict-Transport-Security", val)
			},
			wantProceed: true,
		},
		{
			name: "http redirected",
			opts: TestRequestOptions{
				Headers: []string{"Host", "example.com"},
			},
			validateRes: func(t *testing.T, res *TestResponse) {
				res.AssertStatusCode(t, StatusMovedPermanently)
				res.AssertHeaderMatchesString(t, "Location", "https://example.com:443/foo")
			},
		},
	}
	mware := upgradeToHttpsMiddleware(443, 365*24*time.Hour)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := NewTestRequest(t, "/foo", GET, tc.opts)

			res, proceeded, err := TestMiddleware(mware, req)

			if err != nil {
				t.Errorf(`err = %+v, wanted nil`, err)
			}
			if proceeded != tc.wantProceed {
				t.Errorf(`proceeded = %t, wanted %t`, proceeded, tc.wantProceed)
			}
			tc.validateRes(t, res)
		})
	}
}
