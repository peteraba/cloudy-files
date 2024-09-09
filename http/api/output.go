package api

import (
	"encoding/json"
	"net/http"

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
)

func Problem(w http.ResponseWriter, err error, logger *log.Logger) {
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
	header.Del(HeaderContentLength)

	header.Set(HeaderContentTypeOptions, "nosniff")
	w.WriteHeader(problem.Status)

	send(w, problem, logger)
}

func send(w http.ResponseWriter, content interface{}, logger *log.Logger) {
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
