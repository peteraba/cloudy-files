package store

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/apperr"
)

// S3 is a store that uses AWS S3 to store and retrieve data.
type S3 struct {
	client     *s3.Client
	logger     log.Logger
	bucket     string
	key        string
	lockKey    string
	maxRetries int
}

// NewS3 creates a new S3 instance.
func NewS3(client *s3.Client, logger log.Logger, bucket, path string) *S3 {
	return &S3{
		client:     client,
		logger:     logger,
		bucket:     bucket,
		key:        path,
		lockKey:    path + ".lock",
		maxRetries: defaultMaxRetries,
	}
}

// Read reads the data from S3.
func (s *S3) Read(ctx context.Context) ([]byte, error) {
	// Waiting for the lock to avoid reading inconsistent data
	err := s.waitForLockToBeRemoved(ctx)
	if err != nil {
		return nil, fmt.Errorf("error waiting for lock: %w", err)
	}

	// Reading the file (without locking)
	s.logger.Debug().Str("bucket", s.bucket).Str("key", s.key).Msg("reading file")

	data, err := os.ReadFile(s.key)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return data, nil
}

// ReadForWrite reads the data from S3, but also locks it for later writing.
func (s *S3) ReadForWrite(ctx context.Context) ([]byte, error) {
	// Waiting for the lock to be able to lock the file
	err := s.waitForLockToBeRemoved(ctx)
	if err != nil {
		return nil, fmt.Errorf("error waiting for lock: %w", err)
	}

	// Locking the file
	err = s.lock(ctx)
	if err != nil {
		return nil, fmt.Errorf("error locking file: %w", err)
	}

	// Reading the file
	s.logger.Debug().Str("bucket", s.bucket).Str("key", s.key).Msg("reading file")

	data, err := s.read(ctx)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return data, nil
}

// Write writes the data to S3.
func (s *S3) Write(ctx context.Context, data []byte) error {
	err := s.waitForLockToBeRemoved(ctx)
	if err != nil {
		return fmt.Errorf("error waiting for lock: %w", err)
	}

	err = s.lock(ctx)
	if err != nil {
		return fmt.Errorf("error locking file: %w", err)
	}
	defer s.Unlock(ctx)

	s.logger.Debug().Str("bucket", s.bucket).Str("key", s.key).Str("method", "Write").Msg("writing file")

	err = s.write(ctx, s.key, data)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	s.logger.Debug().Str("bucket", s.bucket).Str("key", s.key).Str("method", "Write").Msg("file written")

	return nil
}

// WriteLocked writes the data to S3 and then unlocks it.
func (s *S3) WriteLocked(ctx context.Context, data []byte) error {
	// Checking if the lock file exists
	_, err := os.Stat(s.lockKey)

	// Unexpected error
	exists, err := s.lockExists(ctx)
	if err != nil {
		return fmt.Errorf("error checking lock file: %s, err: %w", s.lockKey, err)
	}
	defer s.Unlock(ctx)

	// Lock file does not exist (it should)
	if !exists {
		return fmt.Errorf("lock file does not exist: %s, err: %w", s.lockKey, err)
	}

	// Writing the file
	s.logger.Debug().Str("bucket", s.bucket).Str("key", s.key).Str("method", "WriteLocked").Msg("writing file")

	err = s.write(ctx, s.key, data)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	s.logger.Debug().Str("bucket", s.bucket).Str("key", s.key).Str("method", "WriteLocked").Msg("file written")

	return nil
}

// waitForLockToBeRemoved waits for the lock file to be removed by any other process
// which may hold it. It retries for a maximum of N times.
func (s *S3) waitForLockToBeRemoved(ctx context.Context) error {
	s.logger.Debug().Str("bucket", s.bucket).Str("lockKey", s.lockKey).Msg("waiting for lock to be removed")

	count := 0

	for {
		exists, err := s.lockExists(ctx)
		if err != nil {
			return fmt.Errorf("error checking lock file: %s, err: %w", s.lockKey, err)
		}

		if !exists {
			s.logger.Debug().Str("bucket", s.bucket).Str("lockKey", s.lockKey).Msg("lock file does not exist, continue")

			return nil
		}

		s.logger.Debug().Str("bucket", s.bucket).Str("lockKey", s.lockKey).Msg("lock file exists")

		// Retrying logic
		count++

		if count > s.maxRetries {
			return fmt.Errorf("error waiting for lock file: %s, err: %w", s.lockKey, apperr.ErrLockTimeout)
		}

		time.Sleep(DefaultWaitTime)
	}
}

// lock creates a lock file to prevent other processes from writing to the file.
// It returns an error if the lock file already exists at the time of creation.
func (s *S3) lock(ctx context.Context) error {
	s.logger.Debug().Str("bucket", s.bucket).Str("lockKey", s.lockKey).Msg("acquiring lock")

	err := s.write(ctx, s.lockKey, []byte{})
	if err != nil {
		return fmt.Errorf("error creating lock file: %s, err: %w", s.lockKey, err)
	}

	return nil
}

// Unlock unlocks the S3 store.
func (s *S3) Unlock(ctx context.Context) error {
	s.logger.Debug().Str("bucket", s.bucket).Str("lockKey", s.lockKey).Msg("releasing lock")

	exists, err := s.lockExists(ctx)
	if err != nil {
		return fmt.Errorf("error checking lock file: %s, err: %w", s.lockKey, err)
	}

	if !exists {
		s.logger.Debug().Str("bucket", s.bucket).Str("lockKey", s.lockKey).Msg("lock file does not exist, continue")

		return nil
	}

	_, err = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{ //nolint:exhaustruct // No way to avoid this
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.lockKey),
	})
	if err != nil {
		return fmt.Errorf("error removing lock file: %s, err %w", s.lockKey, err)
	}

	s.logger.Debug().Str("bucket", s.bucket).Str("lockKey", s.lockKey).Msg("lock released")

	return nil
}

// write uploads a string to the specified S3 bucket and key.
func (s *S3) write(ctx context.Context, key string, data []byte) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{ //nolint:exhaustruct // No way to avoid this
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return fmt.Errorf("failed to put object, err: %w", err)
	}

	return nil
}

// lockExists retrieves a string from the specified S3 bucket and key.
func (s *S3) lockExists(ctx context.Context) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{ //nolint:exhaustruct // No way to avoid this
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.lockKey),
	})
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			return false, nil
		}

		return false, fmt.Errorf("unable to check if object exists, err: %w, debug: %#v", err, err)
	}

	return true, nil
}

// read retrieves a string from the specified S3 bucket and key.
func (s *S3) read(ctx context.Context) ([]byte, error) {
	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{ //nolint:exhaustruct // No way to avoid this
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object, err: %w", err)
	}

	buf := new(bytes.Buffer)

	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read from response body, err: %w", err)
	}

	return buf.Bytes(), nil
}

// Delete deletes the S3 store.
func (s *S3) Delete(ctx context.Context, path string) error {
	s.logger.Debug().Str("bucket", s.bucket).Str("path", path).Msg("deleting file")

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{ //nolint:exhaustruct // No way to avoid this
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return fmt.Errorf("error removing file: %s, err %w", path, err)
	}

	return nil
}
