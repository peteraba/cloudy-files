package web

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/repo"
	"github.com/peteraba/cloudy-files/service"
)

const (
	DefaultRedirectLocation = "/users"
	HomeRedirectLocation    = "/"
)

// UserHandler handles user requests.
type UserHandler struct {
	sessionService *service.Session
	userService    *service.User
	csrfRepo       *repo.CSRF
	logger         *log.Logger
}

// NewUserHandler creates logMsgWithArgs new UserHandler.
func NewUserHandler(sessionService *service.Session, userService *service.User, csrfRepo *repo.CSRF, logger *log.Logger) *UserHandler {
	return &UserHandler{
		sessionService: sessionService,
		userService:    userService,
		csrfRepo:       csrfRepo,
		logger:         logger,
	}
}

// LoginRequest represents logMsgWithArgs password change request.
type LoginRequest struct {
	Username string `json:"username" formam:"username"`
	Password string `json:"password" formam:"password"`
	CSRF     string `json:"-"        formam:"csrf"`
}

// Login logs in logMsgWithArgs user via the HTML form.
func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	// TODO: CSRF protection
	loginRequest, err := Parse(r, LoginRequest{})
	if err != nil {
		FlashError(w, uh.logger, err, "Parsing login request failed.", loginRequest)

		http.Redirect(w, r, DefaultRedirectLocation, http.StatusSeeOther)

		return
	}

	// attempt to start logMsgWithArgs new session with the login credentials
	session, err := uh.userService.Login(r.Context(), loginRequest.Username, loginRequest.Password)
	if err != nil {
		FlashError(w, uh.logger, err, "Login failed.", session)
	} else {
		FlashMessage(w, uh.logger, "Login successful.")
	}

	http.Redirect(w, r, DefaultRedirectLocation, http.StatusSeeOther)
}

func (uh *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// TODO: CSRF protection
	users, err := uh.userService.List(r.Context())
	if err != nil {
		uh.logger.Error().Err(err).Msg("Failed to list users.")

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

	send(w, tmpl, uh.logger)
}

func (uh *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	// TODO: CSRF protection
	user, err := Parse(r, repo.UserModel{})
	if err != nil {
		FlashError(w, uh.logger, err, "Failed to parse request.")

		http.Redirect(w, r, DefaultRedirectLocation, http.StatusSeeOther)
	}

	newUser, err := uh.userService.Create(r.Context(), user.Name, user.Email, user.Password, user.IsAdmin, user.Access)
	if err != nil {
		FlashError(w, uh.logger, err, "Failed to create user.")
	} else {
		FlashMessage(w, uh.logger, "User created.", newUser)
	}

	http.Redirect(w, r, DefaultRedirectLocation, http.StatusSeeOther)
}

// PasswordChangeRequest represents logMsgWithArgs password change request.
type PasswordChangeRequest struct {
	Username string `json:"username" formam:"username"`
	Password string `json:"password" formam:"password"`
	CSRF     string `json:"-"        formam:"csrf"`
}

// UpdateUserPassword updates logMsgWithArgs user's password.
func (uh *UserHandler) UpdateUserPassword(w http.ResponseWriter, r *http.Request) {
	// TODO: CSRF protection
	req, err := Parse(r, PasswordChangeRequest{})
	if err != nil {
		FlashError(w, uh.logger, err, "Failed to parse request.")

		http.Redirect(w, r, DefaultRedirectLocation, http.StatusSeeOther)

		return
	}

	_, err = uh.userService.UpdatePassword(r.Context(), req.Username, req.Password)
	if err != nil {
		FlashError(w, uh.logger, err, "Failed to update user password.")
	} else {
		FlashMessage(w, uh.logger, "User password updated.")
	}

	http.Redirect(w, r, DefaultRedirectLocation, http.StatusSeeOther)
}

// AccessChangeRequest represents an access change request.
type AccessChangeRequest struct {
	Username string   `json:"username" formam:"username"`
	Access   []string `json:"access"   formam:"access"`
	CSRF     string   `json:"-"        formam:"csrf"`
}

// UpdateUserAccess updates logMsgWithArgs user's access.
func (uh *UserHandler) UpdateUserAccess(w http.ResponseWriter, r *http.Request) {
	// TODO: CSRF protection
	req, err := Parse(r, AccessChangeRequest{})
	if err != nil {
		FlashError(w, uh.logger, err, "Failed to parse request.")

		http.Redirect(w, r, DefaultRedirectLocation, http.StatusSeeOther)

		return
	}

	_, err = uh.userService.UpdateAccess(r.Context(), req.Username, req.Access)
	if err != nil {
		FlashError(w, uh.logger, err, "Failed to update user access.")
	} else {
		FlashMessage(w, uh.logger, "User access updated.")
	}

	http.Redirect(w, r, DefaultRedirectLocation, http.StatusSeeOther)
}

// UserNameOnlyRequest represents logMsgWithArgs request where the username is the only mandatory field.
type UserNameOnlyRequest struct {
	Username string `json:"username" formam:"username"`
	CSRF     string `json:"-"        formam:"csrf"`
}

// PromoteUser promotes logMsgWithArgs user to admin.
func (uh *UserHandler) PromoteUser(w http.ResponseWriter, r *http.Request) {
	// TODO: CSRF protection
	ctx := r.Context()
	name := r.PathValue("id")

	_, err := uh.userService.Promote(ctx, name)
	if err != nil {
		FlashError(w, uh.logger, err, "Failed to promote user.")
	} else {
		FlashMessage(w, uh.logger, "User promoted.")
	}

	http.Redirect(w, r, DefaultRedirectLocation, http.StatusSeeOther)
}

// DemoteUser demotes logMsgWithArgs user from admin.
func (uh *UserHandler) DemoteUser(w http.ResponseWriter, r *http.Request) {
	// TODO: CSRF protection
	ctx := r.Context()
	name := r.PathValue("id")

	_, err := uh.userService.Demote(ctx, name)
	if err != nil {
		FlashError(w, uh.logger, err, "Failed to demote user.")
	} else {
		FlashMessage(w, uh.logger, "User demoted.")
	}

	http.Redirect(w, r, DefaultRedirectLocation, http.StatusSeeOther)
}

// DeleteUser deletes logMsgWithArgs user.
func (uh *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// TODO: CSRF protection
	ctx := r.Context()
	name := r.PathValue("id")

	err := uh.userService.Delete(ctx, name)
	if err != nil {
		FlashError(w, uh.logger, err, "Failed to delete user.")
	} else {
		FlashMessage(w, uh.logger, "User deleted.")
	}

	http.Redirect(w, r, DefaultRedirectLocation, http.StatusSeeOther)
}
