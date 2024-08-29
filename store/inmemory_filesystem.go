package store

import (
	"fmt"
	"sync"

	"github.com/peteraba/cloudy-files/apperr"
	"github.com/peteraba/cloudy-files/util"
)

type InMemoryFileSystem struct {
	m    *sync.RWMutex
	data map[string][]byte
	spy  *util.Spy
}

func NewInMemoryFileSystem(spy *util.Spy) *InMemoryFileSystem {
	return &InMemoryFileSystem{
		m:    &sync.RWMutex{},
		data: make(map[string][]byte),
		spy:  spy,
	}
}

// Write writes data to memory.
func (imfs *InMemoryFileSystem) Write(name string, data []byte) error {
	imfs.m.Lock()
	defer imfs.m.Unlock()

	if err := imfs.spy.GetError("Write", name, data); err != nil {
		return err
	}

	imfs.data[name] = data

	return nil
}

// Read returns data previously written.
func (imfs *InMemoryFileSystem) Read(name string) ([]byte, error) {
	imfs.m.RLock()
	defer imfs.m.RUnlock()

	if err := imfs.spy.GetError("Read", name); err != nil {
		return nil, err
	}

	if data, ok := imfs.data[name]; ok {
		return data, nil
	}

	return nil, fmt.Errorf("error reading file: %w", apperr.ErrNotFound)
}
