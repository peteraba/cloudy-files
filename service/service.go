package service

import (
	"context"

	"github.com/peteraba/cloudy-files/repo"
)

type UserRepo interface {
	Get(ctx context.Context, name string) (repo.UserModel, error)
	List(ctx context.Context) (repo.UserModels, error)
	Create(ctx context.Context, name, email, password string, isAdmin bool, access []string) (repo.UserModel, error)
	UpdatePassword(ctx context.Context, name, password string) (repo.UserModel, error)
	UpdateAccess(ctx context.Context, name string, access []string) (repo.UserModel, error)
	Promote(ctx context.Context, name string) (repo.UserModel, error)
	Demote(ctx context.Context, name string) (repo.UserModel, error)
	Delete(ctx context.Context, name string) error
}

type SessionRepo interface {
	Get(ctx context.Context, name string) (repo.SessionUser, error)
	Start(ctx context.Context, name string, isAdmin bool, access []string) (repo.SessionUser, error)
	CleanUp(ctx context.Context) error
}

type PasswordHasher interface {
	Check(ctx context.Context, password, hashedPassword string) error
	Hash(ctx context.Context, password string) (string, error)
}

type PasswordChecker interface {
	IsOK(ctx context.Context, password string) error
}

type FileSystem interface {
	Write(ctx context.Context, name string, data []byte) error
	Read(ctx context.Context, name string) ([]byte, error)
}

type FileRepo interface {
	Get(ctx context.Context, name string) (repo.FileModel, error)
	List(ctx context.Context) (repo.FileModels, error)
	Create(ctx context.Context, name string, access []string) (repo.FileModel, error)
}
