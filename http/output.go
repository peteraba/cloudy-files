package http

import (
	"net/http"

	"github.com/peteraba/cloudy-files/util"
)

const (
	HeaderAccept = "Accept"
)

const (
	ContentTypeJSON  = "application/json"
	ContentTypePlain = "text/plain"
	ContentTypeHTML  = "text/html"
)

var supportedTypes = []string{ContentTypePlain, ContentTypeJSON, ContentTypeHTML} //nolint:gochecknoglobals // This is a constant.

func IsJSONRequest(r *http.Request) bool {
	accept := r.Header.Get(HeaderAccept)

	contentType := util.NegotiateContentType(accept, supportedTypes)

	return contentType == ContentTypeJSON
}
