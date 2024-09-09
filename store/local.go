package store

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/apperr"
)

// Local is a file-based Store implementation.
type Local struct {
	logger       *log.Logger
	fileName     string
	lockFileName string
	maxRetries   int
}

// NewLocal creates a new file instance.
func NewLocal(logger *log.Logger, fileName string) *Local {
	return &Local{
		logger:       logger,
		fileName:     fileName,
		lockFileName: fileName + ".lock",
		maxRetries:   defaultMaxRetries,
	}
}

// Read reads the file without acquiring the lock.
func (l *Local) Read(_ context.Context) ([]byte, error) {
	// Waiting for the lock to avoid reading inconsistent data
	err := l.waitForLockToBeRemoved()
	if err != nil {
		return nil, fmt.Errorf("error waiting for lock: %w", err)
	}

	// Reading the file (without locking)
	l.logger.Debug().Msg("reading file")

	data, err := os.ReadFile(l.fileName)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return data, nil
}

// ReadForWrite reads the file after acquiring the lock.
func (l *Local) ReadForWrite(_ context.Context) ([]byte, error) {
	// Waiting for the lock to be able to lock the file
	err := l.waitForLockToBeRemoved()
	if err != nil {
		return nil, fmt.Errorf("error waiting for lock: %w", err)
	}

	// Locking the file
	err = l.lock()
	if err != nil {
		return nil, fmt.Errorf("error locking file: %w", err)
	}

	// Reading the file
	l.logger.Debug().Msg("reading file")

	data, err := os.ReadFile(l.fileName)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return data, nil
}

// Write writes the data to the file after acquiring the lock.
func (l *Local) Write(ctx context.Context, data []byte) error {
	err := l.waitForLockToBeRemoved()
	if err != nil {
		return fmt.Errorf("error waiting for lock: %w", err)
	}

	err = l.lock()
	if err != nil {
		return fmt.Errorf("error locking file: %w", err)
	}
	defer l.Unlock(ctx)

	l.logger.Debug().Str("method", "Write").Msg("writing file")

	err = os.WriteFile(l.fileName, data, defaultPermissions)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	l.logger.Debug().Str("method", "Write").Msg("file written")

	return nil
}

// WriteLocked writes data to the file after acquiring a lock
// used in pair with ReadForWrite.
// It returns an error if the lock file does not exist.
// It will unlock the file after writing.
func (l *Local) WriteLocked(ctx context.Context, data []byte) error {
	// Checking if the lock file exists
	_, err := os.Stat(l.lockFileName)
	// Truly unexpected error
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("error checking lock file: %s, err: %w", l.lockFileName, err)
		}

		return fmt.Errorf("lock file does not exist: %s, err: %w", l.lockFileName, err)
	}
	defer l.Unlock(ctx)

	// Writing the file
	l.logger.Debug().Str("method", "WriteLocked").Msg("writing file")

	err = os.WriteFile(l.fileName, data, defaultPermissions)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	l.logger.Debug().Str("method", "WriteLocked").Msg("file written")

	return nil
}

// waitForLockToBeRemoved waits for the lock file to be removed by any other process
// which may hold it. It retries for a maximum of N times.
func (l *Local) waitForLockToBeRemoved() error {
	l.logger.Debug().Msg("waiting for lock to be removed")

	count := 0

	for {
		_, err := os.Stat(l.lockFileName)
		// We're good to go, lock does not exist anymore
		if err != nil && os.IsNotExist(err) {
			l.logger.Debug().Msg("lock file does not exist, continue")

			return nil
		}
		// Unexpected error
		if err != nil {
			return fmt.Errorf("error checking lock file: %s, err: %w", l.lockFileName, err)
		}

		l.logger.Debug().Msg("lock file exists")

		// Retrying logic
		count++

		if count > l.maxRetries {
			return fmt.Errorf("error waiting for lock file: %s, err: %w", l.lockFileName, apperr.ErrLockTimeout)
		}

		time.Sleep(DefaultWaitTime)
	}
}

// lock creates a lock file to prevent other processes from writing to the file.
// It returns an error if the lock file already exists at the time of creation.
func (l *Local) lock() error {
	l.logger.Debug().Msg("acquiring lock")

	lockFile, err := os.OpenFile(l.lockFileName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, defaultPermissions)
	if err != nil {
		return fmt.Errorf("error creating lock file: %s, err: %w", l.lockFileName, err)
	}

	defer lockFile.Close()

	return nil
}

// Unlock removes the lock file.
// It returns an error if the lock file does not exist.
func (l *Local) Unlock(_ context.Context) error {
	l.logger.Debug().Msg("checking lock")

	// Checking if the lock file exists
	_, err := os.Stat(l.lockFileName)
	if err != nil && !os.IsNotExist(err) {
		if !os.IsNotExist(err) {
			return fmt.Errorf("error checking lock file: %s, err: %w", l.lockFileName, err)
		}

		return fmt.Errorf("lock file does not exist: %s, err: %w", l.lockFileName, err)
	}

	l.logger.Debug().Msg("releasing lock")

	err = os.Remove(l.lockFileName)
	if err != nil {
		return fmt.Errorf("error removing lock file: %s, err %w", l.lockFileName, err)
	}

	return nil
}
