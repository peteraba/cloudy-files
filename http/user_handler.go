package http

import (
	"net/http"

	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/http/api"
	"github.com/peteraba/cloudy-files/http/web"
)

// UserHandler handles user requests.
type UserHandler struct {
	api    *api.UserHandler
	web    *web.UserHandler
	logger *log.Logger
}

// NewUserHandler creates a new handler for user.
func NewUserHandler(apiHandler *api.UserHandler, webHandler *web.UserHandler, logger *log.Logger) *UserHandler {
	return &UserHandler{
		api:    apiHandler,
		web:    webHandler,
		logger: logger,
	}
}

// SetupRoutes sets up the HTTP handlers.
func (uh *UserHandler) SetupRoutes(mux *http.ServeMux) *http.ServeMux {
	mux.HandleFunc("POST /user-logins", uh.Login)
	mux.HandleFunc("POST /users", uh.CreateUser)
	mux.HandleFunc("GET /users", uh.ListUsers)
	mux.HandleFunc("PUT /users/{id}/passwords", uh.UpdateUserPassword)
	mux.HandleFunc("PUT /users/{id}/accesses", uh.UpdateUserAccess)
	mux.HandleFunc("PUT /users/{id}/promotions", uh.PromoteUser)
	mux.HandleFunc("PUT /users/{id}/demotions", uh.DemoteUser)
	mux.HandleFunc("DELETE /users/{id}", uh.DeleteUser)

	return mux
}

// Login logs in a user.
func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if IsJSONRequest(r) {
		uh.api.Login(w, r)

		return
	}

	uh.web.Login(w, r)
}

// ListUsers lists users.
func (uh *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if IsJSONRequest(r) {
		uh.api.ListUsers(w, r)

		return
	}

	uh.web.ListUsers(w, r)
}

// CreateUser creates a user.
func (uh *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if IsJSONRequest(r) {
		uh.api.CreateUser(w, r)

		return
	}

	uh.web.CreateUser(w, r)
}

// UpdateUserPassword updates a user's password.
func (uh *UserHandler) UpdateUserPassword(w http.ResponseWriter, r *http.Request) {
	if IsJSONRequest(r) {
		uh.api.UpdateUserPassword(w, r)

		return
	}

	uh.web.UpdateUserPassword(w, r)
}

func (uh *UserHandler) UpdateUserAccess(w http.ResponseWriter, r *http.Request) {
	if IsJSONRequest(r) {
		uh.api.UpdateUserAccess(w, r)

		return
	}

	uh.web.UpdateUserAccess(w, r)
}

// PromoteUser promotes a user to admin.
func (uh *UserHandler) PromoteUser(w http.ResponseWriter, r *http.Request) {
	if IsJSONRequest(r) {
		uh.api.PromoteUser(w, r)

		return
	}

	uh.web.PromoteUser(w, r)
}

// DemoteUser demotes a user from admin.
func (uh *UserHandler) DemoteUser(w http.ResponseWriter, r *http.Request) {
	if IsJSONRequest(r) {
		uh.api.DemoteUser(w, r)

		return
	}

	uh.web.DemoteUser(w, r)
}

// DeleteUser deletes a user.
func (uh *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if IsJSONRequest(r) {
		uh.api.DeleteUser(w, r)

		return
	}

	uh.web.DeleteUser(w, r)
}
