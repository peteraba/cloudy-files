package store

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3 is a store that uses AWS S3 to store and retrieve data.
type S3 struct {
	client *s3.Client
}

// NewS3 creates a new S3 instance.
func NewS3(client *s3.Client) *S3 {
	return &S3{client: client}
}

// Read reads the data from S3.
func (s *S3) Read(_ context.Context) ([]byte, error) {
	return nil, nil
}

// ReadForWrite reads the data from S3, but also locks it for later writing.
func (s *S3) ReadForWrite(_ context.Context) ([]byte, error) {
	return nil, nil
}

// Write writes the data to S3.
func (s *S3) Write(_ context.Context, _ []byte) error {
	return nil
}

// WriteLocked writes the data to S3 and then unlocks it.
func (s *S3) WriteLocked(_ context.Context, _ []byte) error {
	return nil
}

// Unlock unlocks the S3 store.
func (s *S3) Unlock(_ context.Context) error {
	return nil
}

// // putObject uploads a string to the specified S3 bucket and key.
// func (s *S3) putObject(ctx context.Context, bucket, key string, content []byte) error {
// 	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{ //nolint:exhaustruct // No way to avoid this
// 		Bucket: aws.String(bucket),
// 		Key:    aws.String(key),
// 		Body:   bytes.NewReader(content),
// 	})
// 	if err != nil {
// 		return fmt.Errorf("failed to put object, err: %w", err)
// 	}
//
// 	return nil
// }
//
// // getObject retrieves a string from the specified S3 bucket and key.
// func (s *S3) getObject(ctx context.Context, bucket, key string) ([]byte, error) {
// 	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{ //nolint:exhaustruct // No way to avoid this
// 		Bucket: aws.String(bucket),
// 		Key:    aws.String(key),
// 	})
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get object, err: %w", err)
// 	}
//
// 	buf := new(bytes.Buffer)
//
// 	_, err = buf.ReadFrom(resp.Body)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read from response body, err: %w", err)
// 	}
//
// 	return buf.Bytes(), nil
// }
