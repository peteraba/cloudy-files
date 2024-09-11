package api

import (
	"net/http"

	"github.com/phuslu/log"
)

type FallbackHandler struct {
	logger *log.Logger
}

func NewFallbackHandler(logger *log.Logger) *FallbackHandler {
	return &FallbackHandler{
		logger: logger,
	}
}

type HealthResponse struct {
	Status string `json:"status"`
}

func (fh *FallbackHandler) Home(w http.ResponseWriter) {
	Send(w, HealthResponse{Status: "ok"}, fh.logger)
}
