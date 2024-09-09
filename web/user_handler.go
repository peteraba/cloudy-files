package web

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/apperr"
	"github.com/peteraba/cloudy-files/repo"
	"github.com/peteraba/cloudy-files/service"
)

type UserHandler struct {
	sessionService *service.Session
	userService    *service.User
	logger         *log.Logger
}

func NewUserHandler(sessionService *service.Session, userService *service.User, logger *log.Logger) *UserHandler {
	return &UserHandler{
		sessionService: sessionService,
		userService:    userService,
		logger:         logger,
	}
}

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

// LoginRequest represents a password change request.
type LoginRequest struct {
	Username string `json:"username" formam:"username"`
	Password string `json:"password" formam:"password"`
	CSRF     string `json:"-"        formam:"csrf"`
}

// Login logs in a user.
func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if IsJSONRequest(r) {
		uh.LoginAPI(w, r)

		return
	}

	uh.LoginHTML(w, r)
}

// LoginHTML logs in a user via the HTML form.
func (uh *UserHandler) LoginHTML(w http.ResponseWriter, r *http.Request) {
	loginRequest, err := Parse(r, LoginRequest{})
	if err != nil {
		// TODO: flash errors
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	// TODO: verify CSRF token

	// attempt to start a new session with the login credentials
	session, err := uh.userService.Login(r.Context(), loginRequest.Username, loginRequest.Password)
	if err != nil {
		// TODO: flash errors
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	// log the successful login
	uh.logger.Info().
		Str("username", loginRequest.Username).
		Str("hash", session.Hash).
		Msg("Login successful.")

	// TODO: flash errors

	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

// LoginAPI logs in a user via the API.
func (uh *UserHandler) LoginAPI(w http.ResponseWriter, r *http.Request) {
	loginRequest, err := Parse(r, LoginRequest{})
	if err != nil {
		problem(w, r, err, uh.logger)

		return
	}

	session, err := uh.userService.Login(r.Context(), loginRequest.Username, loginRequest.Password)
	if err != nil {
		problem(w, r, err, uh.logger)

		return
	}

	uh.logger.Info().
		Str("username", loginRequest.Username).
		Str("hash", session.Hash).
		Msg("Login successful.")

	sendJSON(w, session, uh.logger)
}

func (uh *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := uh.userService.List(r.Context())
	if err != nil {
		problem(w, r, err, uh.logger)

		return
	}

	if IsJSONRequest(r) {
		sendJSON(w, users, uh.logger)

		return
	}

	userHTML := make([]string, 0, len(users))
	for _, user := range users {
		userHTML = append(userHTML, fmt.Sprintf(
			`<tr>
	<td>%s</td>
	<td>%s</td>
</tr>
`,
			user.Name,
			strings.Join(user.Access, ", "),
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
		strings.Join(userHTML, ""),
	)

	sendHTML(w, tmpl, uh.logger)
}

func (uh *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	userModel, err := Parse(r, repo.UserModel{})
	if err != nil {
		problem(w, r, fmt.Errorf("failed to Parse user, err: %w", apperr.ErrBadRequest(err)), uh.logger)

		return
	}

	userModel, err = uh.userService.Create(r.Context(), userModel.Name, userModel.Email, userModel.Password, userModel.IsAdmin, userModel.Access)
	if err != nil {
		problem(w, r, err, uh.logger)

		return
	}

	uh.logger.Info().Str("username", userModel.Name).Msg("User created.")

	if IsJSONRequest(r) {
		sendJSON(w, userModel, uh.logger)

		return
	}

	// TODO: flash errors

	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

// PasswordChangeRequest represents a password change request.
type PasswordChangeRequest struct {
	Username string `json:"username" formam:"username"`
	Password string `json:"password" formam:"password"`
	CSRF     string `json:"-"        formam:"csrf"`
}

// UpdateUserPassword updates a user's password.
func (uh *UserHandler) UpdateUserPassword(w http.ResponseWriter, r *http.Request) {
	req, err := Parse(r, PasswordChangeRequest{})
	if err != nil {
		problem(w, r, err, uh.logger)

		return
	}

	user, err := uh.userService.UpdatePassword(r.Context(), req.Username, req.Password)
	if err != nil {
		problem(w, r, err, uh.logger)

		return
	}

	if IsJSONRequest(r) {
		sendJSON(w, user, uh.logger)

		return
	}

	// TODO: flash errors

	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

// AccessChangeRequest represents an access change request.
type AccessChangeRequest struct {
	Username string   `json:"username" formam:"username"`
	Access   []string `json:"access"   formam:"access"`
	CSRF     string   `json:"-"        formam:"csrf"`
}

// UpdateUserAccess updates a user's access.
func (uh *UserHandler) UpdateUserAccess(w http.ResponseWriter, r *http.Request) {
	req, err := Parse(r, AccessChangeRequest{})
	if err != nil {
		problem(w, r, err, uh.logger)

		return
	}

	user, err := uh.userService.UpdateAccess(r.Context(), req.Username, req.Access)
	if err != nil {
		problem(w, r, err, uh.logger)

		return
	}

	if IsJSONRequest(r) {
		sendJSON(w, user, uh.logger)

		return
	}

	// TODO: flash errors

	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

// UserNameOnlyRequest represents a request where the username is the only mandatory field.
type UserNameOnlyRequest struct {
	Username string `json:"username" formam:"username"`
	CSRF     string `json:"-"        formam:"csrf"`
}

// PromoteUser promotes a user to admin.
func (uh *UserHandler) PromoteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := r.PathValue("id")

	user, err := uh.userService.Promote(ctx, name)
	if err != nil {
		problem(w, r, err, uh.logger)

		return
	}

	if IsJSONRequest(r) {
		sendJSON(w, user, uh.logger)

		return
	}

	// TODO: flash errors

	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

// DemoteUser demotes a user from admin.
func (uh *UserHandler) DemoteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := r.PathValue("id")

	user, err := uh.userService.Demote(ctx, name)
	if err != nil {
		problem(w, r, err, uh.logger)

		return
	}

	if IsJSONRequest(r) {
		sendJSON(w, user, uh.logger)

		return
	}

	// TODO: flash errors

	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

// DeleteUser deletes a user.
func (uh *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := r.PathValue("id")

	err := uh.userService.Delete(ctx, name)
	if err != nil {
		problem(w, r, err, uh.logger)

		return
	}

	if IsJSONRequest(r) {
		w.WriteHeader(http.StatusNoContent)

		return
	}

	// TODO: flash errors

	http.Redirect(w, r, "/users", http.StatusSeeOther)
}
