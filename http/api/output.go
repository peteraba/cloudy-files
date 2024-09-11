package api

import (
	"encoding/json"
	"net/http"

	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/apperr"
	"github.com/peteraba/cloudy-files/http/inandout"
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
	header.Del(inandout.HeaderContentLength)

	header.Set(inandout.HeaderContentTypeOptions, "nosniff")
	w.WriteHeader(problem.Status)

	Send(w, problem, logger)
}

func Send(w http.ResponseWriter, content interface{}, logger *log.Logger) {
	w.Header().Set(inandout.HeaderContentType, inandout.ContentTypeJSONUTF8)

	if content == nil {
		return
	}

	payload, err := json.Marshal(content)
	if err != nil {
		logger.Error().Err(err).Msg("Error during marshaling JSON.")

		return
	}

	w.Write(payload) //nolint:errcheck // We don't care about the error here.
}
