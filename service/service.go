package service

import "github.com/peteraba/cloudy-files/repo"

type UserRepo interface {
	Get(name string) (repo.UserModel, error)
	Create(name, email, password string, isAdmin bool, access []string) error
}

type SessionRepo interface {
	Check(name, hash string) (bool, error)
	Start(name string) (string, error)
	CleanUp() error
}

type PasswordHasher interface {
	Check(password, hashedPassword string) error
	Hash(password string) (string, error)
}

type PasswordChecker interface {
	IsOK(password string) error
}

type FileSystem interface {
	Write(name string, data []byte) error
	Read(name string) ([]byte, error)
}

type FileRepo interface {
	Get(name string) (repo.FileModel, error)
	Create(name string, access []string) error
}
