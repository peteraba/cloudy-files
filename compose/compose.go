package compose

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gorilla/securecookie"
	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/appconfig"
	"github.com/peteraba/cloudy-files/cli"
	"github.com/peteraba/cloudy-files/filesystem"
	"github.com/peteraba/cloudy-files/http"
	"github.com/peteraba/cloudy-files/http/api"
	"github.com/peteraba/cloudy-files/http/web"
	"github.com/peteraba/cloudy-files/password"
	"github.com/peteraba/cloudy-files/repo"
	"github.com/peteraba/cloudy-files/service"
	"github.com/peteraba/cloudy-files/store"
)

// DataType represents the type of data stored in a store.
type DataType int

const (
	// UserStore represents a store for user data.
	UserStore DataType = iota
	// FileStore represents a store for file data.
	FileStore
	// CSRFStore represents a store for CSRF data.
	CSRFStore
)

// Factory is a factory for creating services.
type Factory struct {
	mutex                  *sync.RWMutex
	fileSystemInstance     service.FileSystem
	stores                 [3]repo.Store
	passwordHasherInstance service.PasswordHasher
	s3Client               *s3.Client
	appConfig              *appconfig.Config
	display                cli.Display
	cookieStore            *securecookie.SecureCookie
	logger                 *log.Logger
}

var filePaths = [...]string{"users.json", "files.json", "csrf.json"} //nolint:gochecknoglobals // This is a constant

// NewFactory creates a new factory.
func NewFactory(appConfig *appconfig.Config) *Factory {
	return &Factory{
		mutex:                  &sync.RWMutex{},
		fileSystemInstance:     nil,
		stores:                 [...]repo.Store{nil, nil, nil},
		passwordHasherInstance: nil,
		s3Client:               nil,
		appConfig:              appConfig,
		display:                nil,
		cookieStore:            nil,
		logger:                 &log.DefaultLogger,
	}
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
		f.CreateUserService(),
		f.CreateFileService(),
		f.GetDisplay(),
		f.logger,
	)
}

// CreateHTTPApp creates an HTTP app.
func (f *Factory) CreateHTTPApp() *http.App {
	return http.NewApp(
		f.CreateUserHandler(),
		f.CreateFileHandler(),
		f.CreateFallbackHandler(),
		f.logger,
	)
}

func (f *Factory) CreateUserHandler() *http.UserHandler {
	return http.NewUserHandler(
		f.CreateAPIUserHandler(),
		f.CreateWebUserHandler(),
		f.logger,
	)
}

func (f *Factory) CreateFileHandler() *http.FileHandler {
	return http.NewFileHandler(
		f.CreateAPIFileHandler(),
		f.CreateWebFileHandler(),
		f.logger,
	)
}

func (f *Factory) CreateFallbackHandler() *http.FallbackHandler {
	return http.NewFallbackHandler(
		f.CreateAPIFallbackHandler(),
		f.CreateWebFallbackHandler(),
		f.logger,
	)
}

func (f *Factory) CreateAPIUserHandler() *api.UserHandler {
	return api.NewUserHandler(
		f.CreateUserService(),
		f.logger,
	)
}

func (f *Factory) CreateAPIFileHandler() *api.FileHandler {
	return api.NewFileHandler(
		f.CreateFileService(),
		f.logger,
	)
}

func (f *Factory) CreateAPIFallbackHandler() *api.FallbackHandler {
	return api.NewFallbackHandler(
		f.logger,
	)
}

func (f *Factory) CreateWebUserHandler() *web.UserHandler {
	csrfRepo := f.GetStore(CSRFStore)

	return web.NewUserHandler(
		f.CreateUserService(),
		f.CreateCSRFRepo(csrfRepo),
		f.CreateCookieService(),
		f.logger,
	)
}

func (f *Factory) CreateWebFileHandler() *web.FileHandler {
	return web.NewFileHandler(
		f.CreateFileService(),
		f.CreateCookieService(),
		f.logger,
	)
}

func (f *Factory) CreateWebFallbackHandler() *web.FallbackHandler {
	csrfRepo := f.GetStore(CSRFStore)

	return web.NewFallbackHandler(
		f.CreateCSRFRepo(csrfRepo),
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
	fileStore := f.GetStore(FileStore)
	fileRepo := f.CreateFileRepo(fileStore)

	fsStore := f.getFileSystem()

	return service.NewFile(fileRepo, fsStore, *f.logger)
}

// CreateUserService creates a user service.
func (f *Factory) CreateUserService() *service.User {
	userStore := f.GetStore(UserStore)
	userRepo := f.CreateUserRepo(userStore)
	hasher := f.getHasher()
	rawChecker := f.createRawPasswordChecker()

	return service.NewUser(userRepo, hasher, rawChecker, *f.logger)
}

// CreateCookieService creates a cookie service.
func (f *Factory) CreateCookieService() *service.Cookie {
	return service.NewCookie(f.getCookieStore(), *f.logger)
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

func (f *Factory) GetStore(dataType DataType) repo.Store {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.stores[dataType] == nil {
		f.stores[dataType] = f.createStore(dataType)
	}

	return f.stores[dataType]
}

func (f *Factory) getCookieStore() *securecookie.SecureCookie {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.cookieStore != nil {
		return f.cookieStore
	}

	hashKey, err := hex.DecodeString(f.appConfig.CookieHashKey)
	if err != nil {
		panic(err)
	}

	blockKey, err := hex.DecodeString(f.appConfig.CookieBlockKey)
	if err != nil {
		panic(err)
	}

	f.cookieStore = securecookie.New(hashKey, blockKey)

	return f.cookieStore
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

// CreateCSRFRepo creates a CSRF repository.
func (f *Factory) CreateCSRFRepo(csrfStore repo.Store) *repo.CSRF {
	return repo.NewCSRF(csrfStore)
}

func (f *Factory) CreateFileRepo(fileStore repo.Store) *repo.File {
	return repo.NewFile(fileStore)
}

func (f *Factory) CreateUserRepo(userStore repo.Store) *repo.User {
	return repo.NewUser(userStore)
}

func (f *Factory) getHasher() service.PasswordHasher {
	f.mutex.Lock()
	defer f.mutex.Unlock()

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
	f.mutex.Lock()
	defer f.mutex.Unlock()

	return f.logger
}

// GetDisplay returns the display.
func (f *Factory) GetDisplay() cli.Display {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.display == nil {
		f.display = cli.NewStdout()
	}

	return f.display
}

// SetDisplay sets the display for the factory.
func (f *Factory) SetDisplay(display cli.Display) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.display = display
}
