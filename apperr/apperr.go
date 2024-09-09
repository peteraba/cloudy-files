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

// errBadRequest is returned when a method is not implemented.
var errBadRequest = errors.New("bad request")

// ErrLockDoesNotExist is returned when a lock does not exist.
var ErrLockDoesNotExist = errors.New("lock not locked")

// ErrBadRequest returns a bad request error.
func ErrBadRequest(err error) error {
	return fmt.Errorf("%s, err: %w", err.Error(), errBadRequest)
}

// ErrEmptyForm is returned when a form is empty that shouldn't be.
var ErrEmptyForm = fmt.Errorf("form is empty, err: %w", errBadRequest)

// ErrValidation returns a validation error.
func ErrValidation(msg string) error {
	return fmt.Errorf("%s, err: %w", msg, errBadRequest)
}

// Problem represents a JSON error response.
type Problem struct { //nolint:errname // This is Zalando standard naming
	Type   string `json:"type,omitempty"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
}

// Error returns the error message.
func (e *Problem) Error() string {
	return fmt.Sprintf("%d %s %s", e.Status, e.Title, e.Detail)
}

// GetProblem returns an error response.
func GetProblem(err error) *Problem {
	firstPart := err.Error()
	if idx := strings.Index(firstPart, ":"); idx > 0 {
		firstPart = firstPart[:idx]
	}

	detail := strings.TrimRight(strings.Title(firstPart), ".") + "." //nolint:staticcheck // No need for unicode punctuation

	if errors.Is(err, ErrAccessDenied) {
		return &Problem{
			Type:   "",
			Title:  "Access denied",
			Status: http.StatusForbidden,
			Detail: detail,
		}
	}

	if errors.Is(err, ErrNotFound) {
		return &Problem{
			Type:   "",
			Title:  "Not found",
			Status: http.StatusNotFound,
			Detail: detail,
		}
	}

	if errors.Is(err, ErrNotImplemented) {
		return &Problem{
			Type:   "",
			Title:  "Not implemented",
			Status: http.StatusNotFound,
			Detail: detail,
		}
	}

	if errors.Is(err, errBadRequest) {
		return &Problem{
			Type:   "",
			Title:  "Bad request",
			Status: http.StatusBadRequest,
			Detail: detail,
		}
	}

	return &Problem{
		Type:   "",
		Title:  "Internal error",
		Status: http.StatusInternalServerError,
		Detail: detail,
	}
}
