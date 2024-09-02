package filesystem

import (
	"context"
	"fmt"
	"sync"

	"github.com/peteraba/cloudy-files/apperr"
	"github.com/peteraba/cloudy-files/util"
)

type InMemory struct {
	mutex *sync.RWMutex
	data  map[string][]byte
	spy   *util.Spy
}

func NewInMemory(spy *util.Spy) *InMemory {
	return &InMemory{
		mutex: &sync.RWMutex{},
		data:  make(map[string][]byte),
		spy:   spy,
	}
}

// Write writes data to memory.
func (i *InMemory) Write(_ context.Context, name string, data []byte) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	if err := i.spy.GetError("Write", name, data); err != nil {
		return err
	}

	i.data[name] = data

	return nil
}

// Read returns data previously written.
func (i *InMemory) Read(_ context.Context, name string) ([]byte, error) {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	if err := i.spy.GetError("Read", name); err != nil {
		return nil, err
	}

	if data, ok := i.data[name]; ok {
		return data, nil
	}

	return nil, fmt.Errorf("error reading file: %w", apperr.ErrNotFound)
}
