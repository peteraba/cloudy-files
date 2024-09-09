package api

import (
	"net/http"

	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/service"
)

type FileHandler struct {
	sessionService *service.Session
	fileService    *service.File
	logger         *log.Logger
}

func NewFileHandler(sessionService *service.Session, fileService *service.File, logger *log.Logger) *FileHandler {
	return &FileHandler{
		sessionService: sessionService,
		fileService:    fileService,
		logger:         logger,
	}
}

// ListFiles lists files.
func (fh *FileHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	files, err := fh.fileService.List(r.Context(), nil, true)
	if err != nil {
		Problem(w, err, fh.logger)

		return
	}

	send(w, files, fh.logger)
}
