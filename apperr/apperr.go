package apperr

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// ErrAccessDenied represents an access denied error.
var ErrAccessDenied = errors.New("access denied")

// ErrNotFound is returned when a resource cannot be found.
var ErrNotFound = errors.New("not found")

// ErrExists is returned when a resource already exists.
var ErrExists = errors.New("already exists")

// ErrLockTimeout is returned when a lock cannot be acquired.
var ErrLockTimeout = errors.New("lock timeout")

// ErrPasswordTooLong is returned when the password is too long.
var ErrPasswordTooLong = errors.New("password is too long")

// ErrInvalidArgument is returned when a method is called with an invalid argument.
var ErrInvalidArgument = errors.New("invalid argument")

// ErrNotImplemented is returned when a method is not implemented.
var ErrNotImplemented = errors.New("not implemented")

// HTTPError represents a JSON error response.
type HTTPError struct {
	Type   string `json:"type,omitempty"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
}

// Error returns the error message.
func (e *HTTPError) Error() string {
	return fmt.Sprintf("%d %s %s", e.Status, e.Title, e.Detail)
}

// GetHTTPError returns an error response.
func GetHTTPError(err error) *HTTPError {
	detail := strings.TrimRight(strings.Title(err.Error()), ".") + "." //nolint:staticcheck // No need for unicode punctuation

	if errors.Is(err, ErrAccessDenied) {
		return &HTTPError{
			Type:   "",
			Title:  "Access denied",
			Status: http.StatusForbidden,
			Detail: detail,
		}
	}

	if errors.Is(err, ErrNotFound) {
		return &HTTPError{
			Type:   "",
			Title:  "Not found",
			Status: http.StatusNotFound,
			Detail: detail,
		}
	}

	if errors.Is(err, ErrNotImplemented) {
		return &HTTPError{
			Type:   "",
			Title:  "Not found",
			Status: http.StatusNotFound,
			Detail: detail,
		}
	}

	return &HTTPError{
		Type:   "",
		Title:  "Internal error",
		Status: http.StatusInternalServerError,
		Detail: detail,
	}
}
