package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/apperr"
)

const (
	HeaderContentType        = "Content-Type"
	HeaderAccept             = "Accept"
	HeaderContentLength      = "Content-Length"
	HeaderContentTypeOptions = "X-Content-Type-Options"
)

const (
	ContentTypeJSON     = "application/json"
	ContentTypeJSONUTF8 = "application/json; charset=utf-8"
	ContentTypeText     = "text/plain"
	ContentTypeHTML     = "text/html"
	ContentTypeHTMLUTF8 = "text/html; charset=utf-8"
)

var supportedTypes = []string{ContentTypeJSON, ContentTypeHTML} //nolint:gochecknoglobals // This is a constant.

func isJSONRequest(r *http.Request) bool {
	accept := r.Header.Get(HeaderAccept)

	contentType := negotiateContentType(accept, supportedTypes)

	return contentType == ContentTypeJSON
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

	// No match found, assume the first supported type
	return supportedTypes[0]
}

func problem(w http.ResponseWriter, r *http.Request, err error, logger *log.Logger) {
	logger.Error().Err(err).Msg("Error")

	httpError := apperr.GetProblem(err)

	header := w.Header()

	// Delete the Content-Length header, which might be for some other content.
	// Assuming the error string fits in the writer's buffer, we'll figure
	// out the correct Content-Length for it later.
	//
	// We don't delete Content-Encoding, because some middleware sets
	// Content-Encoding: gzip and wraps the ResponseWriter to compress on-the-fly.
	// See https://go.dev/issue/66343.
	header.Del(HeaderContentLength)

	header.Set(HeaderContentTypeOptions, "nosniff")
	w.WriteHeader(httpError.Status)

	if isJSONRequest(r) {
		sendJSON(w, httpError, logger)
	} else {
		sendHTML(w, httpError, logger)
	}
}

func sendJSON(w http.ResponseWriter, content interface{}, logger *log.Logger) {
	w.Header().Set(HeaderContentType, ContentTypeJSONUTF8)

	if content == nil {
		return
	}

	payload, err := json.Marshal(content)
	if err != nil {
		logger.Error().Err(err).Msg("Error during rendering JSON.")

		return
	}

	_, err = w.Write(payload)
	if err != nil {
		logger.Error().Err(err).Msg("Error during writing content.")

		return
	}
}

func sendHTML(w http.ResponseWriter, content interface{}, logger *log.Logger) {
	w.Header().Set(HeaderContentType, ContentTypeHTMLUTF8)

	body := fmt.Sprint(content)
	if httpError, ok := content.(*apperr.Problem); ok {
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
		logger.Error().Err(err).Msg("Error during writing content.")
	}
}
