package password

import (
	"fmt"

	passwordValidator "github.com/wagslane/go-password-validator"

	"github.com/peteraba/cloudy-files/apperr"
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

// NewCheckerWithEntropy creates a new Checker with a custom minimum entropy.
func NewCheckerWithEntropy(minimumEntropy float64) *Checker {
	return &Checker{minimumEntropy: minimumEntropy}
}

const bcryptPasswordMaxLength = 72

// IsOK checks if the password is strong enough and not in the pwned password database.
func (p Checker) IsOK(password string) error {
	// password length is checked as []byte to avoid issues with multibyte characters
	if len([]byte(password)) > bcryptPasswordMaxLength {
		return apperr.ErrPasswordTooLong
	}

	return p.isStrongEnough(password)
}

// isStrongEnough checks if the password is strong enough.
func (p Checker) isStrongEnough(password string) error {
	err := passwordValidator.Validate(password, p.minimumEntropy)
	if err != nil {
		return fmt.Errorf("password is not strong enough: %w", err)
	}

	return nil
}
