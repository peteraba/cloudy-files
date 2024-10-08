package web

import (
	"fmt"
	"net/http"

	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/apperr"
	"github.com/peteraba/cloudy-files/http/inandout"
)

func Problem(w http.ResponseWriter, logger *log.Logger, err error) {
	logger.Error().Err(err).Msg("Error")

	problem := apperr.GetProblem(err)

	header := w.Header()

	// Delete the Content-Length header, which might be for some other content.
	// Assuming the error string fits in the writer's buffer, we'll figure
	// out the correct Content-Length for it later.
	//
	// We don't delete Content-Encoding, because some middleware sets
	// Content-Encoding: gzip and wraps the ResponseWriter to compress on-the-fly.
	// See https://go.dev/issue/66343.
	header.Del(inandout.HeaderContentLength)

	header.Set(inandout.HeaderContentTypeOptions, "nosniff")
	w.WriteHeader(problem.Status)

	Send(w, problem)
}

func Send(w http.ResponseWriter, content interface{}) {
	w.Header().Set(inandout.HeaderContentType, inandout.ContentTypeHTMLUTF8)
	w.WriteHeader(http.StatusOK)

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

	fmt.Fprintf(w, tmpl, body)
}
