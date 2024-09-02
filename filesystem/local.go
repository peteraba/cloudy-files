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
	logger *log.Logger
	root   string
}

// NewLocal creates a new Local instance.
func NewLocal(logger *log.Logger, rootPath string) *Local {
	return &Local{
		logger: logger,
		root:   rootPath,
	}
}

// Write writes the given data to the file with the given name using the bucket path.
// Subdirectory creation is not supported.
func (l *Local) Write(_ context.Context, fileName string, data []byte) error {
	l.logger.Debug().Str("root", l.root).Str("fileName", fileName).Msg("writing file")

	err := os.WriteFile(filepath.Join(l.root, fileName), data, defaultPermissions)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	l.logger.Debug().Str("root", l.root).Str("fileName", fileName).Msg("file written")

	return nil
}

// Read reads the file with the given name using the bucket path.
func (l *Local) Read(_ context.Context, fileName string) ([]byte, error) {
	l.logger.Debug().Str("root", l.root).Str("fileName", fileName).Msg("reading file")

	data, err := os.ReadFile(filepath.Join(l.root, fileName))
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	l.logger.Debug().Str("root", l.root).Str("fileName", fileName).Msg("file read")

	return data, nil
}
