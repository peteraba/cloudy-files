package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/phuslu/log"
)

const (
	defaultPermissions = 0o600
)

// Local can store and retrieve any file from a directory.
type Local struct {
	logger log.Logger
	root   string
}

// NewLocal creates a new Local instance.
func NewLocal(logger log.Logger, rootPath string) *Local {
	return &Local{
		logger: logger,
		root:   rootPath,
	}
}

// Write writes the given data to the file with the given name using the bucket path.
// Subdirectory creation is not supported.
func (fs *Local) Write(_ context.Context, name string, data []byte) error {
	fs.logger.Debug().Msg("writing file: " + name)

	err := os.WriteFile(filepath.Join(fs.root, name), data, defaultPermissions)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

// Read reads the file with the given name using the bucket path.
func (fs *Local) Read(_ context.Context, name string) ([]byte, error) {
	fs.logger.Debug().Msg("reading file: " + name)

	data, err := os.ReadFile(filepath.Join(fs.root, name))
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return data, nil
}
