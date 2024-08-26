package apperr

import "errors"

// ErrAccessDenied represents an access denied error.
var ErrAccessDenied = errors.New("access denied")

// ErrNotFound is returned when a resource cannot be found.
var ErrNotFound = errors.New("not found")

// ErrExists is returned when a resource already exists.
var ErrExists = errors.New("already exists")

// ErrLockTimeout is returned when a lock cannot be acquired.
var ErrLockTimeout = errors.New("lock timeout")

// ErrPwnedPassword is returned when the password is in the pwned password database.
var ErrPwnedPassword = errors.New("password is pwned")

// ErrPasswordTooLong is returned when the password is too long.
var ErrPasswordTooLong = errors.New("password is too long")

// ErrInvalidArgument is returned when a method is called with an invalid argument.
var ErrInvalidArgument = errors.New("invalid argument")
