package password

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// BcryptHasher is a password hashing and checking implementation using bcrypt.
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher returns a new BcryptHasher instance.
func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{cost: bcrypt.DefaultCost}
}

func NewBcryptWithCost(cost int) *BcryptHasher {
	return &BcryptHasher{cost: cost}
}

// Hash returns the bcrypt hash of the password.
func (b BcryptHasher) Hash(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedPassword), nil
}

// Check checks if the provided password is correct or not.
func (b BcryptHasher) Check(password, hashedPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return fmt.Errorf("password is incorrect: %w", err)
	}

	return nil
}
