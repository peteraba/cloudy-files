package service

import (
	"context"
	"fmt"

	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/repo"
)

// Session represents a session service.
type Session struct {
	repo   SessionRepo
	logger log.Logger
}

// NewSession creates a new session instance.
func NewSession(sessionRepo SessionRepo, logger log.Logger) *Session {
	return &Session{
		repo:   sessionRepo,
		logger: logger,
	}
}

// Check checks if a session is valid.
func (s *Session) Check(ctx context.Context, name, hash string) (bool, error) {
	ok, err := s.repo.Check(ctx, name, hash)
	if err != nil {
		return false, fmt.Errorf("failed to check session. name: %s, hash: %s: %w", name, hash, err)
	}

	return ok, nil
}

// Get checks if a session is valid.
func (s *Session) Get(ctx context.Context, name, hash string) (repo.SessionModel, error) {
	model, err := s.repo.Get(ctx, name, hash)
	if err != nil {
		return model, fmt.Errorf("failed to check session. name: %s, hash: %s: %w", name, hash, err)
	}

	return model, nil
}

// CleanUp cleans up sessions.
func (s *Session) CleanUp(ctx context.Context) error {
	err := s.repo.CleanUp(ctx)
	if err != nil {
		return fmt.Errorf("failed to clean up sessions: %w", err)
	}

	return nil
}
