package repo

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/peteraba/cloudy-files/util"
)

var defaultSessionTime = time.Minute * 30

// SessionModel represents a session model.
type SessionModel struct {
	Hash    string `json:"hash"`
	Expires int64  `json:"expires"`
}

// Store represents a session store.
type Store interface {
	Read() ([]byte, error)
	ReadForWrite() ([]byte, error)
	WriteLocked(data []byte) error
	Unlock() error
	Write(data []byte) error
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

	err := json.Unmarshal(data, &entries)
	if err != nil {
		return fmt.Errorf("error unmarshaling data: %w", err)
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

// Check checks if the session is valid.
func (s *Session) Check(name, hash string) (bool, error) {
	data, err := s.store.Read()
	if err != nil {
		return false, fmt.Errorf("error reading file: %w", err)
	}

	err = s.createEntries(data)
	if err != nil {
		return false, fmt.Errorf("error creating entries: %w", err)
	}

	storedValue, ok := s.entries[name]
	if !ok {
		return false, nil
	}

	now := time.Now().Unix()
	if storedValue.Expires < now {
		return false, nil
	}

	return storedValue.Hash == hash && storedValue.Expires > now, nil
}

const hashLength = 32

// Start starts a new session.
func (s *Session) Start(name string) (string, error) {
	err := s.readForWrite()
	if err != nil {
		return "", err
	}
	defer s.store.Unlock()

	hash, err := util.RandomHex(hashLength)
	if err != nil {
		return "", fmt.Errorf("error generating random hash: %w", err)
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.entries[name] = SessionModel{
		Hash:    hash,
		Expires: time.Now().Add(defaultSessionTime).Unix(),
	}

	err = s.writeAfterRead()
	if err != nil {
		return "", err
	}

	return hash, nil
}

// CleanUp cleans up the session.
func (s *Session) CleanUp() error {
	err := s.readForWrite()
	if err != nil {
		return err
	}
	defer s.store.Unlock()

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

	err = s.writeAfterRead()
	if err != nil {
		return err
	}

	return nil
}

// readForWrite reads the session data from the store and creates entries.
// IMPORTANT!!! Do not forget to unlock the store after writing!
// Note: This function assumes that the store is NOT locked!
func (s *Session) readForWrite() error {
	data, err := s.store.ReadForWrite()
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
func (s *Session) writeAfterRead() error {
	data, err := s.getData()
	if err != nil {
		return fmt.Errorf("error getting data: %w", err)
	}

	err = s.store.WriteLocked(data)
	if err != nil {
		return fmt.Errorf("error storing data: %w", err)
	}

	return nil
}
