package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/peteraba/cloudy-files/apperr"
)

const (
	headerContentType        = "Content-Type"
	headerAccept             = "Accept"
	headerContentLength      = "Content-Length"
	headerContentTypeOptions = "X-Content-Type-Options"
)

const (
	contentTypeJSON = "application/json"
	contentTypeText = "text/plain"
	contentTypeHTML = "text/html"
)

var supportedTypes = []string{contentTypeJSON, contentTypeHTML} //nolint:gochecknoglobals // This is a constant.

func isJSONRequest(r *http.Request) bool {
	accept := r.Header.Get(headerAccept)

	contentType := negotiateContentType(accept, supportedTypes)

	return contentType == contentTypeJSON
}

func negotiateContentType(accept string, supportedTypes []string) string {
	if accept == "" {
		// No Accept header, assume the first supported type
		return supportedTypes[0]
	}

	// Parse the Accept header
	acceptedTypes := strings.Split(accept, ",")
	for _, acceptedType := range acceptedTypes {
		mediaType := strings.TrimSpace(strings.Split(acceptedType, ";")[0])
		for _, supportedType := range supportedTypes {
			if mediaType == supportedType || mediaType == "*/*" {
				return supportedType
			}
		}
	}

	// No match found
	return ""
}

func (a *App) error(w http.ResponseWriter, r *http.Request, err error) {
	a.logger.Error().Err(err).Msg("Error")

	httpError := apperr.GetHTTPError(err)

	header := w.Header()

	// Delete the Content-Length header, which might be for some other content.
	// Assuming the error string fits in the writer's buffer, we'll figure
	// out the correct Content-Length for it later.
	//
	// We don't delete Content-Encoding, because some middleware sets
	// Content-Encoding: gzip and wraps the ResponseWriter to compress on-the-fly.
	// See https://go.dev/issue/66343.
	header.Del(headerContentLength)

	header.Set(headerContentTypeOptions, "nosniff")
	w.WriteHeader(httpError.Status)

	if isJSONRequest(r) {
		a.json(w, httpError)
	} else {
		a.html(w, httpError)
	}
}

func (a *App) nobody(w http.ResponseWriter) {
	w.Header().Set(headerContentType, contentTypeText+"; charset=utf-8")
}

func (a *App) json(w http.ResponseWriter, content interface{}) {
	w.Header().Set(headerContentType, contentTypeJSON+"; charset=utf-8")

	if content == nil {
		return
	}

	payload, err := json.Marshal(content)
	if err != nil {
		a.logger.Error().Err(err).Msg("Error during rendering JSON.")

		return
	}

	_, err = w.Write(payload)
	if err != nil {
		a.logger.Error().Err(err).Msg("Error during writing content.")

		return
	}
}

func (a *App) html(w http.ResponseWriter, content interface{}) {
	w.Header().Set(headerContentType, contentTypeHTML+"; charset=utf-8")

	body := fmt.Sprint(content)
	if httpError, ok := content.(*apperr.HTTPError); ok {
		body = fmt.Sprintf(`
  <h3>Error</h3>
<table>
  <thead>
    <tr>
      <th>Field</th>
      <th>Value</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>Title</td>
      <td>%s</td>
    </tr>
    <tr>
      <td>Status</td>
      <td>%d</td>
    </tr>
    <tr>
      <td>Details</td>
      <td>%s</td>
    </tr>
  </tbody>
</table>
</section>`, httpError.Title, httpError.Status, httpError.Detail)
	}

	tmpl := `<html>
	<head>
		<link rel="stylesheet" href="https://fonts.googleapis.com/css?family=Roboto:300,300italic,700,700italic">
		<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/normalize/8.0.1/normalize.css">
		<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/milligram/1.4.1/milligram.css">
	</head>
	<body>
		<main class="wrapper">
			<nav class="navigation"><section class="container">&nbsp;</section></nav>
			<section class="container">
%s
			</section>
		</main>
	</body>
</html>
`

	_, err := fmt.Fprintf(w, tmpl, body)
	if err != nil {
		a.logger.Error().Err(err).Msg("Error during writing content.")
	}
}
