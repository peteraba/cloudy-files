package service

import (
	"context"
	"fmt"

	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/apperr"
	"github.com/peteraba/cloudy-files/util"
)

type File struct {
	logger log.Logger
	repo   FileRepo
	store  FileSystem
}

func NewFile(repo FileRepo, store FileSystem, logger log.Logger) *File {
	return &File{
		logger: logger,
		repo:   repo,
		store:  store,
	}
}

// Upload uploads a file with the given name and content.
func (f *File) Upload(ctx context.Context, name string, content []byte, access []string) error {
	f.logger.Info().Str("name", name).Msg("uploading file")

	err := f.store.Write(ctx, name, content)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	f.logger.Info().Str("name", name).Msg("updating file DB")

	err = f.repo.Create(ctx, name, access)
	if err != nil {
		return fmt.Errorf("error creating model: %w", err)
	}

	f.logger.Info().Str("name", name).Msg("updated file")

	return nil
}

// Retrieve retrieves the content of a file by name.
func (f *File) Retrieve(ctx context.Context, name string, access []string) ([]byte, error) {
	file, err := f.repo.Get(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("error retrieving model: %w", err)
	}

	if len(util.Intersection(file.Access, access)) == 0 {
		return nil, fmt.Errorf("access denied: %w", apperr.ErrAccessDenied)
	}

	data, err := f.store.Read(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return data, nil
}
