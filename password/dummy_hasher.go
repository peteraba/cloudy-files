package password

import (
	"context"
	"fmt"
	"slices"

	"github.com/peteraba/cloudy-files/apperr"
	"github.com/peteraba/cloudy-files/util"
)

// DummyHasher is a dummy password hashing and checking implementation.
type DummyHasher struct {
	spy *util.Spy
}

// NewDummyHasher returns a new DummyHasher instance.
func NewDummyHasher(spy *util.Spy) *DummyHasher {
	return &DummyHasher{
		spy: spy,
	}
}

// Hash returns the bcrypt hash of the password.
func (b DummyHasher) Hash(_ context.Context, password string) (string, error) {
	if err := b.spy.GetError("Hash", password); err != nil {
		return "", err
	}

	r := []rune(password)
	slices.Reverse(r)

	return string(r), nil
}

// Check checks if the provided password is correct or not.
func (b DummyHasher) Check(_ context.Context, password, hashedPassword string) error {
	if err := b.spy.GetError("Check", password, hashedPassword); err != nil {
		return err
	}

	r := []rune(password)
	slices.Reverse(r)

	if string(r) != hashedPassword {
		return fmt.Errorf("password is incorrect: %w", apperr.ErrPasswordTooLong)
	}

	return nil
}
