package store

import (
	"errors"
	"sync"
)

type InMemoryFile struct {
	m    *sync.Mutex
	data []byte
}

func NewInMemoryFile() *InMemoryFile {
	return &InMemoryFile{
		m:    &sync.Mutex{},
		data: make([]byte, 0),
	}
}

// Read reads the file without acquiring the lock.
func (imf *InMemoryFile) Read() ([]byte, error) {
	// Waiting for the lock to avoid reading inconsistent data
	imf.waitForLock()

	return imf.data, nil
}

// ReadForWrite reads the file after acquiring the lock.
func (imf *InMemoryFile) ReadForWrite() ([]byte, error) {
	// Waiting for the lock to be able to lock the file
	imf.waitForLock()

	// Locking the file
	imf.lock()

	return imf.data, nil
}

// Write writes the data to the file after acquiring the lock.
func (imf *InMemoryFile) Write(data []byte) error {
	// Waiting for the lock to be able to lock the file
	imf.waitForLock()

	// Locking the file
	imf.lock()
	defer imf.Unlock()

	imf.data = data

	return nil
}

var ErrLockDoesNotExist = errors.New("lock not locked")

// WriteLocked writes data to the file after acquiring a lock
// used in pair with ReadForWrite.
// It returns an error if the lock file does not exist.
func (imf *InMemoryFile) WriteLocked(data []byte) error {
	if imf.m.TryLock() {
		defer imf.m.Unlock()

		return ErrLockDoesNotExist
	}

	imf.data = data

	return nil
}

// waitForLock waits for the lock file to be removed by any other process
// which may hold it. It retries for a maximum of N times.
func (imf *InMemoryFile) waitForLock() {
	imf.m.Lock()
	defer imf.m.Unlock()

	_ = ""
}

// lock creates a lock file to prevent other processes from writing to the file.
// It returns an error if the lock file already exists at the time of creation.
func (imf *InMemoryFile) lock() {
	imf.m.Lock()
}

// Unlock removes the lock.
func (imf *InMemoryFile) Unlock() error {
	imf.m.Unlock()

	return nil
}
