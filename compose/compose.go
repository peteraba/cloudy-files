package compose

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/password"
	"github.com/peteraba/cloudy-files/repo"
	"github.com/peteraba/cloudy-files/service"
	"github.com/peteraba/cloudy-files/store"
)

type Factory struct {
	mutex                  *sync.Mutex
	fileSystemInstance     service.FileSystem
	fileStoreInstance      repo.Store
	userStoreInstance      repo.Store
	sessionStoreInstance   repo.Store
	passwordHasherInstance service.PasswordHasher
}

func NewFactory() *Factory {
	return &Factory{
		mutex:                  &sync.Mutex{},
		fileSystemInstance:     nil,
		fileStoreInstance:      nil,
		userStoreInstance:      nil,
		sessionStoreInstance:   nil,
		passwordHasherInstance: nil,
	}
}

func (f *Factory) CreateFileService() *service.File {
	logger := f.GetLogger()

	fileStore := f.getFileStore(logger)
	fileRepo := f.createFileRepo(fileStore)

	fsStore := f.getFileSystem(logger)

	return service.NewFile(fileRepo, fsStore, logger)
}

func (f *Factory) CreateUserService() *service.User {
	logger := f.GetLogger()

	userStore := f.getUserStore(logger)
	userRepo := f.createUserRepo(userStore)
	sessionStore := f.getSessionStore(logger)
	sessionRepo := f.createSessionRepo(sessionStore)
	hasher := f.getHasher()
	rawChecker := f.createRawPasswordChecker()

	return service.NewUser(userRepo, sessionRepo, hasher, rawChecker, logger)
}

func (f *Factory) CreateSessionService() *service.Session {
	logger := f.GetLogger()

	sessionStore := f.getSessionStore(logger)
	sessionRepo := f.createSessionRepo(sessionStore)

	return service.NewSession(sessionRepo, logger)
}

func (f *Factory) getFileSystem(logger log.Logger) service.FileSystem {
	if f.fileSystemInstance == nil {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		f.fileSystemInstance = store.NewFileSystem(logger, filepath.Join(wd, "files"))
	}

	return f.fileSystemInstance
}

func (f *Factory) SetFileSystem(fs service.FileSystem) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.fileSystemInstance = fs
}

func (f *Factory) getFileStore(logger log.Logger) repo.Store {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.fileStoreInstance == nil {
		f.fileStoreInstance = store.NewFile(logger, filepath.Join("data", "files.json"))
	}

	return f.fileStoreInstance
}

func (f *Factory) SetFileStore(storeInstance repo.Store) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.fileStoreInstance = storeInstance
}

func (f *Factory) createFileRepo(fileStore repo.Store) *repo.File {
	return repo.NewFile(fileStore)
}

func (f *Factory) getUserStore(logger log.Logger) repo.Store {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.userStoreInstance == nil {
		f.userStoreInstance = store.NewFile(logger, filepath.Join("data", "users.json"))
	}

	return f.userStoreInstance
}

func (f *Factory) SetUserStore(storeInstance repo.Store) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.userStoreInstance = storeInstance
}

func (f *Factory) createUserRepo(userStore repo.Store) *repo.User {
	return repo.NewUser(userStore)
}

func (f *Factory) getSessionStore(logger log.Logger) repo.Store {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.sessionStoreInstance == nil {
		f.sessionStoreInstance = store.NewFile(logger, filepath.Join("data", "sessions.json"))
	}

	return f.sessionStoreInstance
}

func (f *Factory) SetSessionStore(storeInstance repo.Store) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.sessionStoreInstance = storeInstance
}

func (f *Factory) createSessionRepo(sessionStore repo.Store) *repo.Session {
	return repo.NewSession(sessionStore)
}

func (f *Factory) getHasher() service.PasswordHasher {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.passwordHasherInstance == nil {
		f.passwordHasherInstance = password.NewBcryptHasher()
	}

	return f.passwordHasherInstance
}

func (f *Factory) SetHasher(hasher service.PasswordHasher) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.passwordHasherInstance = hasher
}

func (f *Factory) createRawPasswordChecker() *password.Checker {
	return password.NewChecker()
}

func (f *Factory) GetLogger() log.Logger {
	return log.DefaultLogger
}
