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

const (
	AfterLoginLocation   = "/files"
	UserListLocation     = "/users"
	HomeRedirectLocation = "/"
)

// UserHandler handles user requests.
type UserHandler struct {
	service *service.User
	csrf    *repo.CSRF
	cookie  *service.Cookie
	logger  *log.Logger
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userService *service.User, csrfRepo *repo.CSRF, sessionService *service.Cookie, logger *log.Logger) *UserHandler {
	return &UserHandler{
		service: userService,
		csrf:    csrfRepo,
		cookie:  sessionService,
		logger:  logger,
	}
}

// LoginRequest represents a password change request.
type LoginRequest struct {
	Username string `json:"username" formam:"username"`
	Password string `json:"password" formam:"password"`
	CSRF     string `json:"-"        formam:"csrf"`
}

// Login logs in the user via the HTML form and redirects to the files list page.
// Expects CSRF, but not a valid session.
func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	loginRequest, err := Parse(r, LoginRequest{})
	if err != nil {
		uh.cookie.FlashError(w, r, UserListLocation, err, "Parsing login request failed.", loginRequest)

		return
	}

	ipAddress := GetIPAddress(r)

	err = uh.csrf.Use(r.Context(), ipAddress, loginRequest.CSRF)
	if err != nil {
		uh.cookie.FlashError(w, r, UserListLocation, err, "Checking CSRF token failed.", loginRequest)

		return
	}

	// attempt to start a new session with the login credentials
	session, err := uh.service.Login(r.Context(), loginRequest.Username, loginRequest.Password)
	if err != nil {
		uh.cookie.FlashError(w, r, UserListLocation, err, "Login failed.", session)

		return
	}

	// Start session
	uh.cookie.StoreSessionUser(w, session)

	uh.cookie.FlashMessage(w, r, AfterLoginLocation, "Login successful.")
}

