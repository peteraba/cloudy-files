package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/phuslu/log"
)

const (
	cancelTime = 5 * time.Second
)

// App represents the command line interface.
type App struct {
	userHandler *UserHandler
	fileHandler *FileHandler
	logger      *log.Logger
}

// NewApp creates a new App instance.
func NewApp(users *UserHandler, files *FileHandler, logger *log.Logger) *App {
	return &App{
		userHandler: users,
		fileHandler: files,
		logger:      logger,
	}
}

// Route sets up the HTTP server.
func (a *App) Route() *http.ServeMux {
	mux := http.NewServeMux()

	a.userHandler.SetupRoutes(mux)
	a.fileHandler.SetupRoutes(mux)

	mux.HandleFunc("GET /", a.Home)

	return mux
}

// Start starts the HTTP server.
func (a *App) Start(mux *http.ServeMux) {
	srv := &http.Server{ //nolint:exhaustruct // This is a big struct
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: cancelTime,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

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

type HealthResponse struct {
	Status string `json:"status"`
}

// Home is the default route.
func (a *App) Home(w http.ResponseWriter, r *http.Request) {
	if isJSONRequest(r) {
		sendJSON(w, HealthResponse{Status: "ok"}, a.logger)

		return
	}

	// TODO: Generate and store CSRF token
	csrf := "TODO"

	tmpl := fmt.Sprintf(
		`<form>
  <fieldset>
    <label for="nameField">Name</label>
    <input type="text" name="username" placeholder="peter81" id="nameField">
    <label for="passField">Password</label>
    <input type="password" name="password" placeholder="verysecretpass" id="passField">
    <input type="hidden" name="csrf" value="%s">
    <input class="button-primary" type="submit" value="Send">
  </fieldset>
</form>
`,
		csrf,
	)

	sendHTML(w, tmpl, a.logger)
}
