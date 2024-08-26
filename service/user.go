package service

import (
	"fmt"

	"github.com/phuslu/log"
)

// User is a service that provides user-related operations.
type User struct {
	logger          log.Logger
	repo            UserRepo
	sessionRepo     SessionRepo
	passwordHasher  PasswordHasher
	passwordChecker PasswordChecker
}

// NewUser creates a new User service.
func NewUser(repo UserRepo, sessionRepo SessionRepo, passwordHasher PasswordHasher, passwordChecker PasswordChecker, logger log.Logger) *User {
	return &User{
		logger:          logger,
		repo:            repo,
		sessionRepo:     sessionRepo,
		passwordHasher:  passwordHasher,
		passwordChecker: passwordChecker,
	}
}

// Login logs in a user with the given username and password.
func (u *User) Login(userName, password string) (string, error) {
	user, err := u.repo.Get(userName)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve user: %w", err)
	}

	// CheckPassword if the password matches
	err = u.passwordHasher.Check(password, user.Password)
	if err != nil {
		u.logger.Info().Msg("Password retrieved: " + user.Password)

		return "", fmt.Errorf("password does not match: %w", err)
	}

	// Start a session
	hash, err := u.sessionRepo.Start(userName)
	if err != nil {
		return "", fmt.Errorf("failed to start session: %w", err)
	}

	return hash, nil
}

// CheckPassword checks if the given username and password are correct.
func (u *User) CheckPassword(userName, password string) error {
	// Retrieve the user
	user, err := u.repo.Get(userName)
	if err != nil {
		return fmt.Errorf("failed to retrieve user: %w", err)
	}

	// CheckPassword if the password matches
	err = u.passwordHasher.Check(password, user.Password)
	if err != nil {
		u.logger.Info().Msg("Password retrieved: " + user.Password)

		return fmt.Errorf("failed to check password: %w", err)
	}

	return nil
}

// HashPassword hashes the given password.
func (u *User) HashPassword(password string) (string, error) {
	err := u.passwordChecker.IsOK(password)
	if err != nil {
		return "", fmt.Errorf("password is not OK: %w", err)
	}

	hash, err := u.passwordHasher.Hash(password)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return hash, nil
}

// CheckPasswordHash checks if the given hash is a valid has for the given password.
func (u *User) CheckPasswordHash(password, hash string) error {
	err := u.passwordHasher.Check(password, hash)
	if err != nil {
		return fmt.Errorf("failed to check password: %w", err)
	}

	return nil
}

// Create creates a new user.
// It hashes the password and stores the user in the repository.
// It also checks if the raw password is OK.
func (u *User) Create(name, email, password string, isAdmin bool, access []string) error {
	hash, err := u.HashPassword(password)
	if err != nil {
		return err
	}

	err = u.repo.Create(name, email, hash, isAdmin, access)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}
