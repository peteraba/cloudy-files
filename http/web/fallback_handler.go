package web

import (
	"fmt"
	"net/http"

	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/repo"
	"github.com/peteraba/cloudy-files/util"
)

// FallbackHandler handles fallback requests.
type FallbackHandler struct {
	csrfRepo *repo.CSRF
	logger   *log.Logger
}

// NewFallbackHandler creates a new FallbackHandler.
func NewFallbackHandler(csrfRepo *repo.CSRF, logger *log.Logger) *FallbackHandler {
	return &FallbackHandler{
		csrfRepo: csrfRepo,
		logger:   logger,
	}
}

const tokenLength = 32

// Home handles the home page.
// Does not expect a valid session.
func (fh *FallbackHandler) Home(w http.ResponseWriter, r *http.Request) {
	ipAddress := GetIPAddress(r)

	token, _ := util.RandomHex(tokenLength)

	err := fh.csrfRepo.Create(r.Context(), ipAddress, token)
	if err != nil {
		Problem(w, fh.logger, err)

		return
	}

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
		token,
	)

	Send(w, tmpl)
}
