package http

import (
	"net/http"

	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/apperr"
	"github.com/peteraba/cloudy-files/http/api"
	"github.com/peteraba/cloudy-files/http/web"
)

type FileHandler struct {
	api    *api.FileHandler
	web    *web.FileHandler
	logger *log.Logger
}

func NewFileHandler(apiHandler *api.FileHandler, webHandler *web.FileHandler, logger *log.Logger) *FileHandler {
	return &FileHandler{
		api:    apiHandler,
		web:    webHandler,
		logger: logger,
	}
}

// SetupRoutes sets up the HTTP server.
func (fh *FileHandler) SetupRoutes(mux *http.ServeMux) *http.ServeMux {
	mux.HandleFunc("GET /files", fh.ListFiles)
	mux.HandleFunc("DELETE /files/{id}", fh.NotImplemented)
	mux.HandleFunc("POST /file-uploads", fh.NotImplemented)
	mux.HandleFunc("GET /file-uploads", fh.NotImplemented)

	return mux
}

// ListFiles lists files.
func (fh *FileHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	if IsJSONRequest(r) {
		fh.api.ListFiles(w, r)

		return
	}

	fh.web.ListFiles(w, r)
}

func (fh *FileHandler) NotImplemented(w http.ResponseWriter, r *http.Request) {
	if IsJSONRequest(r) {
		api.Problem(w, apperr.ErrNotImplemented, fh.logger)

		return
	}

	web.Problem(w, fh.logger, apperr.ErrNotImplemented)
}
