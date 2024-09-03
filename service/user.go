package service

import (
	"context"
	"fmt"

	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/repo"
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
func NewUser(userRepo UserRepo, sessionRepo SessionRepo, passwordHasher PasswordHasher, passwordChecker PasswordChecker, logger log.Logger) *User {
	return &User{
		logger:          logger,
		repo:            userRepo,
		sessionRepo:     sessionRepo,
		passwordHasher:  passwordHasher,
		passwordChecker: passwordChecker,
	}
}

// Create creates a new user.
// It hashes the password and stores the user in the repository.
// It also checks if the raw password is OK.
func (u *User) Create(ctx context.Context, name, email, password string, isAdmin bool, access []string) (repo.UserModel, error) {
	hash, err := u.HashPassword(ctx, password)
	if err != nil {
		return repo.UserModel{}, err
	}

	userModel, err := u.repo.Create(ctx, name, email, hash, isAdmin, access)
	if err != nil {
		return repo.UserModel{}, fmt.Errorf("failed to create user: %w", err)
	}

	return userModel, nil
}

// Login logs in a user with the given username and password and returns a session hash.
func (u *User) Login(ctx context.Context, userName, password string) (repo.SessionModel, error) {
	user, err := u.repo.Get(ctx, userName)
	if err != nil {
		return repo.SessionModel{}, fmt.Errorf("failed to retrieve user: %w", err)
	}

	// CheckPassword if the password matches
	err = u.passwordHasher.Check(ctx, password, user.Password)
	if err != nil {
		u.logger.Info().Msg("Password retrieved: " + user.Password)

		return repo.SessionModel{}, fmt.Errorf("password does not match: %w", err)
	}

	// Start a session
	hash, err := u.sessionRepo.Start(ctx, userName, user.IsAdmin, user.Access)
	if err != nil {
		return repo.SessionModel{}, fmt.Errorf("failed to start session: %w", err)
	}

	return hash, nil
}

// CheckPassword checks if the given username and password are correct.
func (u *User) CheckPassword(ctx context.Context, userName, password string) error {
	// Retrieve the user
	user, err := u.repo.Get(ctx, userName)
	if err != nil {
		return fmt.Errorf("failed to retrieve user: %w", err)
	}

	// CheckPassword if the password matches
	err = u.passwordHasher.Check(ctx, password, user.Password)
	if err != nil {
		u.logger.Info().Msg("Password retrieved: " + user.Password)

		return fmt.Errorf("failed to check password: %w", err)
	}

	return nil
}

// HashPassword hashes a given password.
func (u *User) HashPassword(ctx context.Context, password string) (string, error) {
	err := u.passwordChecker.IsOK(ctx, password)
	if err != nil {
		return "", fmt.Errorf("password is not OK: %w", err)
	}

	hash, err := u.passwordHasher.Hash(ctx, password)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return hash, nil
}

// CheckPasswordHash checks if the given hash is a valid hash for a given password.
func (u *User) CheckPasswordHash(ctx context.Context, password, hash string) error {
	err := u.passwordHasher.Check(ctx, password, hash)
	if err != nil {
		return fmt.Errorf("failed to check password: %w", err)
	}

	return nil
}
