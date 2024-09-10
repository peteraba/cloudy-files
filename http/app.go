package http

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/phuslu/log"
)

const (
	cancelTime = 5 * time.Second
)

// App represents the command line interface.
type App struct {
	userHandler     *UserHandler
	fileHandler     *FileHandler
	fallbackHandler *FallbackHandler
	logger          *log.Logger
}

// NewApp creates a new App instance.
func NewApp(users *UserHandler, files *FileHandler, fallback *FallbackHandler, logger *log.Logger) *App {
	return &App{
		userHandler:     users,
		fileHandler:     files,
		fallbackHandler: fallback,
		logger:          logger,
	}
}

// Route sets up the HTTP server.
func (a *App) Route() *http.ServeMux {
	mux := http.NewServeMux()

	a.userHandler.SetupRoutes(mux)
	a.fileHandler.SetupRoutes(mux)
	a.fallbackHandler.SetupRoutes(mux)

	return mux
}

// Start starts the HTTP server.
func (a *App) Start(mux *http.ServeMux) {
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: cancelTime,
	}

	done := make(chan os.Signal, 1)

	go func() {
		a.logger.Info().Str("addr", srv.Addr).Msg("HTTP server starting.")

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Error().Err(err).Msg("HTTP server failed to start")
		}
	}()

	<-done
	a.logger.Info().Msg("HTTP server stopped.")

	ctx, cancel := context.WithTimeout(context.Background(), cancelTime)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		a.logger.Error().Err(err).Msg("Server shutdown failed.")
	}

	a.logger.Info().Msg("Server shutdown failed.")
}
