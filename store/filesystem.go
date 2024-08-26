package store

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/phuslu/log"
)

// FileSystem can store and retrieve any file from a directory.
type FileSystem struct {
	logger log.Logger
	root   string
}

// NewFileSystem creates a new FileSystem instance.
func NewFileSystem(logger log.Logger, rootPath string) *FileSystem {
	return &FileSystem{
		logger: logger,
		root:   rootPath,
	}
}

// Write writes the given data to the file with the given name using the root path.
// Subdirectory creation is not supported.
func (fs *FileSystem) Write(name string, data []byte) error {
	fs.logger.Debug().Msg("writing file: " + name)

	err := os.WriteFile(filepath.Join(fs.root, name), data, defaultPermissions)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

// Read reads the file with the given name using the root path.
func (fs *FileSystem) Read(name string) ([]byte, error) {
	fs.logger.Debug().Msg("reading file: " + name)

	data, err := os.ReadFile(filepath.Join(fs.root, name))
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return data, nil
}
