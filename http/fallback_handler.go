package http

import (
	"net/http"

	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/http/api"
	"github.com/peteraba/cloudy-files/http/web"
)

type FallbackHandler struct {
	api    *api.FallbackHandler
	web    *web.FallbackHandler
	logger *log.Logger
}

func NewFallbackHandler(apiHandler *api.FallbackHandler, webHandler *web.FallbackHandler, logger *log.Logger) *FallbackHandler {
	return &FallbackHandler{
		api:    apiHandler,
		web:    webHandler,
		logger: logger,
	}
}

// SetupRoutes sets up the HTTP server.
func (fh *FallbackHandler) SetupRoutes(mux *http.ServeMux) *http.ServeMux {
	mux.HandleFunc("GET /", fh.Home)

	return mux
}

// Home is the landing page route for users not logged in.
func (fh *FallbackHandler) Home(w http.ResponseWriter, r *http.Request) {
	if IsJSONRequest(r) {
		fh.api.Home(w)

		return
	}

	fh.web.Home(w)
}
