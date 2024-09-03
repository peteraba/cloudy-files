package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/peteraba/cloudy-files/apperr"
	"github.com/peteraba/cloudy-files/util"
)

var defaultSessionTime = time.Minute * 30

// SessionModel represents a session model.
type SessionModel struct {
	Hash    string   `json:"hash"`
	IsAdmin bool     `json:"is_admin"`
	Expires int64    `json:"expires"`
	Access  []string `json:"access"`
}

// Store represents a session store.
type Store interface {
	Read(ctx context.Context) ([]byte, error)
	ReadForWrite(ctx context.Context) ([]byte, error)
	WriteLocked(ctx context.Context, data []byte) error
	Unlock(ctx context.Context) error
	Write(ctx context.Context, data []byte) error
}

// Session represents a session.
type Session struct {
	store   Store
	lock    *sync.Mutex
	entries map[string]SessionModel
}

// NewSession creates a new session instance.
func NewSession(store Store) *Session {
	return &Session{
		store:   store,
		lock:    &sync.Mutex{},
		entries: make(map[string]SessionModel),
	}
}

// createEntries creates entries from data retrieved from store.
func (s *Session) createEntries(data []byte) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	entries := make(map[string]SessionModel)

	if len(data) > 0 {
		err := json.Unmarshal(data, &entries)
		if err != nil {
			return fmt.Errorf("error unmarshaling data: %w", err)
		}
	}

	s.entries = entries

	return nil
}

// getData returns the data to be stored in the store.
func (s *Session) getData() ([]byte, error) {
	data, err := json.Marshal(s.entries)
	if err != nil {
		return nil, fmt.Errorf("error marshaling data: %w", err)
	}

	return data, nil
}

// Get gets the session.
func (s *Session) Get(ctx context.Context, name, hash string) (SessionModel, error) {
	data, err := s.store.Read(ctx)
	if err != nil {
		return SessionModel{}, fmt.Errorf("error reading file: %w", err)
	}

	err = s.createEntries(data)
	if err != nil {
		return SessionModel{}, fmt.Errorf("error creating entries: %w", err)
	}

	storedValue, ok := s.entries[name]
	if !ok {
		return SessionModel{}, nil //nolint:exhaustruct // Checked
	}

	now := time.Now().Unix()
	if storedValue.Expires < now {
		return SessionModel{}, nil //nolint:exhaustruct // Checked
	}

	if storedValue.Hash != hash {
		return SessionModel{}, apperr.ErrAccessDenied
	}

	return storedValue, nil
}

// Check checks if the session is valid.
func (s *Session) Check(ctx context.Context, name, hash string) (bool, error) {
	storedValue, err := s.Get(ctx, name, hash)
	if err != nil {
		return false, fmt.Errorf("error getting session: %w", err)
	}

	return storedValue.Hash == hash, nil
}

const hashLength = 32

// Start starts a new session.
func (s *Session) Start(ctx context.Context, name string, isAdmin bool, access []string) (SessionModel, error) {
	err := s.readForWrite(ctx)
	if err != nil {
		return SessionModel{}, err
	}
	defer s.store.Unlock(ctx)

	hash, err := util.RandomHex(hashLength)
	if err != nil {
		return SessionModel{}, fmt.Errorf("error generating random hash: %w", err)
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.entries[name] = SessionModel{
		Hash:    hash,
		Expires: time.Now().Add(defaultSessionTime).Unix(),
		IsAdmin: isAdmin,
		Access:  access,
	}

	err = s.writeAfterRead(ctx)
	if err != nil {
		return SessionModel{}, err
	}

	return s.entries[name], nil
}

// CleanUp cleans up the session.
func (s *Session) CleanUp(ctx context.Context) error {
	err := s.readForWrite(ctx)
	if err != nil {
		return err
	}
	defer s.store.Unlock(ctx)

	s.lock.Lock()
	defer s.lock.Unlock()

	count := 0

	for key, entry := range s.entries {
		if entry.Expires < time.Now().Unix() {
			delete(s.entries, key)

			count++
		}
	}

	if count == 0 {
		return nil
	}

	err = s.writeAfterRead(ctx)
	if err != nil {
		return err
	}

	return nil
}

// readForWrite reads the session data from the store and creates entries.
// IMPORTANT!!! Do not forget to unlock the store after writing!
// Note: This function assumes that the store is NOT locked!
func (s *Session) readForWrite(ctx context.Context) error {
	data, err := s.store.ReadForWrite(ctx)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	err = s.createEntries(data)
	if err != nil {
		return fmt.Errorf("error creating entries: %w", err)
	}

	return nil
}

// writeAfterRead writes the current session data to the store.
// Note: This function assumes that the store is locked.
func (s *Session) writeAfterRead(ctx context.Context) error {
	data, err := s.getData()
	if err != nil {
		return fmt.Errorf("error getting data: %w", err)
	}

	err = s.store.WriteLocked(ctx, data)
	if err != nil {
		return fmt.Errorf("error storing data: %w", err)
	}

	return nil
}
