package repo

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/peteraba/cloudy-files/apperr"
)

// UserModel represents a user model.
type UserModel struct {
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Password string   `json:"password"`
	IsAdmin  bool     `json:"is_admin"`
	Access   []string `json:"access"`
}

// User represents a user.
type User struct {
	store   Store
	lock    *sync.Mutex
	entries map[string]UserModel
}

// NewUser creates a new user instance.
func NewUser(store Store) *User {
	return &User{
		store:   store,
		lock:    &sync.Mutex{},
		entries: make(map[string]UserModel),
	}
}

// createEntries creates entries from data retrieved from store.
func (u *User) createEntries(data []byte) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	entries := make(map[string]UserModel)

	if len(data) > 0 {
		err := json.Unmarshal(data, &entries)
		if err != nil {
			return fmt.Errorf("error unmarshaling data: %w", err)
		}
	}

	u.entries = entries

	return nil
}

// getData returns the data to be stored in the store.
func (u *User) getData() ([]byte, error) {
	data, err := json.Marshal(u.entries)
	if err != nil {
		return nil, fmt.Errorf("error marshaling data: %w", err)
	}

	return data, nil
}

// Get retrieves a user by name.
func (u *User) Get(name string) (UserModel, error) {
	err := u.read()
	if err != nil {
		return UserModel{}, err
	}

	u.lock.Lock()
	defer u.lock.Unlock()

	entry, ok := u.entries[name]
	if !ok {
		return UserModel{}, fmt.Errorf("user not found: %s, err: %w", name, apperr.ErrNotFound)
	}

	return entry, nil
}

// Create creates a new user.
func (u *User) Create(name, email, password string, isAdmin bool, access []string) error {
	err := u.readForWrite()
	if err != nil {
		return err
	}
	defer u.store.Unlock()

	u.lock.Lock()
	defer u.lock.Unlock()

	_, ok := u.entries[name]
	if ok {
		return fmt.Errorf("user already exists: %s, err: %w", name, apperr.ErrExists)
	}

	u.entries[name] = UserModel{
		Email:    email,
		Name:     name,
		Access:   access,
		Password: password,
		IsAdmin:  isAdmin,
	}

	err = u.writeAfterRead()
	if err != nil {
		return err
	}

	return nil
}

// read reads the session data from the store and creates entries.
func (u *User) read() error {
	data, err := u.store.Read()
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	err = u.createEntries(data)
	if err != nil {
		return fmt.Errorf("error creating entries: %w", err)
	}

	return nil
}

// readForWrite reads the session data from the store and creates entries.
// IMPORTANT!!! Do not forget to unlock the store after writing!
// Note: This function assumes that the store is NOT locked!
func (u *User) readForWrite() error {
	data, err := u.store.ReadForWrite()
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	err = u.createEntries(data)
	if err != nil {
		return fmt.Errorf("error creating entries: %w", err)
	}

	return nil
}

// writeAfterRead writes the current session data to the store.
// Note: This function assumes that the store is locked.
func (u *User) writeAfterRead() error {
	data, err := u.getData()
	if err != nil {
		return fmt.Errorf("error getting data: %w", err)
	}

	err = u.store.WriteLocked(data)
	if err != nil {
		return fmt.Errorf("error storing data: %w", err)
	}

	return nil
}
