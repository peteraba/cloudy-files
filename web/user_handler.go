package web

import (
	"encoding/json"
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
	mux.HandleFunc("GET /users", uh.ListUsers)
	mux.HandleFunc("GET /users/{id}", uh.GetUser)
	mux.HandleFunc("POST /users", uh.CreateUser)
	mux.HandleFunc("PUT /users/{id}", uh.UpdateUser)
	mux.HandleFunc("DELETE /users/{id}", uh.DeleteUser)

	return mux
}

// Login logs in a user.
func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if isJSONRequest(r) {
		uh.LoginHTML(w, r)

		return
	}

	uh.LoginAPI(w, r)
}

// LoginForm represents a login form.
type LoginForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginHTML logs in a user via the HTML form.
func (uh *UserHandler) LoginHTML(w http.ResponseWriter, r *http.Request) {
	loginForm := LoginForm{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
	}

	// TODO: verify CSRF token

	// attempt to start a new session with the login credentials
	session, err := uh.userService.Login(r.Context(), loginForm.Username, loginForm.Password)
	if err != nil {
		problem(w, r, err, uh.logger)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	// log the successful login
	uh.logger.Info().
		Str("username", loginForm.Username).
		Str("hash", session.Hash).
		Msg("Login successful.")

	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

// LoginAPIRequest represents a login request.
type LoginAPIRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginAPI logs in a user via the API.
func (uh *UserHandler) LoginAPI(w http.ResponseWriter, r *http.Request) {
	var loginRequest LoginAPIRequest

	err := json.NewDecoder(r.Body).Decode(&loginRequest)
	if err != nil {
		problem(w, r, fmt.Errorf("failed to decode user, err: %w", apperr.ErrBadRequest(err)), uh.logger)

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

	if isJSONRequest(r) {
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

func (uh *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	problem(w, r, apperr.ErrNotImplemented, uh.logger)
}

func (uh *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user repo.UserModel

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		problem(w, r, fmt.Errorf("failed to decode user, err: %w", apperr.ErrBadRequest(err)), uh.logger)

		return
	}

	userModel, err := uh.userService.Create(r.Context(), user.Name, user.Email, user.Password, user.IsAdmin, user.Access)
	if err != nil {
		problem(w, r, err, uh.logger)
	}

	uh.logger.Info().Str("username", user.Name).Msg("User created.")

	if isJSONRequest(r) {
		sendJSON(w, userModel, uh.logger)

		return
	}

	userHTML := []string{
		fmt.Sprintf("<tr><td>%s</td><td>%s</td></tr>", "name", userModel.Name),
		fmt.Sprintf("<tr><td>%s</td><td>%s</td></tr>", "email", userModel.Email),
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

func (uh *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	problem(w, r, apperr.ErrNotImplemented, uh.logger)
}

func (uh *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	problem(w, r, apperr.ErrNotImplemented, uh.logger)
}
