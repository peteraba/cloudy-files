package filesystem

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/phuslu/log"
)

// S3 can store and retrieve any file from an S3 bucket.
type S3 struct {
	client *s3.Client
	logger *log.Logger
	bucket string
}

// NewS3 creates a new S3 instance.
func NewS3(client *s3.Client, logger *log.Logger, bucket string) *S3 {
	return &S3{
		client: client,
		logger: logger,
		bucket: bucket,
	}
}

// Write writes the given data to the file with the given name using the bucket path.
// Subdirectory creation is not supported.
func (s *S3) Write(ctx context.Context, path string, data []byte) error {
	s.logger.Debug().Str("bucket", s.bucket).Str("path", path).Msg("writing file")

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{ //nolint:exhaustruct // No way to avoid this
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return fmt.Errorf("failed to put object, err: %w", err)
	}

	s.logger.Debug().Str("bucket", s.bucket).Str("path", path).Msg("file written")

	return nil
}

// Read reads the file with the given name using the bucket path.
func (s *S3) Read(ctx context.Context, path string) ([]byte, error) {
	s.logger.Debug().Str("bucket", s.bucket).Str("path", path).Msg("reading file")

	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{ //nolint:exhaustruct // No way to avoid this
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object, err: %w", err)
	}

	s.logger.Debug().Str("bucket", s.bucket).Str("path", path).Msg("file received")

	buf := new(bytes.Buffer)

	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read from response body, err: %w", err)
	}

	s.logger.Debug().Str("bucket", s.bucket).Str("path", path).Msg("file read")

	return buf.Bytes(), nil
}
