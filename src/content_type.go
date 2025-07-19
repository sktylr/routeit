package routeit

import (
	"strings"
)

var (
	CTApplicationFormUrlEncoded = ContentType{part: "application", subtype: "x-www-form-urlencoded"}
	CTApplicationGraphQL        = ContentType{part: "application", subtype: "graphql"}
	CTApplicationJavaScript     = ContentType{part: "application", subtype: "javascript"}
	CTApplicationJson           = ContentType{part: "application", subtype: "json"}
	CTApplicationOctetStream    = ContentType{part: "application", subtype: "octet-stream"}
	CTApplicationPdf            = ContentType{part: "application", subtype: "pdf"}
	CTApplicationXml            = ContentType{part: "application", subtype: "xml"}
	CTApplicationZip            = ContentType{part: "application", subtype: "zip"}

	CTTextCss        = ContentType{part: "text", subtype: "css"}
	CTTextCsv        = ContentType{part: "text", subtype: "csv"}
	CTTextHtml       = ContentType{part: "text", subtype: "html"}
	CTTextJavaScript = ContentType{part: "text", subtype: "javascript"}
	CTTextMarkdown   = ContentType{part: "text", subtype: "markdown"}
	CTTextPlain      = ContentType{part: "text", subtype: "plain"}

	CTImageAvif = ContentType{part: "image", subtype: "avif"}
	CTImageGif  = ContentType{part: "image", subtype: "gif"}
	CTImageJpeg = ContentType{part: "image", subtype: "jpeg"}
	CTImagePng  = ContentType{part: "image", subtype: "png"}
	CTImageSvg  = ContentType{part: "image", subtype: "svg+xml"}
	CTImageWebp = ContentType{part: "image", subtype: "webp"}

	CTAudioMpeg = ContentType{part: "audio", subtype: "mpeg"}
	CTAudioOgg  = ContentType{part: "audio", subtype: "ogg"}
	CTAudioWav  = ContentType{part: "audio", subtype: "wav"}

	CTVideoMp4  = ContentType{part: "video", subtype: "mp4"}
	CTVideoOgg  = ContentType{part: "video", subtype: "ogg"}
	CTVideoWebm = ContentType{part: "video", subtype: "webm"}

	CTMultipartByteranges = ContentType{part: "multipart", subtype: "byteranges"}
	CTMultipartFormData   = ContentType{part: "multipart", subtype: "form-data"}
)

type ContentType struct {
	part    string
	subtype string
	charset string
}

func parseContentType(raw string) ContentType {
	if raw == "" {
		return ContentType{}
	}

	splitParams := strings.Split(raw, ";")
	partSubtype := strings.SplitN(splitParams[0], "/", 2)
	if len(partSubtype) != 2 {
		// The content type is unknown due to not having a part/subtype split
		return ContentType{}
	}

	ct := ContentType{
		part:    partSubtype[0],
		subtype: partSubtype[1],
	}

	for _, param := range splitParams[1:] {
		kvp := strings.Split(strings.TrimSpace(param), "=")
		if len(kvp) == 2 && kvp[0] == "charset" {
			return ct.WithCharset(kvp[1])
		}
	}

	return ct
}

// Destructively sets the charset of the content type
func (ct ContentType) WithCharset(cs string) ContentType {
	ct.charset = strings.ToLower(cs)
	return ct
}

// Compare two content types for equality. For ease, this considers two content
// types to be equal of they share the same part and subtype, and their charset
// is the same OR one charset is UTF-8 and the other is unset. This is because
// UTF-8 is the default charset used by routeit but is sometimes omitted due to
// being the de-facto standard across the web.
func (a ContentType) Equals(b ContentType) bool {
	if a.part != b.part || a.subtype != b.subtype {
		return false
	}
	if (a.charset == "" || a.charset == "utf-8") && (b.charset == "" || b.charset == "utf-8") {
		return true
	}
	return a.charset == b.charset
}

func (ct ContentType) string() string {
	var sb strings.Builder
	sb.WriteString(ct.part)
	sb.WriteRune('/')
	sb.WriteString(ct.subtype)
	if ct.charset != "" {
		sb.WriteString("; charset=")
		sb.WriteString(ct.charset)
	}
	return sb.String()
}
