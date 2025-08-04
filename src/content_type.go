package routeit

import (
	"strconv"
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

	CTAcceptAll = ContentType{part: "*", subtype: "*"}
)

type ContentType struct {
	part    string
	subtype string
	charset string
	q       float32
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
		if len(kvp) != 2 {
			continue
		}
		switch kvp[0] {
		case "charset":
			ct = ct.WithCharset(kvp[1])
		case "q":
			q, err := strconv.ParseFloat(kvp[1], 32)
			if err == nil {
				if q <= 0.0 {
					// Golang's 0 value for a float32 is 0, so we explicitly
					// set this to a negative number to differentiate between
					// unset and 0 values.
					ct.q = -1
				} else {
					ct.q = float32(q)
				}
			}
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
func (a ContentType) Matches(b ContentType) bool {
	if a.q < 0 {
		return false
	}
	samePart := a.part == "*" || a.part == b.part
	sameSubtype := a.subtype == "*" || a.subtype == b.subtype
	if !(samePart && sameSubtype) {
		return false
	}
	if (a.charset == "" || a.charset == "utf-8") && (b.charset == "" || b.charset == "utf-8") {
		return true
	}
	return a.charset == b.charset
}

func (ct ContentType) isValid() bool {
	return ct.part != "" && ct.subtype != ""
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

// TODO: this will need to handle the multiple possible `Accept` headers
func parseAcceptHeader(h headers) []ContentType {
	var accept []ContentType
	acceptRaw, hasAccept := h.Get("Accept")
	if !hasAccept {
		// The Accept header is not required to be sent by the client, so we
		// take a lenient approach and replace it with acceptance of all
		// content types.
		accept = []ContentType{CTAcceptAll}
	} else {
		for acc := range strings.SplitSeq(acceptRaw, ",") {
			ct := parseContentType(strings.TrimSpace(acc))
			if ct.isValid() {
				accept = append(accept, ct)
			}
		}
	}
	return accept
}
