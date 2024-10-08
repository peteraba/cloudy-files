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
	passwordHasher  PasswordHasher
	passwordChecker PasswordChecker
}

// NewUser creates a new User service.
func NewUser(userRepo UserRepo, passwordHasher PasswordHasher, passwordChecker PasswordChecker, logger log.Logger) *User {
	return &User{
		logger:          logger,
		repo:            userRepo,
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
func (u *User) Login(ctx context.Context, userName, password string) (repo.SessionUser, error) {
	user, err := u.repo.Get(ctx, userName)
	if err != nil {
		return repo.SessionUser{}, fmt.Errorf("failed to retrieve user: %w", err)
	}

	// CheckPassword if the password matches
	err = u.passwordHasher.Check(ctx, password, user.Password)
	if err != nil {
		u.logger.Info().Msg("Password retrieved: " + user.Password)

		return repo.SessionUser{}, fmt.Errorf("password does not match: %w", err)
	}

	return user.ToSession(), nil
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

// List lists all users.
func (u *User) List(ctx context.Context) (repo.UserModels, error) {
	list, err := u.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return list, nil
}

// UpdatePassword updates the password of a user.
func (u *User) UpdatePassword(ctx context.Context, name, password string) (repo.UserModel, error) {
	hash, err := u.HashPassword(ctx, password)
	if err != nil {
		return repo.UserModel{}, fmt.Errorf("failed to hash password: %w", err)
	}

	userModel, err := u.repo.UpdatePassword(ctx, name, hash)
	if err != nil {
		return repo.UserModel{}, fmt.Errorf("failed to update password: %w", err)
	}

	return userModel, nil
}

// UpdateAccess updates the access of a user.
func (u *User) UpdateAccess(ctx context.Context, name string, access []string) (repo.UserModel, error) {
	userModel, err := u.repo.UpdateAccess(ctx, name, access)
	if err != nil {
		return repo.UserModel{}, fmt.Errorf("failed to update access: %w", err)
	}

	return userModel, nil
}

// Promote promotes a user to an admin.
func (u *User) Promote(ctx context.Context, name string) (repo.UserModel, error) {
	userModel, err := u.repo.Promote(ctx, name)
	if err != nil {
		return repo.UserModel{}, fmt.Errorf("failed to promote user: %w", err)
	}

	return userModel, nil
}

// Demote demotes a user from an admin.
func (u *User) Demote(ctx context.Context, name string) (repo.UserModel, error) {
	userModel, err := u.repo.Demote(ctx, name)
	if err != nil {
		return repo.UserModel{}, fmt.Errorf("failed to demote user: %w", err)
	}

	return userModel, nil
}

// Delete deletes a user.
func (u *User) Delete(ctx context.Context, name string) error {
	err := u.repo.Delete(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}
