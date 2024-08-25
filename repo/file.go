package repo

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

// FileModel represents a file model.
type FileModel struct {
	Name   string   `json:"name"`
	Access []string `json:"access"`
}

// File represents a file.
type File struct {
	store   Store
	lock    *sync.Mutex
	entries map[string]FileModel
}

// NewFile creates a new file instance.
func NewFile(store Store) *File {
	return &File{
		store:   store,
		lock:    &sync.Mutex{},
		entries: make(map[string]FileModel),
	}
}

// createEntries creates entries from data retrieved from store.
func (f *File) createEntries(data []byte) error {
	f.lock.Lock()
	defer f.lock.Unlock()

	entries := make(map[string]FileModel)

	err := json.Unmarshal(data, &entries)
	if err != nil {
		return fmt.Errorf("error unmarshaling data: %w", err)
	}

	f.entries = entries

	return nil
}

// getData returns the data to be stored in the store.
func (f *File) getData() ([]byte, error) {
	data, err := json.Marshal(f.entries)
	if err != nil {
		return nil, fmt.Errorf("error marshaling data: %w", err)
	}

	return data, nil
}

// ErrFileNotFound represents a file not found error.
var ErrFileNotFound = errors.New("file not found")

// Get retrieves a file by name.
func (f *File) Get(name string) (FileModel, error) {
	err := f.read()
	if err != nil {
		return FileModel{}, fmt.Errorf("error reading file: %w", err)
	}

	f.lock.Lock()
	defer f.lock.Unlock()

	entry, ok := f.entries[name]
	if !ok {
		return FileModel{}, fmt.Errorf("file not found: %s, err: %w", name, ErrFileNotFound)
	}

	return entry, nil
}

// Create creates a file with the given name and access.
func (f *File) Create(name string, access []string) error {
	err := f.readForWrite()
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}
	defer f.store.Unlock()

	f.lock.Lock()
	defer f.lock.Unlock()

	f.entries[name] = FileModel{
		Name:   name,
		Access: access,
	}

	err = f.writeAfterRead()
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

func (f *File) read() error {
	data, err := f.store.Read()
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	err = f.createEntries(data)
	if err != nil {
		return fmt.Errorf("error creating entries: %w", err)
	}

	return nil
}

// readForWrite reads the session data from the store and creates entries.
// IMPORTANT!!! Do not forget to unlock the store after writing!
// Note: This function assumes that the store is NOT locked!
func (f *File) readForWrite() error {
	data, err := f.store.ReadForWrite()
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	err = f.createEntries(data)
	if err != nil {
		return fmt.Errorf("error creating entries: %w", err)
	}

	return nil
}

// writeAfterRead writes the current session data to the store.
// Note: This function assumes that the store is locked.
func (f *File) writeAfterRead() error {
	data, err := f.getData()
	if err != nil {
		return fmt.Errorf("error getting data: %w", err)
	}

	err = f.store.WriteLocked(data)
	if err != nil {
		return fmt.Errorf("error storing data: %w", err)
	}

	return nil
}
