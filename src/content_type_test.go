package routeit

import (
	"fmt"
	"testing"
)

func TestContentTypeString(t *testing.T) {
	tests := []struct {
		name string
		in   ContentType
		want string
	}{
		{
			name: "no charset",
			in:   CTApplicationJavaScript,
			want: "application/javascript",
		},
		{
			name: "with charset (upper case)",
			in:   CTImagePng.WithCharset("UTF-8"),
			want: "image/png; charset=utf-8",
		},
		{
			name: "with charset (lower case)",
			in:   CTImagePng.WithCharset("utf-8"),
			want: "image/png; charset=utf-8",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out := tc.in.string()
			if out != tc.want {
				t.Errorf(`ContentType.string() = %#q, wanted %#q`, out, tc.want)
			}
		})
	}
}

func TestContentTypeMatches(t *testing.T) {
	tests := []struct {
		a    ContentType
		b    ContentType
		want bool
	}{
		{
			a:    CTApplicationFormUrlEncoded,
			b:    CTApplicationFormUrlEncoded,
			want: true,
		},
		{
			a:    CTTextCss,
			b:    CTTextCss,
			want: true,
		},
		{
			a:    CTTextPlain.WithCharset("UTF-8"),
			b:    CTTextPlain,
			want: true,
		},
		{
			a:    CTTextPlain.WithCharset("utf-8"),
			b:    CTTextPlain,
			want: true,
		},
		{
			a:    CTTextPlain.WithCharset("utf-8"),
			b:    CTTextPlain.WithCharset("UTF-8"),
			want: true,
		},

		{
			a:    CTApplicationFormUrlEncoded,
			b:    CTApplicationJson,
			want: false,
		},
		{
			a:    CTApplicationJavaScript,
			b:    CTTextJavaScript,
			want: false,
		},
		{
			a:    CTTextPlain.WithCharset("UTF-8"),
			b:    CTTextPlain.WithCharset("UTF-16"),
			want: false,
		},
		{
			a:    CTImagePng,
			b:    CTImagePng.WithCharset("UTF-16"),
			want: false,
		},
		{
			a:    CTAcceptAll,
			b:    CTApplicationFormUrlEncoded,
			want: true,
		},
		{
			a:    CTAcceptAll,
			b:    CTAcceptAll,
			want: true,
		},
		{
			a:    ContentType{part: "application", subtype: "*"},
			b:    CTApplicationGraphQL,
			want: true,
		},
		{
			a:    ContentType{part: "application", subtype: "*"},
			b:    CTTextJavaScript,
			want: false,
		},
	}

	for _, tc := range tests {
		name := fmt.Sprintf(`"%s".equals("%s")`, tc.a.string(), tc.b.string())
		t.Run(name, func(t *testing.T) {
			if tc.a.Matches(tc.b) != tc.want {
				t.Errorf(`%#q.equals(%#q) = %t, wanted %t`, tc.a.string(), tc.b.string(), tc.a.Matches(tc.b), tc.want)
			}
		})
	}
}

func TestParseContentType(t *testing.T) {
	tests := []struct {
		raw  string
		want ContentType
	}{
		{
			"application/json", CTApplicationJson,
		},
		{
			"unknown", ContentType{},
		},
		{
			"", ContentType{},
		},
		{
			"application/javascript; charset=UTF-16", CTApplicationJavaScript.WithCharset("utf-16"),
		},
		{
			"image/png; charset=utf-16", CTImagePng.WithCharset("utf-16"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.raw, func(t *testing.T) {
			actual := parseContentType(tc.raw)
			if !actual.Matches(tc.want) {
				t.Errorf(`parseContentType(%#q) = %#q, wanted %#q`, tc.raw, actual.string(), tc.want.string())
			}
		})
	}
}
