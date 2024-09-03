package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/service"
	"github.com/peteraba/cloudy-files/util"
)

// App represents the command line interface.
type App struct {
	sessionService *service.Session
	userService    *service.User
	fileService    *service.File
	logger         *log.Logger
}

// NewApp creates a new App instance.
func NewApp(sessionService *service.Session, userService *service.User, fileService *service.File, logger *log.Logger) *App {
	return &App{
		sessionService: sessionService,
		userService:    userService,
		fileService:    fileService,
		logger:         logger,
	}
}

// Route routes the command to the appropriate method.
func (a *App) Route(subCommand string, args ...string) {
	start := time.Now()

	switch subCommand {
	case "createUser":
		a.CreateUser(args...)
	case "hashPassword":
		a.HashPassword(args...)
	case "login":
		a.Login(args...)
	case "checkPassword":
		a.CheckPassword(args...)
	case "checkPasswordHash":
		a.CheckPasswordHash(args...)
	case "checkSession":
		a.CheckSession(args...)
	case "cleanUp":
		a.CleanUp()
	case "upload":
		a.Upload(args...)
	case "size":
		a.Size(args...)
	default:
		a.logger.Error().Str("subCommand", subCommand).Msg("Unknown subCommand")
	}

	a.logger.Info().Dur("duration", time.Since(start)).Msg("Execution time")
}

// Help prints the help message.
func (a *App) Help() {
	fmt.Println("TODO...")
}

// HashPassword hashes a password.
func (a *App) HashPassword(args ...string) {
	if len(args) < 1 {
		a.ExitWithHelp("Please provide a password to hashPassword")
	}

	password := args[0]

	hash, err := a.userService.HashPassword(context.Background(), password)
	if err != nil {
		a.Exit("Failed to hash password", err)
	}

	a.logger.Info().Str("hash", hash).Msg("Password hashed")
}

// Login logs in a user.
func (a *App) Login(args ...string) {
	if len(args) < 2 {
		a.ExitWithHelp("Please provide a user name and a password to log in with")
	}

	userName := args[0]
	pass := args[1]

	sessionModel, err := a.userService.Login(context.Background(), userName, pass)
	if err != nil {
		a.Exit("Login failed", err)
	}

	a.logger.Info().Str("hash", sessionModel.Hash).Msg("Session started")
}

// CheckPassword checks if a password matches the password hash stored for a user.
func (a *App) CheckPassword(args ...string) {
	if len(args) < 1 {
		a.ExitWithHelp("Please provide a user name and a password to check")
	}

	userName := args[0]
	pass := args[1]

	err := a.userService.CheckPassword(context.Background(), userName, pass)
	if err != nil {
		a.Exit("Password does not match", err)
	}

	a.logger.Info().Msg("Password matches")
}

// CheckPasswordHash checks if a password matches a password hash.
func (a *App) CheckPasswordHash(args ...string) {
	if len(args) < 2 {
		a.ExitWithHelp("Please provide a password and a hashPassword to check")
	}

	password := args[0]
	hash := args[1]

	err := a.userService.CheckPasswordHash(context.Background(), password, hash)
	if err != nil {
		a.Exit("Password does not match", err)
	}

	a.logger.Info().Msg("Password matches")
}

// CheckSession checks if a session exists.
func (a *App) CheckSession(args ...string) {
	if len(args) < 2 {
		a.ExitWithHelp("Please provide a name and a hashPassword to check")
	}

	name := args[0]
	hash := args[1]

	ok, err := a.sessionService.Check(context.Background(), name, hash)
	if err != nil {
		a.Exit("Session does not exist", err)
	}

	a.logger.Info().Bool("ok", ok).Msg("Session checked")
}

// CleanUp cleans up all sessions.
func (a *App) CleanUp() {
	err := a.sessionService.CleanUp(context.Background())
	if err != nil {
		a.Exit("Session cleanup failed", err)
	}

	a.logger.Info().Msg("Cleaned up")
}

// CreateUser creates a user.
func (a *App) CreateUser(args ...string) {
	if len(args) < 4 {
		a.ExitWithHelp("Please provide at least name, email, password to create a user")
	}

	name := args[0]
	email := args[1]
	password := args[2]
	isAdmin := false

	switch strings.ToLower(args[3]) {
	case "true", "1", "yes", "y":
		isAdmin = true
	}

	var access []string
	if len(args) >= 4 {
		access = args[4:]
	}

	userModel, err := a.userService.Create(context.Background(), name, email, password, isAdmin, access)
	if err != nil {
		a.Exit("User creation failed", err)
	}

	a.logger.Info().Str("name", userModel.Name).Msg("User created")
}

// Upload uploads a file.
func (a *App) Upload(args ...string) {
	if len(args) < 1 {
		a.ExitWithHelp("Please provide the path of the file to store and at least one access label")
	}

	filePath := args[0]

	stats, err := os.Stat(filePath)
	if err != nil {
		a.Exit("File could not be found", err)
	}

	var access []string
	if len(args) >= 1 {
		access = args[1:]
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		a.Exit("File could not be read", err)
	}

	fileModel, err := a.fileService.Upload(context.Background(), stats.Name(), data, access)
	if err != nil {
		a.Exit("File could not be stored", err)
	}

	a.logger.Info().Str("name", fileModel.Name).Msg("File stored")
}

// Size retrieves a file and displays its size.
func (a *App) Size(args ...string) {
	if len(args) < 1 {
		a.ExitWithHelp("Please provide the path of the file to check the size of")
	}

	filePath := args[0]

	stats, err := os.Stat(filePath)
	if err != nil {
		a.Exit("File could not be found", err)
	}

	var access []string
	if len(args) >= 1 {
		access = args[1:]
	}

	data, err := a.fileService.Retrieve(context.Background(), stats.Name(), access)
	if err != nil {
		a.Exit("File could not be read", err)
	}

	fileSize := util.FileSizeFromSize(len(data))

	a.logger.Info().
		Str("name", stats.Name()).
		Str("size", fileSize.String()).
		Msg("File size retrieved")
}

// ExitWithHelp exits the application with a help message.
func (a *App) ExitWithHelp(msg string) {
	a.logger.Error().Msg(msg)

	a.Help()

	os.Exit(1)
}

// Exit exits the application after displaying a message and an error.
func (a *App) Exit(msg string, err error) {
	a.logger.Error().Err(err).Msg(msg)

	os.Exit(1)
}
