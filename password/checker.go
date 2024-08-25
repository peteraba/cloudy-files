package password

import (
	"errors"
	"fmt"

	passwordValidator "github.com/wagslane/go-password-validator"
)

// Checker is a struct that checks if a password is good enough.
type Checker struct {
	minimumEntropy float64
}

var defaultMinimumEntropy = 60.0

// NewChecker creates a new Checker.
func NewChecker() *Checker {
	return &Checker{minimumEntropy: defaultMinimumEntropy}
}

// ErrPwnedPassword is returned when the password is in the pwned password database.
var ErrPwnedPassword = errors.New("password is pwned")

// IsOK checks if the password is strong enough and not in the pwned password database.
func (p Checker) IsOK(password string) error {
	if p.IsPwned() {
		return ErrPwnedPassword
	}

	return p.IsStrongEnough(password)
}

// IsStrongEnough checks if the password is strong enough.
func (p Checker) IsStrongEnough(password string) error {
	err := passwordValidator.Validate(password, p.minimumEntropy)
	if err != nil {
		return fmt.Errorf("password is not strong enough: %w", err)
	}

	return nil
}

// IsPwned checks if the password is in the pwned password database.
func (p Checker) IsPwned() bool {
	//nolint:godox // It would be a nice feature, but it's considered to be an overkill for now
	// TODO: Implement this function (maybe?)
	return false
}
