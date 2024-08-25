package store

import (
	"fmt"
	"os"
	"path"

	"github.com/phuslu/log"
)

type FileSystem struct {
	logger log.Logger
	root   string
}

func NewFileSystem(logger log.Logger, rootPath string) *FileSystem {
	return &FileSystem{
		logger: logger,
		root:   rootPath,
	}
}

// Write writes the given data to the file with the given name using the root path.
// Subdirectory creation is not supported.
func (r *FileSystem) Write(name string, data []byte) error {
	r.logger.Debug().Msg("writing file: " + name)

	err := os.WriteFile(path.Join(r.root, name), data, defaultPermissions)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

// Read reads the file with the given name using the root path.
func (r *FileSystem) Read(name string) ([]byte, error) {
	r.logger.Debug().Msg("reading file: " + name)

	data, err := os.ReadFile(path.Join(r.root, name))
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return data, nil
}
