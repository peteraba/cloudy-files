package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/apperr"
	"github.com/peteraba/cloudy-files/repo"
)

// Cookie represents a cookie service.
type Cookie struct {
	cookieStore         *securecookie.SecureCookie
	logger              log.Logger
	userKey             string
	flashKey            string
	userCookieLifespan  time.Duration
	flashCookieLifespan time.Duration
}

const (
	day = 24 * time.Hour
)

// NewCookie creates a new Cookie service.
func NewCookie(cookieStore *securecookie.SecureCookie, logger log.Logger) *Cookie {
	return &Cookie{
		cookieStore:         cookieStore,
		logger:              logger,
		userKey:             "user",
		flashKey:            "flash",
		userCookieLifespan:  day,
		flashCookieLifespan: time.Hour,
	}
}

// FlashError sets up a temporary session message for an error.
func (s *Cookie) FlashError(w http.ResponseWriter, r *http.Request, path string, err error, msg string, args ...interface{}) {
	logMsgWithArgs(s.logger.Error().Err(err), msg, args...)

	flashErr := s.AddFlashMessage(w, r, FlashMessage{Level: LevelError, Message: msg})
	if flashErr != nil {
		s.logger.Error().Err(flashErr).Msg("Failed to add flash message.")
	}

	http.Redirect(w, r, path, http.StatusSeeOther)
}

// FlashMessage sets up a temporary session message.
func (s *Cookie) FlashMessage(w http.ResponseWriter, r *http.Request, path, msg string, args ...interface{}) {
	logMsgWithArgs(s.logger.Info(), msg, args...)

	flashErr := s.AddFlashMessage(w, r, FlashMessage{Level: LevelInfo, Message: msg})
	if flashErr != nil {
		s.logger.Error().Err(flashErr).Msg("Failed to add flash message.")
	}

	http.Redirect(w, r, path, http.StatusSeeOther)
}

func logMsgWithArgs(entry *log.Entry, msg string, args ...interface{}) {
	for i, arg := range args {
		if stringer, ok := arg.(fmt.Stringer); ok {
			entry.Stringer(fmt.Sprintf("arg%d", i), stringer)
		} else {
			entry.Interface(fmt.Sprintf("arg%d", i), arg)
		}
	}

	entry.Msg(msg)
}

type Level string

const (
	LevelInfo  Level = "info"
	LevelError Level = "error"
)

// FlashMessage represents a message to be displayed to the user.
type FlashMessage struct {
	Level   Level
	Message string
}

// GetFlashMessages retrieves a slice of FlashMessage objects from a cookie and deletes it.
func (s *Cookie) GetFlashMessages(w http.ResponseWriter, r *http.Request) ([]FlashMessage, error) {
	flashMessages, err := s.getFlashMessages(r)
	if err != nil || len(flashMessages) == 0 {
		return nil, err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     s.flashKey,
		Value:    "",
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
	})

	return flashMessages, nil
}

func (s *Cookie) getFlashMessages(r *http.Request) ([]FlashMessage, error) {
	cookie, err := r.Cookie(s.flashKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve flash messages, err: %w", err)
	}

	if cookie.Value == "" {
		return nil, nil
	}

	var flashMessages []FlashMessage

	err = s.cookieStore.Decode(s.flashKey, cookie.Value, &flashMessages)
	if err != nil {
		return nil, fmt.Errorf("failed to decode flash message session, err: %w", err)
	}

	return flashMessages, nil
}

// AddFlashMessage stores a new FlashMessage in a cookie.
func (s *Cookie) AddFlashMessage(w http.ResponseWriter, r *http.Request, flashMessage FlashMessage) error {
	flashMessages, err := s.getFlashMessages(r)
	if err != nil {
		flashMessages = []FlashMessage{}
	}

	flashMessages = append(flashMessages, flashMessage)

	encoded, _ := s.cookieStore.Encode(s.flashKey, flashMessages)

	cookie := &http.Cookie{
		Name:     s.flashKey,
		Value:    encoded,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		Expires:  time.Now().Add(s.flashCookieLifespan),
	}

	http.SetCookie(w, cookie)

	return nil
}

// GetSessionUser retrieves a SessionUser from a cookie.
func (s *Cookie) GetSessionUser(r *http.Request) (repo.SessionUser, error) {
	c, err := r.Cookie(s.userKey)
	if err != nil || c.Value == "" {
		return repo.SessionUser{}, fmt.Errorf("no session user, err: %w", apperr.ErrAccessDenied)
	}

	userSession := repo.SessionUser{}

	err = s.cookieStore.Decode(s.userKey, c.Value, &userSession)
	if err != nil {
		return repo.SessionUser{}, fmt.Errorf("failed to decode user session, err: %w", err)
	}

	if userSession.Name == "" {
		return repo.SessionUser{}, fmt.Errorf("no session user, err: %w", apperr.ErrAccessDenied)
	}

	return userSession, nil
}

// StoreSessionUser stores a SessionUser in a cookie.
func (s *Cookie) StoreSessionUser(w http.ResponseWriter, sessionUser repo.SessionUser) {
	encoded, _ := s.cookieStore.Encode(s.userKey, sessionUser)

	cookie := &http.Cookie{
		Name:     s.userKey,
		Value:    encoded,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		Expires:  time.Now().Add(s.userCookieLifespan),
	}

	http.SetCookie(w, cookie)
}

// DeleteSessionUser deletes a SessionUser from a cookie.
func (s *Cookie) DeleteSessionUser(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     s.userKey,
		Value:    "",
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
	}

	http.SetCookie(w, cookie)
}
