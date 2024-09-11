package service

import (
	"context"
	"fmt"

	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/apperr"
	"github.com/peteraba/cloudy-files/repo"
	"github.com/peteraba/cloudy-files/util"
)

// File is a service that provides file-related operations.
type File struct {
	logger log.Logger
	repo   FileRepo
	store  FileSystem
}

// NewFile creates a new File service.
func NewFile(fileRepo FileRepo, store FileSystem, logger log.Logger) *File {
	return &File{
		logger: logger,
		repo:   fileRepo,
		store:  store,
	}
}

// Upload uploads a file with the given name and content.
func (f *File) Upload(ctx context.Context, name string, content []byte, access []string) (repo.FileModel, error) {
	f.logger.Info().Str("name", name).Msg("uploading file")

	err := f.store.Write(ctx, name, content)
	if err != nil {
		return repo.FileModel{}, fmt.Errorf("error writing file: %w", err)
	}

	f.logger.Info().Str("name", name).Msg("updating file DB")

	fileModel, err := f.repo.Create(ctx, name, access)
	if err != nil {
		return repo.FileModel{}, fmt.Errorf("error creating model: %w", err)
	}

	f.logger.Info().Str("name", name).Msg("updated file")

	return fileModel, nil
}

// Get retrieves a file model.
func (f *File) Get(ctx context.Context, name string) (repo.FileModel, error) {
	file, err := f.repo.Get(ctx, name)
	if err != nil {
		return repo.FileModel{}, fmt.Errorf("error retrieving model: %w", err)
	}

	return file, nil
}

// Retrieve retrieves the content of a file by name.
func (f *File) Retrieve(ctx context.Context, name string, access []string) ([]byte, error) {
	file, err := f.repo.Get(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("error retrieving model: %w", err)
	}

	if !util.HasIntersection(file.Access, access) {
		return nil, fmt.Errorf("access denied: %w", apperr.ErrAccessDenied)
	}

	data, err := f.store.Read(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return data, nil
}

// List lists files.
func (f *File) List(ctx context.Context, access []string, isAdmin bool) (repo.FileModels, error) {
	fileModels, err := f.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing files: %w", err)
	}

	if isAdmin {
		return fileModels, nil
	}

	var accessibleFiles []repo.FileModel

	for _, file := range fileModels {
		if util.HasIntersection(file.Access, access) {
			accessibleFiles = append(accessibleFiles, file)
		}
	}

	return accessibleFiles, nil
}
