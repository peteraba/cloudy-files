package password

import (
	"context"

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

	return password, nil
}

// Check checks if the provided password is correct or not.
func (b DummyHasher) Check(_ context.Context, password, hashedPassword string) error {
	if err := b.spy.GetError("Check", password, hashedPassword); err != nil {
		return err
	}

	return nil
}
