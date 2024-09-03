package compose

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/appconfig"
	"github.com/peteraba/cloudy-files/cli"
	"github.com/peteraba/cloudy-files/filesystem"
	"github.com/peteraba/cloudy-files/http"
	"github.com/peteraba/cloudy-files/password"
	"github.com/peteraba/cloudy-files/repo"
	"github.com/peteraba/cloudy-files/service"
	"github.com/peteraba/cloudy-files/store"
)

// DataType represents the type of data stored in a store.
type DataType int

const (
	// SessionStore represents a store for session data.
	SessionStore DataType = iota
	// UserStore represents a store for user data.
	UserStore
	// FileStore represents a store for file data.
	FileStore
)

// Factory is a factory for creating services.
type Factory struct {
	mutex                  *sync.RWMutex
	fileSystemInstance     service.FileSystem
	stores                 [3]repo.Store
	passwordHasherInstance service.PasswordHasher
	s3Client               *s3.Client
	appConfig              *appconfig.Config
	logger                 *log.Logger
}

var filePaths = [...]string{"sessions.json", "users.json", "files.json"} //nolint:gochecknoglobals // This is a constant

// NewFactory creates a new factory.
func NewFactory(appConfig *appconfig.Config) *Factory {
	return &Factory{
		mutex:                  &sync.RWMutex{},
		fileSystemInstance:     nil,
		stores:                 [...]repo.Store{nil, nil, nil},
		passwordHasherInstance: nil,
		s3Client:               nil,
		logger:                 &log.DefaultLogger,
		appConfig:              appConfig,
	}
}

// NewTestFactory creates a new factory for testing.
func NewTestFactory(appConfig *appconfig.Config) *Factory {
	f := NewFactory(appConfig)

	f.SetLogLevel(log.PanicLevel)

	return f
}

func (f *Factory) SetLogLevel(level log.Level) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.logger = &log.Logger{
		Level:        level,
		Caller:       0,
		TimeField:    "",
		TimeFormat:   "",
		TimeLocation: nil,
		Context:      nil,
		Writer:       log.IOWriter{Writer: os.Stderr},
	}
}

// SetAWS sets the AWS configuration for the factory.
func (f *Factory) SetAWS(awsConfig aws.Config) *Factory { //nolint:gocritic // aws.Config might be huge (320 bytes, but it's a one-off)
	f.s3Client = s3.NewFromConfig(awsConfig)

	return f
}

// CreateCliApp creates a CLI app.
func (f *Factory) CreateCliApp() *cli.App {
	return cli.NewApp(
		f.CreateSessionService(),
		f.CreateUserService(),
		f.CreateFileService(),
		f.logger,
	)
}

// CreateHTTPApp creates an HTTP app.
func (f *Factory) CreateHTTPApp() *http.App {
	return http.NewApp(
		f.CreateSessionService(),
		f.CreateUserService(),
		f.CreateFileService(),
		f.logger,
	)
}

// GetS3Client returns the S3 client.
func (f *Factory) GetS3Client() *s3.Client {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	return f.s3Client
}

// CreateFileService creates a file service.
func (f *Factory) CreateFileService() *service.File {
	fileStore := f.getStore(FileStore)
	fileRepo := f.createFileRepo(fileStore)

	fsStore := f.getFileSystem()

	return service.NewFile(fileRepo, fsStore, *f.logger)
}

// CreateUserService creates a user service.
func (f *Factory) CreateUserService() *service.User {
	userStore := f.getStore(UserStore)
	userRepo := f.createUserRepo(userStore)
	sessionStore := f.getStore(SessionStore)
	sessionRepo := f.createSessionRepo(sessionStore)
	hasher := f.getHasher()
	rawChecker := f.createRawPasswordChecker()

	return service.NewUser(userRepo, sessionRepo, hasher, rawChecker, *f.logger)
}

// CreateSessionService creates a session service.
func (f *Factory) CreateSessionService() *service.Session {
	sessionStore := f.getStore(SessionStore)
	sessionRepo := f.createSessionRepo(sessionStore)

	return service.NewSession(sessionRepo, *f.logger)
}

func (f *Factory) getFileSystem() service.FileSystem {
	if f.fileSystemInstance == nil {
		f.createFileSystem()
	}

	return f.fileSystemInstance
}

func (f *Factory) createFileSystem() {
	if f.s3Client != nil {
		f.fileSystemInstance = filesystem.NewS3(f.s3Client, f.logger, f.appConfig.FileSystemAwsBucket)

		return
	}

	workDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	f.fileSystemInstance = filesystem.NewLocal(f.logger, filepath.Join(workDir, f.appConfig.FileSystemLocalPath))
}

// SetFileSystem sets the file system for the factory.
func (f *Factory) SetFileSystem(fs service.FileSystem) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.fileSystemInstance = fs
}

func (f *Factory) getStore(dataType DataType) repo.Store {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	if f.stores[dataType] == nil {
		f.stores[dataType] = f.createStore(dataType)
	}

	return f.stores[dataType]
}

func (f *Factory) createStore(dataType DataType) repo.Store {
	if f.s3Client != nil {
		return store.NewS3(f.s3Client, f.logger, f.appConfig.StoreAwsBucket, filePaths[dataType])
	}

	workDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return store.NewLocal(f.logger, filepath.Join(workDir, f.appConfig.StoreLocalPath, filePaths[dataType]))
}

// SetStore sets the store for the factory.
func (f *Factory) SetStore(storeInstance repo.Store, dataType DataType) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.stores[dataType] = storeInstance
}

func (f *Factory) createFileRepo(fileStore repo.Store) *repo.File {
	return repo.NewFile(fileStore)
}

func (f *Factory) createUserRepo(userStore repo.Store) *repo.User {
	return repo.NewUser(userStore)
}

func (f *Factory) createSessionRepo(sessionStore repo.Store) *repo.Session {
	return repo.NewSession(sessionStore)
}

func (f *Factory) getHasher() service.PasswordHasher {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	if f.passwordHasherInstance == nil {
		f.passwordHasherInstance = password.NewBcryptHasher()
	}

	return f.passwordHasherInstance
}

// SetHasher sets the password hasher for the factory.
func (f *Factory) SetHasher(hasher service.PasswordHasher) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.passwordHasherInstance = hasher
}

func (f *Factory) createRawPasswordChecker() *password.Checker {
	return password.NewChecker()
}

// GetLogger returns the logger.
func (f *Factory) GetLogger() *log.Logger {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	return f.logger
}

// SetLogger sets the logger for the factory.
func (f *Factory) SetLogger(logger *log.Logger) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.logger = logger
}
