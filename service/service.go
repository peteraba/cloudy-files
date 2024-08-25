package service

import "github.com/peteraba/cloudy-files/repo"

type userRepo interface {
	Get(name string) (repo.UserModel, error)
	Create(name, email, password string, isAdmin bool, access []string) error
}

type sessionRepo interface {
	Check(name, hash string) (bool, error)
	Start(name string) (string, error)
	CleanUp() error
}

type passwordHasher interface {
	Check(password, hashedPassword string) error
	Hash(password string) (string, error)
}

type passwordChecker interface {
	IsOK(password string) error
}

type fileSystem interface {
	Write(name string, data []byte) error
	Read(name string) ([]byte, error)
}

type fileRepo interface {
	Get(name string) (repo.FileModel, error)
	Create(name string, access []string) error
}
