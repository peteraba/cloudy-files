package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/peteraba/cloudy-files/apperr"
)

// SessionUser represents a user as stored in a session.
type SessionUser struct {
	Name    string   `json:"name"     formam:"name"`
	IsAdmin bool     `json:"is_admin" formam:"is_admin"`
	Access  []string `json:"access"   formam:"access"`
}

// UserModel represents a user model.
type UserModel struct {
	Name     string   `json:"name"     formam:"name"`
	Email    string   `json:"email"    formam:"email"`
	Password string   `json:"password" formam:"password"`
	IsAdmin  bool     `json:"is_admin" formam:"is_admin"`
	Access   []string `json:"access"   formam:"access"`
}

// ToSession converts a user model to a session model.
func (u UserModel) ToSession() SessionUser { //nolint:gocritic // Models are not to be passed as a pointers
	return SessionUser{
		Name:    u.Name,
		IsAdmin: u.IsAdmin,
		Access:  u.Access,
	}
}

// UserModels represents a user model list.
type UserModels []UserModel

// UserModelMap represents a user model map.
type UserModelMap map[string]UserModel

// Slice returns the user models as a slice.
func (u UserModelMap) Slice() UserModels {
	users := UserModels{}

	for _, user := range u {
		users = append(users, user)
	}

	return users
}

// User represents a user.
type User struct {
	store   Store
	lock    *sync.Mutex
	entries UserModelMap
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

// List lists all users.
func (u *User) List(ctx context.Context) (UserModels, error) {
	err := u.read(ctx)
	if err != nil {
		return nil, err
	}

	u.lock.Lock()
	defer u.lock.Unlock()

	return u.entries.Slice(), nil
}

// Get retrieves a user by name.
func (u *User) Get(ctx context.Context, name string) (UserModel, error) {
	err := u.read(ctx)
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
func (u *User) Create(ctx context.Context, name, email, password string, isAdmin bool, access []string) (UserModel, error) {
	err := u.readForWrite(ctx)
	if err != nil {
		return UserModel{}, err
	}
	defer u.store.Unlock(ctx)

	u.lock.Lock()
	defer u.lock.Unlock()

	_, ok := u.entries[name]
	if ok {
		return UserModel{}, fmt.Errorf("user already exists: %s, err: %w", name, apperr.ErrExists)
	}

	u.entries[name] = UserModel{
		Email:    email,
		Name:     name,
		Access:   access,
		Password: password,
		IsAdmin:  isAdmin,
	}

	err = u.writeAfterRead(ctx)
	if err != nil {
		return UserModel{}, err
	}

	return u.entries[name], nil
}

// UpdatePassword updates the password of a user.
func (u *User) UpdatePassword(ctx context.Context, name, password string) (UserModel, error) {
	err := u.readForWrite(ctx)
	if err != nil {
		return UserModel{}, err
	}
	defer u.store.Unlock(ctx)

	u.lock.Lock()
	defer u.lock.Unlock()

	entry, ok := u.entries[name]
	if !ok {
		return UserModel{}, fmt.Errorf("user not found: %s, err: %w", name, apperr.ErrNotFound)
	}

	entry.Password = password

	u.entries[name] = entry

	err = u.writeAfterRead(ctx)
	if err != nil {
		return UserModel{}, err
	}

	return entry, nil
}

// UpdateAccess updates the access of a user.
func (u *User) UpdateAccess(ctx context.Context, name string, access []string) (UserModel, error) {
	err := u.readForWrite(ctx)
	if err != nil {
		return UserModel{}, err
	}
	defer u.store.Unlock(ctx)

	u.lock.Lock()
	defer u.lock.Unlock()

	entry, ok := u.entries[name]
	if !ok {
		return UserModel{}, fmt.Errorf("user not found: %s, err: %w", name, apperr.ErrNotFound)
	}

	entry.Access = access

	u.entries[name] = entry

	err = u.writeAfterRead(ctx)
	if err != nil {
		return UserModel{}, err
	}

	return entry, nil
}

// Promote promotes a user to admin.
func (u *User) Promote(ctx context.Context, name string) (UserModel, error) {
	err := u.readForWrite(ctx)
	if err != nil {
		return UserModel{}, err
	}
	defer u.store.Unlock(ctx)

	u.lock.Lock()
	defer u.lock.Unlock()

	entry, ok := u.entries[name]
	if !ok {
		return UserModel{}, fmt.Errorf("user not found: %s, err: %w", name, apperr.ErrNotFound)
	}

	if entry.IsAdmin {
		return UserModel{}, apperr.ErrValidation("user is already an admin")
	}

	entry.IsAdmin = true

	u.entries[name] = entry

	err = u.writeAfterRead(ctx)
	if err != nil {
		return UserModel{}, err
	}

	return entry, nil
}

// Demote demotes an admin to user.
func (u *User) Demote(ctx context.Context, name string) (UserModel, error) {
	err := u.readForWrite(ctx)
	if err != nil {
		return UserModel{}, err
	}
	defer u.store.Unlock(ctx)

	u.lock.Lock()
	defer u.lock.Unlock()

	entry, ok := u.entries[name]
	if !ok {
		return UserModel{}, fmt.Errorf("user not found: %s, err: %w", name, apperr.ErrNotFound)
	}

	if !entry.IsAdmin {
		return UserModel{}, apperr.ErrValidation("user is not an admin")
	}

	entry.IsAdmin = false

	u.entries[name] = entry

	err = u.writeAfterRead(ctx)
	if err != nil {
		return UserModel{}, err
	}

	return entry, nil
}

// Delete deletes a user.
func (u *User) Delete(ctx context.Context, name string) error {
	err := u.readForWrite(ctx)
	if err != nil {
		return fmt.Errorf("error reading for write: %w", err)
	}
	defer u.store.Unlock(ctx)

	u.lock.Lock()
	defer u.lock.Unlock()

	delete(u.entries, name)

	err = u.writeAfterRead(ctx)
	if err != nil {
		return fmt.Errorf("error writing after read: %w", err)
	}

	return nil
}

// read reads the session data from the store and creates entries.
func (u *User) read(ctx context.Context) error {
	data, err := u.store.Read(ctx)
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
func (u *User) readForWrite(ctx context.Context) error {
	data, err := u.store.ReadForWrite(ctx)
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
func (u *User) writeAfterRead(ctx context.Context) error {
	data, _ := json.Marshal(u.entries) //nolint:errchkjson // We are sure that the data can be marshaled correctly

	err := u.store.WriteLocked(ctx, data)
	if err != nil {
		return fmt.Errorf("error storing data: %w", err)
	}

	return nil
}
