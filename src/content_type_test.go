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
		{
			a:    ContentType{part: "application", subtype: "json", q: -1},
			b:    CTApplicationJson,
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
		{
			"image/png;q=0.0", ContentType{part: "image", subtype: "png", q: -1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.raw, func(t *testing.T) {
			actual := parseContentType(tc.raw)
			if actual != tc.want {
				t.Errorf(`parseContentType(%#q) = %#q, wanted %#q`, tc.raw, actual.string(), tc.want.string())
			}
		})
	}
}

func TestParseAcceptHeader(t *testing.T) {
	tests := []struct {
		name string
		in   map[string]string
		want []ContentType
	}{
		{
			name: "header not present",
			in:   map[string]string{},
			want: []ContentType{CTAcceptAll},
		},
		{
			name: "header present but empty",
			in:   map[string]string{"accept": ""},
			want: []ContentType{},
		},
		{
			name: "header present single element",
			in:   map[string]string{"accept": "application/json"},
			want: []ContentType{CTApplicationJson},
		},
		{
			name: "header present single element whitespaced",
			in:   map[string]string{"accept": "  application/json\t"},
			want: []ContentType{CTApplicationJson},
		},
		{
			name: "header present single element with weight",
			in:   map[string]string{"accept": "application/json;q=0.9"},
			want: []ContentType{{part: "application", subtype: "json", q: 0.9}},
		},
		{
			name: "header present multiple elements",
			in:   map[string]string{"accept": "application/json,text/html"},
			want: []ContentType{CTApplicationJson, CTTextHtml},
		},
		{
			name: "header present multiple elements with weight",
			in:   map[string]string{"accept": "application/json;q=0.9,text/html;q=0.8"},
			want: []ContentType{
				{part: "application", subtype: "json", q: 0.9},
				{part: "text", subtype: "html", q: 0.8},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := headers{}
			for k, v := range tc.in {
				h.Set(k, v)
			}

			accept := parseAcceptHeader(h)

			if len(accept) != len(tc.want) {
				t.Errorf(`length = %d, wanted %d`, len(accept), len(tc.want))
			}
			for i, ct := range accept {
				if ct != tc.want[i] {
					t.Errorf(`got = %v, wanted %v`, ct, tc.want[i])
				}
			}
		})
	}
}
