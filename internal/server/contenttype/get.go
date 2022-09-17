package contenttype

import (
	"net/http"
	"strings"
)

type ContentType int

const (
	ContentTypeUnknown = iota
	ContentTypePlainText
	ContentTypeHTML
	ContentTypeJSON
	ContentTypeXML
	ContentTypeForm
	ContentTypeEventStream
)

func GetContentType(header http.Header) ContentType {
	contentTypeFull := header.Get("Content-Type")
	contentType := strings.TrimSpace(strings.Split(contentTypeFull, ";")[0])

	switch contentType {
	case "text/plain":
		return ContentTypePlainText
	case "text/html", "application/xhtml+xml":
		return ContentTypeHTML
	case "application/json", "text/javascript":
		return ContentTypeJSON
	case "text/xml", "application/xml":
		return ContentTypeXML
	case "application/x-www-form-urlencoded":
		return ContentTypeForm
	case "text/event-stream":
		return ContentTypeEventStream
	default:
		return ContentTypeUnknown
	}
}

func CheckContentType(header http.Header, alowedContentTypes ...ContentType) bool {
	contentType := GetContentType(header)

	for _, alowedContentType := range alowedContentTypes {
		if contentType == alowedContentType {
			return true
		}
	}

	return false
}
