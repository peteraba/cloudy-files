package store

import (
	"fmt"
	"sync"

	"github.com/peteraba/cloudy-files/apperr"
)

type InMemoryFileSystem struct {
	m    *sync.RWMutex
	data map[string][]byte
}

func NewInMemoryFileSystem() *InMemoryFileSystem {
	return &InMemoryFileSystem{
		m:    &sync.RWMutex{},
		data: make(map[string][]byte),
	}
}

// Write writes data to memory.
func (imfs *InMemoryFileSystem) Write(name string, data []byte) error {
	imfs.m.Lock()
	defer imfs.m.Unlock()

	imfs.data[name] = data

	return nil
}

// Read returns data previously written.
func (imfs *InMemoryFileSystem) Read(name string) ([]byte, error) {
	imfs.m.RLock()
	defer imfs.m.RUnlock()

	if data, ok := imfs.data[name]; ok {
		return data, nil
	}

	return nil, fmt.Errorf("error reading file: %w", apperr.ErrNotFound)
}
