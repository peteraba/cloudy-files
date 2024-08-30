package store

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/apperr"
)

// File is a file-based Store implementation.
type File struct {
	logger       log.Logger
	fileName     string
	lockFileName string
	maxRetries   int
}

// DefaultWaitTime is the default time to wait between retries.
var DefaultWaitTime = 100 * time.Millisecond

const (
	defaultMaxRetries  = 10
	defaultPermissions = 0o600
)

// NewFile creates a new file instance.
func NewFile(logger log.Logger, fileName string) *File {
	return &File{
		logger:       logger,
		fileName:     fileName,
		lockFileName: fileName + ".lock",
		maxRetries:   defaultMaxRetries,
	}
}

// Read reads the file without acquiring the lock.
func (f *File) Read(_ context.Context) ([]byte, error) {
	// Waiting for the lock to avoid reading inconsistent data
	err := f.waitForLockToBeRemoved()
	if err != nil {
		return nil, fmt.Errorf("error waiting for lock: %w", err)
	}

	// Reading the file (without locking)
	f.logger.Debug().Msg("reading file")

	data, err := os.ReadFile(f.fileName)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return data, nil
}

// ReadForWrite reads the file after acquiring the lock.
func (f *File) ReadForWrite(_ context.Context) ([]byte, error) {
	// Waiting for the lock to be able to lock the file
	err := f.waitForLockToBeRemoved()
	if err != nil {
		return nil, fmt.Errorf("error waiting for lock: %w", err)
	}

	// Locking the file
	err = f.lock()
	if err != nil {
		return nil, fmt.Errorf("error locking file: %w", err)
	}

	// Reading the file
	f.logger.Debug().Msg("reading file")

	data, err := os.ReadFile(f.fileName)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return data, nil
}

// Write writes the data to the file after acquiring the lock.
func (f *File) Write(ctx context.Context, data []byte) error {
	err := f.waitForLockToBeRemoved()
	if err != nil {
		return fmt.Errorf("error waiting for lock: %w", err)
	}

	err = f.lock()
	if err != nil {
		return fmt.Errorf("error locking file: %w", err)
	}
	defer f.Unlock(ctx)

	f.logger.Debug().Str("method", "Write").Msg("writing file")

	err = os.WriteFile(f.fileName, data, defaultPermissions)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	f.logger.Debug().Str("method", "Write").Msg("file written")

	return nil
}

// WriteLocked writes data to the file after acquiring a lock
// used in pair with ReadForWrite.
// It returns an error if the lock file does not exist.
// It will unlock the file after writing.
func (f *File) WriteLocked(ctx context.Context, data []byte) error {
	// Checking if the lock file exists
	_, err := os.Stat(f.lockFileName)

	// Unexpected error
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error checking lock file: %s, err: %w", f.lockFileName, err)
	}
	defer f.Unlock(ctx)

	// Lock file does not exist (it should)
	if err != nil && os.IsNotExist(err) {
		return fmt.Errorf("lock file does not exist: %s, err: %w", f.lockFileName, err)
	}

	// Writing the file
	f.logger.Debug().Str("method", "WriteLocked").Msg("writing file")

	err = os.WriteFile(f.fileName, data, defaultPermissions)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	f.logger.Debug().Str("method", "WriteLocked").Msg("file written")

	return nil
}

// waitForLockToBeRemoved waits for the lock file to be removed by any other process
// which may hold it. It retries for a maximum of N times.
func (f *File) waitForLockToBeRemoved() error {
	f.logger.Debug().Msg("waiting for lock to be removed")

	count := 0

	for {
		_, err := os.Stat(f.lockFileName)
		// We're good to go, lock does not exist anymore
		if err != nil && os.IsNotExist(err) {
			f.logger.Debug().Msg("lock file does not exist, continue")

			return nil
		}
		// Unexpected error
		if err != nil {
			return fmt.Errorf("error checking lock file: %s, err: %w", f.lockFileName, err)
		}

		f.logger.Debug().Msg("lock file exists")

		// Retrying logic
		count++

		if count > f.maxRetries {
			return fmt.Errorf("error waiting for lock file: %s, err: %w", f.lockFileName, apperr.ErrLockTimeout)
		}

		time.Sleep(DefaultWaitTime)
	}
}

// lock creates a lock file to prevent other processes from writing to the file.
// It returns an error if the lock file already exists at the time of creation.
func (f *File) lock() error {
	f.logger.Debug().Msg("acquiring lock")

	lockFile, err := os.OpenFile(f.lockFileName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, defaultPermissions)
	if err != nil {
		return fmt.Errorf("error creating lock file: %s, err: %w", f.lockFileName, err)
	}

	defer lockFile.Close()

	return nil
}

// Unlock removes the lock file.
// It returns an error if the lock file does not exist.
func (f *File) Unlock(_ context.Context) error {
	f.logger.Debug().Msg("checking lock")

	// Checking if the lock file exists
	_, err := os.Stat(f.lockFileName)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error checking lock file: %s, err: %w", f.lockFileName, err)
	}

	if err != nil && os.IsNotExist(err) {
		return fmt.Errorf("lock file does not exist: %s, err: %w", f.lockFileName, err)
	}

	f.logger.Debug().Msg("releasing lock")

	err = os.Remove(f.lockFileName)
	if err != nil {
		return fmt.Errorf("error removing lock file: %s, err %w", f.lockFileName, err)
	}

	return nil
}
