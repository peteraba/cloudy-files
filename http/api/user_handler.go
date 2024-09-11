package api

import (
	"fmt"
	"net/http"

	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/apperr"
	"github.com/peteraba/cloudy-files/repo"
	"github.com/peteraba/cloudy-files/service"
)

type UserHandler struct {
	userService *service.User
	logger      *log.Logger
}

func NewUserHandler(userService *service.User, logger *log.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

// LoginRequest represents a password change request.
type LoginRequest struct {
	Username string `json:"username" formam:"username"`
	Password string `json:"password" formam:"password"`
	CSRF     string `json:"-"        formam:"csrf"`
}

// Login logs in a user via the API.
func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	loginRequest, err := Parse(r, LoginRequest{})
	if err != nil {
		Problem(w, err, uh.logger)

		return
	}

	session, err := uh.userService.Login(r.Context(), loginRequest.Username, loginRequest.Password)
	if err != nil {
		Problem(w, err, uh.logger)

		return
	}

	uh.logger.Info().
		Str("username", loginRequest.Username).
		Msg("Login successful.")

	send(w, session, uh.logger)
}

func (uh *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// TODO: Auth admin-only
	users, err := uh.userService.List(r.Context())
	if err != nil {
		Problem(w, err, uh.logger)

		return
	}

	send(w, users, uh.logger)
}

func (uh *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	// TODO: Auth admin-only
	userModel, err := Parse(r, repo.UserModel{})
	if err != nil {
		Problem(w, fmt.Errorf("failed to Parse user, err: %w", apperr.ErrBadRequest(err)), uh.logger)

		return
	}

	userModel, err = uh.userService.Create(r.Context(), userModel.Name, userModel.Email, userModel.Password, userModel.IsAdmin, userModel.Access)
	if err != nil {
		Problem(w, err, uh.logger)

		return
	}

	uh.logger.Info().Str("username", userModel.Name).Msg("User created.")

	send(w, userModel, uh.logger)
}

// PasswordChangeRequest represents a password change request.
type PasswordChangeRequest struct {
	Username string `json:"username" formam:"username"`
	Password string `json:"password" formam:"password"`
	CSRF     string `json:"-"        formam:"csrf"`
}

// UpdateUserPassword updates a user's password.
func (uh *UserHandler) UpdateUserPassword(w http.ResponseWriter, r *http.Request) {
	// TODO: Auth admin-only or self-service
	req, err := Parse(r, PasswordChangeRequest{})
	if err != nil {
		Problem(w, err, uh.logger)

		return
	}

	user, err := uh.userService.UpdatePassword(r.Context(), req.Username, req.Password)
	if err != nil {
		Problem(w, err, uh.logger)

		return
	}

	send(w, user, uh.logger)
}

// AccessChangeRequest represents an access change request.
type AccessChangeRequest struct {
	Username string   `json:"username" formam:"username"`
	Access   []string `json:"access"   formam:"access"`
	CSRF     string   `json:"-"        formam:"csrf"`
}

// UpdateUserAccess updates a user's access.
func (uh *UserHandler) UpdateUserAccess(w http.ResponseWriter, r *http.Request) {
	// TODO: Auth admin-only
	req, err := Parse(r, AccessChangeRequest{})
	if err != nil {
		Problem(w, err, uh.logger)

		return
	}

	user, err := uh.userService.UpdateAccess(r.Context(), req.Username, req.Access)
	if err != nil {
		Problem(w, err, uh.logger)

		return
	}

	send(w, user, uh.logger)
}

// UserNameOnlyRequest represents a request where the username is the only mandatory field.
type UserNameOnlyRequest struct {
	Username string `json:"username" formam:"username"`
	CSRF     string `json:"-"        formam:"csrf"`
}

// PromoteUser promotes a user to admin.
func (uh *UserHandler) PromoteUser(w http.ResponseWriter, r *http.Request) {
	// TODO: Auth admin-only
	ctx := r.Context()
	name := r.PathValue("id")

	user, err := uh.userService.Promote(ctx, name)
	if err != nil {
		Problem(w, err, uh.logger)

		return
	}

	send(w, user, uh.logger)
}

// DemoteUser demotes a user from admin.
func (uh *UserHandler) DemoteUser(w http.ResponseWriter, r *http.Request) {
	// TODO: Auth admin-only
	ctx := r.Context()
	name := r.PathValue("id")

	user, err := uh.userService.Demote(ctx, name)
	if err != nil {
		Problem(w, err, uh.logger)

		return
	}

	send(w, user, uh.logger)
}

// DeleteUser deletes a user.
func (uh *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// TODO: Auth admin-only
	ctx := r.Context()
	name := r.PathValue("id")

	err := uh.userService.Delete(ctx, name)
	if err != nil {
		Problem(w, err, uh.logger)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
