package web

import (
	"fmt"
	"net/http"

	"github.com/phuslu/log"
)

type FallbackHandler struct {
	logger *log.Logger
}

func NewFallbackHandler(logger *log.Logger) *FallbackHandler {
	return &FallbackHandler{
		logger: logger,
	}
}

func (fh *FallbackHandler) Home(w http.ResponseWriter) {
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

	send(w, tmpl, fh.logger)
}
