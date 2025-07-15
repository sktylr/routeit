package routeit

import (
	"fmt"
	"testing"
)

func TestStringNoCharset(t *testing.T) {
	ct := CTApplicationJavaScript

	out := ct.string()
	if out != "application/javascript" {
		t.Errorf(`ct.string() = %#q, wanted "application/javascript"`, out)
	}
}

func TestStringWithCharset(t *testing.T) {
	ct := CTImagePng
	ct.charset = "UTF-8"

	out := ct.string()
	if out != "image/png; charset=UTF-8" {
		t.Errorf(`ct.string() = %#q, wanted "image/png; charset=UTF-8"`, out)
	}
}

func TestEqualsTrue(t *testing.T) {
	tests := []struct {
		a ContentType
		b ContentType
	}{
		{
			a: CTApplicationFormUrlEncoded,
			b: CTApplicationFormUrlEncoded,
		},
		{
			a: CTTextCss,
			b: CTTextCss,
		},
		{
			a: CTTextPlain.WithCharset("UTF-8"),
			b: CTTextPlain,
		},
	}

	for _, tc := range tests {
		name := fmt.Sprintf(`"%s".equals("%s")`, tc.a.string(), tc.b.string())
		t.Run(name, func(t *testing.T) {
			if !tc.a.Equals(tc.b) {
				t.Errorf(`expected equality for a=%#q, b=%#q`, tc.a.string(), tc.b.string())
			}
		})
	}
}

func TestEqualsFalse(t *testing.T) {
	tests := []struct {
		a ContentType
		b ContentType
	}{
		{
			a: CTApplicationFormUrlEncoded,
			b: CTApplicationJson,
		},
		{
			a: CTApplicationJavaScript,
			b: CTTextJavaScript,
		},
		{
			a: CTTextPlain.WithCharset("UTF-8"),
			b: CTTextPlain.WithCharset("UTF-16"),
		},
		{
			a: CTImagePng,
			b: CTImagePng.WithCharset("UTF-16"),
		},
	}

	for _, tc := range tests {
		name := fmt.Sprintf(`"%s".equals("%s")`, tc.a.string(), tc.b.string())
		t.Run(name, func(t *testing.T) {
			if tc.a.Equals(tc.b) {
				t.Errorf(`expected inequality for a=%#q, b=%#q`, tc.a.string(), tc.b.string())
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
			"application/javascript; charset=UTF-16", CTApplicationJavaScript.WithCharset("UTF-16"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.raw, func(t *testing.T) {
			actual := parseContentType(tc.raw)
			if !actual.Equals(tc.want) {
				t.Errorf(`parseContentType(%#q) = %#q, wanted %#q`, tc.raw, actual.string(), tc.want.string())
			}
		})
	}
}
