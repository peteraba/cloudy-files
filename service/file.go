package service

import (
	"fmt"

	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/apperr"
	"github.com/peteraba/cloudy-files/util"
)

type File struct {
	logger log.Logger
	repo   fileRepo
	store  fileSystem
}

func NewFile(repo fileRepo, store fileSystem, logger log.Logger) *File {
	return &File{
		logger: logger,
		repo:   repo,
		store:  store,
	}
}

// Upload uploads a file with the given name and content.
func (f *File) Upload(name string, content []byte, access []string) error {
	err := f.store.Write(name, content)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	err = f.repo.Create(name, access)
	if err != nil {
		return fmt.Errorf("error creating model: %w", err)
	}

	return nil
}

// Retrieve retrieves the content of a file by name.
func (f *File) Retrieve(name string, access []string) ([]byte, error) {
	file, err := f.repo.Get(name)
	if err != nil {
		return nil, fmt.Errorf("error retrieving model: %w", err)
	}

	if len(util.Intersection(file.Access, access)) == 0 {
		return nil, fmt.Errorf("access denied: %w", apperr.ErrAccessDenied)
	}

	data, err := f.store.Read(name)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return data, nil
}