// ListUsers lists all users.
// Expects a valid session and admin rights.
func (uh *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	userSession, err := uh.cookie.GetSessionUser(r)
	if err != nil {
		Problem(w, uh.logger, err)

		return
	}

	if !userSession.IsAdmin {
		Problem(w, uh.logger, apperr.ErrAccessDenied)

		return
	}

	users, err := uh.service.List(r.Context())
	if err != nil {
		Problem(w, uh.logger, err)

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

	Send(w, tmpl)
}

// CreateUser creates a new user and redirects to the users list page.
// Expects a valid session and admin rights.
// Expects a valid CSRF token.
func (uh *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	// TODO: CSRF protection
	userSession, err := uh.cookie.GetSessionUser(r)
	if err != nil {
		uh.cookie.FlashError(w, r, HomeRedirectLocation, err, "No session found.")

		return
	}

	if !userSession.IsAdmin {
		uh.cookie.FlashError(w, r, AfterLoginLocation, apperr.ErrAccessDenied, "User is not an admin.")

		return
	}

	user, err := Parse(r, repo.UserModel{})
	if err != nil {
		uh.cookie.FlashError(w, r, UserListLocation, err, "Failed to parse request.")

		return
	}

	newUser, err := uh.service.Create(r.Context(), user.Name, user.Email, user.Password, user.IsAdmin, user.Access)
	if err != nil {
		uh.cookie.FlashError(w, r, UserListLocation, err, "Failed to create user.")

		return
	}

	uh.cookie.FlashMessage(w, r, UserListLocation, "User created.", newUser)
}

// PasswordChangeRequest represents a password change request.
type PasswordChangeRequest struct {
	Username string `json:"username" formam:"username"`
	Password string `json:"password" formam:"password"`
	CSRF     string `json:"-"        formam:"csrf"`
}

// UpdateUserPassword updates the password of a user and redirects to the users list page.
// Expects a valid session and admin rights.
// Expects a valid CSRF token.
func (uh *UserHandler) UpdateUserPassword(w http.ResponseWriter, r *http.Request) {
	// TODO: CSRF protection
	userSession, err := uh.cookie.GetSessionUser(r)
	if err != nil {
		uh.cookie.FlashError(w, r, HomeRedirectLocation, err, "No session found.")

		return
	}

	if !userSession.IsAdmin {
		uh.cookie.FlashError(w, r, AfterLoginLocation, apperr.ErrAccessDenied, "User is not an admin.")

		return
	}

	req, err := Parse(r, PasswordChangeRequest{})
	if err != nil {
		uh.cookie.FlashError(w, r, UserListLocation, err, "Failed to parse request.")

		return
	}

	_, err = uh.service.UpdatePassword(r.Context(), req.Username, req.Password)
	if err != nil {
		uh.cookie.FlashError(w, r, UserListLocation, err, "Failed to update user password.")

		return
	}

	uh.cookie.FlashMessage(w, r, UserListLocation, "User password updated.")
}

// AccessChangeRequest represents an access change request.
type AccessChangeRequest struct {
	Username string   `json:"username" formam:"username"`
	Access   []string `json:"access"   formam:"access"`
	CSRF     string   `json:"-"        formam:"csrf"`
}

// UpdateUserAccess updates the access list of a user and redirects to the users list page.
// Expects a valid session and admin rights.
// Expects a valid CSRF token.
func (uh *UserHandler) UpdateUserAccess(w http.ResponseWriter, r *http.Request) {
	// TODO: CSRF protection
	userSession, err := uh.cookie.GetSessionUser(r)
	if err != nil {
		uh.cookie.FlashError(w, r, HomeRedirectLocation, err, "No session found.")

		return
	}

	if !userSession.IsAdmin {
		uh.cookie.FlashError(w, r, AfterLoginLocation, apperr.ErrAccessDenied, "User is not an admin.")

		return
	}

	req, err := Parse(r, AccessChangeRequest{})
	if err != nil {
		uh.cookie.FlashError(w, r, UserListLocation, err, "Failed to parse request.")

		return
	}

	_, err = uh.service.UpdateAccess(r.Context(), req.Username, req.Access)
	if err != nil {
		uh.cookie.FlashError(w, r, UserListLocation, err, "Failed to update user access.")

		return
	}

	uh.cookie.FlashMessage(w, r, UserListLocation, "User access updated.")
}

// UserNameOnlyRequest represents a request where the username is the only mandatory field.
type UserNameOnlyRequest struct {
	Username string `json:"username" formam:"username"`
	CSRF     string `json:"-"        formam:"csrf"`
}

// PromoteUser promotes a user to admin and redirects to the users list page.
// Expects a valid session and admin rights.
// Expects a valid CSRF token.
func (uh *UserHandler) PromoteUser(w http.ResponseWriter, r *http.Request) {
	// TODO: CSRF protection
	userSession, err := uh.cookie.GetSessionUser(r)
	if err != nil {
		uh.cookie.FlashError(w, r, HomeRedirectLocation, err, "No session found.")

		return
	}

	if !userSession.IsAdmin {
		uh.cookie.FlashError(w, r, AfterLoginLocation, apperr.ErrAccessDenied, "User is not an admin.")

		return
	}

	ctx := r.Context()
	name := r.PathValue("id")

	_, err = uh.service.Promote(ctx, name)
	if err != nil {
		uh.cookie.FlashError(w, r, UserListLocation, err, "Failed to promote user.")

		return
	}

	uh.cookie.FlashMessage(w, r, UserListLocation, "User promoted.")
}

// DemoteUser demotes a user from admin to regular user and redirects to the users list page.
// Expects a valid session and admin rights.
// Expects a valid CSRF token.
func (uh *UserHandler) DemoteUser(w http.ResponseWriter, r *http.Request) {
	// TODO: CSRF protection
	userSession, err := uh.cookie.GetSessionUser(r)
	if err != nil {
		uh.cookie.FlashError(w, r, HomeRedirectLocation, err, "No session found.")

		return
	}

	if !userSession.IsAdmin {
		uh.cookie.FlashError(w, r, AfterLoginLocation, apperr.ErrAccessDenied, "User is not an admin.")

		return
	}

	ctx := r.Context()
	name := r.PathValue("id")

	_, err = uh.service.Demote(ctx, name)
	if err != nil {
		uh.cookie.FlashError(w, r, UserListLocation, err, "Failed to demote user.")

		return
	}

	uh.cookie.FlashMessage(w, r, UserListLocation, "User demoted.")
}

// DeleteUser deletes a user and redirects to the users list page.
// Expects a valid session and admin rights.
// Expects a valid CSRF token.
func (uh *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// TODO: CSRF protection
	userSession, err := uh.cookie.GetSessionUser(r)
	if err != nil {
		uh.cookie.FlashError(w, r, HomeRedirectLocation, err, "No session found.")

		return
	}

	if !userSession.IsAdmin {
		uh.cookie.FlashError(w, r, AfterLoginLocation, apperr.ErrAccessDenied, "User is not an admin.")

		return
	}

	ctx := r.Context()
	name := r.PathValue("id")

	err = uh.service.Delete(ctx, name)
	if err != nil {
		uh.cookie.FlashError(w, r, UserListLocation, err, "Failed to delete user.")

		return
	}

	uh.cookie.FlashMessage(w, r, UserListLocation, "User deleted.")
}
