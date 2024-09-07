package service

import (
	"context"
	"fmt"
	"time"

	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/apperr"
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
func (s *Session) Check(ctx context.Context, name, hash string) error {
	_, err := s.Get(ctx, name, hash)
	if err != nil {
		return fmt.Errorf("failed to check session, err: %w", err)
	}

	return nil
}

// Get checks if a session is valid.
func (s *Session) Get(ctx context.Context, name, hash string) (repo.SessionModel, error) {
	session, err := s.repo.Get(ctx, name)
	if err != nil {
		return repo.SessionModel{}, fmt.Errorf("failed to check session. name: %s, err: %w", name, err)
	}

	now := time.Now().Unix()
	if session.Expires < now || session.Hash != hash {
		return repo.SessionModel{}, fmt.Errorf("invalid session. name: %s, err: %w", name, apperr.ErrAccessDenied)
	}

	return session, nil
}

// CleanUp cleans up sessions.
func (s *Session) CleanUp(ctx context.Context) error {
	err := s.repo.CleanUp(ctx)
	if err != nil {
		return fmt.Errorf("failed to clean up sessions: %w", err)
	}

	return nil
}
