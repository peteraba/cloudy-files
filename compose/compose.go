package compose

import (
	"os"
	"path"

	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/password"
	"github.com/peteraba/cloudy-files/repo"
	"github.com/peteraba/cloudy-files/service"
	"github.com/peteraba/cloudy-files/store"
)

func CreateFileService() *service.File {
	logger := GetLogger()

	fileStore := createFileStore(logger)
	fileRepo := createFileRepo(fileStore)

	fsStore := createFileSystem(logger)

	return service.NewFile(fileRepo, fsStore, logger)
}

func CreateUserService() *service.User {
	logger := GetLogger()

	userStore := createUserStore(logger)
	userRepo := createUserRepo(userStore)
	sessionStore := createSessionStore(logger)
	sessionRepo := createSessionRepo(sessionStore)
	bcrypt := createHasher()
	rawChecker := createRawPasswordChecker()

	return service.NewUser(userRepo, sessionRepo, bcrypt, rawChecker, logger)
}

func CreateSessionService() *service.Session {
	logger := GetLogger()

	sessionStore := createSessionStore(logger)
	sessionRepo := createSessionRepo(sessionStore)

	return service.NewSession(sessionRepo, logger)
}

func createFileSystem(logger log.Logger) *store.FileSystem {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return store.NewFileSystem(logger, path.Join(wd, "files"))
}

func createFileStore(logger log.Logger) *store.File {
	return store.NewFile(logger, "data/files.json")
}

func createFileRepo(fileStore repo.Store) *repo.File {
	return repo.NewFile(fileStore)
}

func createUserStore(logger log.Logger) *store.File {
	return store.NewFile(logger, "data/users.json")
}

func createUserRepo(userStore repo.Store) *repo.User {
	return repo.NewUser(userStore)
}

func createSessionStore(logger log.Logger) *store.File {
	return store.NewFile(logger, "data/sessions.json")
}

func createSessionRepo(sessionStore repo.Store) *repo.Session {
	return repo.NewSession(sessionStore)
}

func createHasher() *password.Bcrypt {
	return password.NewBcrypt()
}

func createRawPasswordChecker() *password.Checker {
	return password.NewChecker()
}

func GetLogger() log.Logger {
	return log.DefaultLogger
}
