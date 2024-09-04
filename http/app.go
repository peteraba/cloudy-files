package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/apperr"
	"github.com/peteraba/cloudy-files/service"
)

const (
	paramUsername = "username"
	paramEmail    = "email"
	paramPassword = "password"
	paramIsAdmin  = "isAdmin"
	paramAccess   = "access"
)

const (
	cancelTime = 5 * time.Second
)

// App represents the command line interface.
type App struct {
	sessionService *service.Session
	userService    *service.User
	fileService    *service.File
	logger         *log.Logger
}

// NewApp creates a new App instance.
func NewApp(sessionService *service.Session, userService *service.User, fileService *service.File, logger *log.Logger) *App {
	return &App{
		sessionService: sessionService,
		userService:    userService,
		fileService:    fileService,
		logger:         logger,
	}
}

// Route sets up the HTTP server.
func (a *App) Route() *http.ServeMux {
	mux := http.NewServeMux()

	// List Users (Admin-only)
	mux.HandleFunc("GET /users", a.ListUsers)

	// Get User (Admin-only)
	mux.HandleFunc("GET /users/{id}", a.GetUser)

	// Create User (Admin-only)
	mux.HandleFunc("POST /users", a.CreateUser)

	// Update User (Admin-only)
	mux.HandleFunc("PUT /users/{id}", a.UpdateUser)

	// Update User (Admin-only)
	mux.HandleFunc("DELETE /users/{id}", a.DeleteUser)

	// List Files (Logged-in)
	mux.HandleFunc("GET /files", a.ListFiles)

	// Delete File (Admin-only)
	mux.HandleFunc("DELETE /files/{id}", a.DeleteFile)

	// Login (Any)
	mux.HandleFunc("GET /user-logins", a.Login)

	// Upload File (Admin-only)
	mux.HandleFunc("POST /file-uploads", a.UploadFile)

	// Retrieve File (Logged-in)
	mux.HandleFunc("GET /file-uploads", a.RetrieveFile)

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

// ListUsers lists users.
func (a *App) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := a.userService.List(r.Context())
	if err != nil {
		a.error(w, r, err)

		return
	}

	if isJSONRequest(r) {
		a.json(w, users)

		return
	}

	a.html(w, users)
}

// GetUser retrieves a user.
func (a *App) GetUser(w http.ResponseWriter, r *http.Request) {
	_ = r

	a.error(w, r, apperr.ErrNotImplemented)
}

// CreateUser creates a user.
func (a *App) CreateUser(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	username := query.Get(paramUsername)
	password := query.Get(paramPassword)
	email := query.Get(paramEmail)
	isAdmin := false

	switch strings.ToLower(query.Get(paramIsAdmin)) {
	case "true", "1", "yes", "y":
		isAdmin = true
	}

	var access []string
	if query.Get(paramAccess) != "" {
		access = strings.Split(query.Get(paramAccess), ",")
	}

	userModel, err := a.userService.Create(r.Context(), username, email, password, isAdmin, access)
	if err != nil {
		a.error(w, r, err)
	}

	a.logger.Info().Msg("User created")

	a.json(w, userModel)
}

// UpdateUser updates a user.
func (a *App) UpdateUser(w http.ResponseWriter, r *http.Request) {
	_ = r

	a.error(w, r, apperr.ErrNotImplemented)
}

// DeleteUser deletes a user.
func (a *App) DeleteUser(w http.ResponseWriter, r *http.Request) {
	_ = r

	a.error(w, r, apperr.ErrNotImplemented)
}

// ListFiles lists files.
func (a *App) ListFiles(w http.ResponseWriter, r *http.Request) {
	files, err := a.fileService.List(r.Context(), nil, true)
	if err != nil {
		a.error(w, r, err)

		return
	}

	if isJSONRequest(r) {
		a.json(w, files)

		return
	}

	fileHTML := make([]string, 0, len(files))
	for _, file := range files {
		fileHTML = append(fileHTML, fmt.Sprintf(
			`<tr>
	<td>%s</td>
	<td>%s</td>
</tr>
`,
			file.Name,
			strings.Join(file.Access, ", "),
		))
	}

	tmpl := fmt.Sprintf(
		`<table>
	<thead>
		<tr>
			<th>Name</th>
			<th>Access</th>
		</tr>
	</thead>
	<tbody>
%s
	</tbody>
</table>
`,
		strings.Join(fileHTML, ""),
	)

	a.html(w, tmpl)
}

// DeleteFile deletes a file.
func (a *App) DeleteFile(w http.ResponseWriter, r *http.Request) {
	_ = r

	a.error(w, r, apperr.ErrNotImplemented)
}

// Login logs in a user.
func (a *App) Login(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	username := query.Get(paramUsername)
	password := query.Get(paramPassword)

	session, err := a.userService.Login(r.Context(), username, password)
	if err != nil {
		a.error(w, r, err)

		return
	}

	a.logger.Info().
		Str(paramUsername, username).
		Str("hash", session.Hash).
		Msg("Login successful.")

	a.nobody(w)
}

// UploadFile uploads a file.
func (a *App) UploadFile(w http.ResponseWriter, r *http.Request) {
	_ = r

	a.error(w, r, apperr.ErrNotImplemented)
}

// RetrieveFile retrieves a file.
func (a *App) RetrieveFile(w http.ResponseWriter, r *http.Request) {
	_ = r

	a.error(w, r, apperr.ErrNotImplemented)
}

// Home is the default route.
func (a *App) Home(w http.ResponseWriter, r *http.Request) {
	_ = r

	a.error(w, r, apperr.ErrNotImplemented)
}
