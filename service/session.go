package service

import (
	"fmt"

	"github.com/phuslu/log"
)

type Session struct {
	repo   sessionRepo
	logger log.Logger
}

func NewSession(repo sessionRepo, logger log.Logger) *Session {
	return &Session{
		repo:   repo,
		logger: logger,
	}
}

func (s *Session) Check(name, hash string) (bool, error) {
	ok, err := s.repo.Check(name, hash)
	if err != nil {
		return false, fmt.Errorf("failed to check session. name: %s, hash: %s: %w", name, hash, err)
	}

	return ok, nil
}

func (s *Session) Start(name string) (string, error) {
	hash, err := s.repo.Start(name)
	if err != nil {
		return "", fmt.Errorf("failed to start session: %w", err)
	}

	return hash, nil
}

func (s *Session) CleanUp() error {
	err := s.repo.CleanUp()
	if err != nil {
		return fmt.Errorf("failed to clean up sessions: %w", err)
	}

	return nil
}
