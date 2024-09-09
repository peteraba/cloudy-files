package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/peteraba/cloudy-files/apperr"
)

// FileModel represents a file model.
type FileModel struct {
	Name   string   `json:"name"`
	Access []string `json:"access"`
}

// FileModels represents a file model list.
type FileModels []FileModel

// FileModelMap represents a file model map.
type FileModelMap map[string]FileModel

// Slice returns the file models as a slice.
func (f FileModelMap) Slice() FileModels {
	files := FileModels{}

	for _, file := range f {
		files = append(files, file)
	}

	return files
}

// File represents a file.
type File struct {
	store   Store
	lock    *sync.Mutex
	entries FileModelMap
}

// NewFile creates a new file instance.
func NewFile(store Store) *File {
	return &File{
		store:   store,
		lock:    &sync.Mutex{},
		entries: make(FileModelMap),
	}
}

// List lists all files.
func (f *File) List(ctx context.Context) (FileModels, error) {
	err := f.read(ctx)
	if err != nil {
		return nil, fmt.Errorf("error fetching from store: %w", err)
	}

	f.lock.Lock()
	defer f.lock.Unlock()

	return f.entries.Slice(), nil
}

// Get retrieves a file by name.
func (f *File) Get(ctx context.Context, name string) (FileModel, error) {
	err := f.read(ctx)
	if err != nil {
		return FileModel{}, fmt.Errorf("error reading file: %w", err)
	}

	f.lock.Lock()
	defer f.lock.Unlock()

	entry, ok := f.entries[name]
	if !ok {
		return FileModel{}, fmt.Errorf("file not found: %s, err: %w", name, apperr.ErrNotFound)
	}

	return entry, nil
}

// Create creates a file with the given name and access.
func (f *File) Create(ctx context.Context, name string, access []string) (FileModel, error) {
	err := f.readForWrite(ctx)
	if err != nil {
		return FileModel{}, fmt.Errorf("error reading file: %w", err)
	}
	defer f.store.Unlock(ctx)

	f.lock.Lock()
	defer f.lock.Unlock()

	f.entries[name] = FileModel{
		Name:   name,
		Access: access,
	}

	err = f.writeAfterRead(ctx)
	if err != nil {
		return FileModel{}, fmt.Errorf("error writing file: %w", err)
	}

	return f.entries[name], nil
}

// read reads the session data from the store and creates entries.
func (f *File) read(ctx context.Context) error {
	data, err := f.store.Read(ctx)
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
func (f *File) readForWrite(ctx context.Context) error {
	data, err := f.store.ReadForWrite(ctx)
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
func (f *File) writeAfterRead(ctx context.Context) error {
	data, err := f.getData()
	if err != nil {
		return fmt.Errorf("error creating raw data: %w", err)
	}

	err = f.store.WriteLocked(ctx, data)
	if err != nil {
		return fmt.Errorf("error storing data: %w", err)
	}

	return nil
}

// createEntries creates entries from data retrieved from store.
func (f *File) createEntries(data []byte) error {
	f.lock.Lock()
	defer f.lock.Unlock()

	entries := make(FileModelMap)

	if len(data) > 0 {
		err := json.Unmarshal(data, &entries)
		if err != nil {
			return fmt.Errorf("error unmarshaling data: %w", err)
		}
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
