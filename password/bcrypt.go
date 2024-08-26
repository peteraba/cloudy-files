package password

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Bcrypt is a password hashing and checking implementation using bcrypt.
type Bcrypt struct {
	cost int
}

// NewBcrypt returns a new Bcrypt instance.
func NewBcrypt() *Bcrypt {
	return &Bcrypt{cost: bcrypt.DefaultCost}
}

func NewBcryptWithCost(cost int) *Bcrypt {
	return &Bcrypt{cost: cost}
}

// Hash returns the bcrypt hash of the password.
func (b Bcrypt) Hash(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedPassword), nil
}

// Check checks if the provided password is correct or not.
func (b Bcrypt) Check(password, hashedPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return fmt.Errorf("password is incorrect: %w", err)
	}

	return nil
}
