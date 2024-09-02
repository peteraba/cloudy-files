package store

import (
	"context"
	"errors"
	"sync"

	"github.com/peteraba/cloudy-files/util"
)

type InMemory struct {
	m    *sync.Mutex
	data []byte
	spy  *util.Spy
}

// NewInMemory creates a new InMemory instance.
func NewInMemory(spy *util.Spy) *InMemory {
	return &InMemory{
		m:    &sync.Mutex{},
		data: make([]byte, 0),
		spy:  spy,
	}
}

// Read reads the file without acquiring the lock.
func (i *InMemory) Read(_ context.Context) ([]byte, error) {
	// Waiting for the lock to avoid reading inconsistent data
	i.waitForLock()

	if err := i.spy.GetError("Read"); err != nil {
		return nil, err
	}

	return i.data, nil
}

// ReadForWrite reads the file after acquiring the lock.
func (i *InMemory) ReadForWrite(_ context.Context) ([]byte, error) {
	if err := i.spy.GetError("ReadForWrite"); err != nil {
		return nil, err
	}

	// Waiting for the lock to be able to lock the file
	i.waitForLock()

	// Locking the file
	i.lock()

	return i.data, nil
}

// Write writes the data to the file after acquiring the lock.
func (i *InMemory) Write(ctx context.Context, data []byte) error {
	if err := i.spy.GetError("Write", data); err != nil {
		return err
	}

	// Waiting for the lock to be able to lock the file
	i.waitForLock()

	// Locking the file
	i.lock()
	defer i.Unlock(ctx)

	i.data = data

	return nil
}

var ErrLockDoesNotExist = errors.New("lock not locked")

// WriteLocked writes data to the file after acquiring a lock
// used in pair with ReadForWrite.
// It returns an error if the lock file does not exist.
func (i *InMemory) WriteLocked(_ context.Context, data []byte) error {
	if err := i.spy.GetError("WriteLocked", data); err != nil {
		return err
	}

	if i.m.TryLock() {
		defer i.m.Unlock()

		return ErrLockDoesNotExist
	}

	i.data = data

	return nil
}

// waitForLock waits for the lock file to be removed by any other process
// which may hold it. It retries for a maximum of N times.
func (i *InMemory) waitForLock() {
	i.m.Lock()
	defer i.m.Unlock()

	_ = ""
}

// lock creates a lock file to prevent other processes from writing to the file.
// It returns an error if the lock file already exists at the time of creation.
func (i *InMemory) lock() {
	i.m.Lock()
}

// Unlock removes the lock.
func (i *InMemory) Unlock(_ context.Context) error {
	i.m.Unlock()

	return nil
}
