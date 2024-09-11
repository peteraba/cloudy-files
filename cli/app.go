package cli

import (
	"context"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/service"
	"github.com/peteraba/cloudy-files/util"
)

// App represents the command line interface.
type App struct {
	userService *service.User
	fileService *service.File
	display     Display
	logger      *log.Logger
	help        string
}

const Help = "TODO..."

// NewApp creates a new App instance.
func NewApp(userService *service.User, fileService *service.File, display Display, logger *log.Logger) *App {
	return &App{
		userService: userService,
		fileService: fileService,
		display:     display,
		logger:      logger,
		help:        Help,
	}
}

// Route routes the command to the appropriate method.
func (a *App) Route(ctx context.Context, subCommand string, args ...string) {
	start := time.Now()

	switch subCommand {
	case "createUser":
		a.CreateUser(ctx, args...)
	case "hashPassword":
		a.HashPassword(ctx, args...)
	case "login":
		a.Login(ctx, args...)
	case "checkPassword":
		a.CheckPassword(ctx, args...)
	case "checkPasswordHash":
		a.CheckPasswordHash(ctx, args...)
	case "upload":
		a.Upload(ctx, args...)
	case "size":
		a.Size(ctx, args...)
	case "cookieKey":
		a.CookieKey(args...)
	default:
		a.display.ExitWithHelp("Unknown subcommand: "+subCommand, a.help)
	}

	a.logger.Info().Dur("duration", time.Since(start)).Msg("Execution time")
}

// HashPassword hashes a password.
func (a *App) HashPassword(ctx context.Context, args ...string) {
	if len(args) < 1 {
		a.display.ExitWithHelp("Please provide a password.", a.help)
	}

	password := args[0]

	hash, err := a.userService.HashPassword(ctx, password)
	if err != nil {
		a.display.Exit("Failed to hash password.", err)
	}

	a.display.Println("Hashed password:", hash)
}

// Login logs in a user.
func (a *App) Login(ctx context.Context, args ...string) {
	if len(args) < 2 {
		a.display.ExitWithHelp("Please provide a user name and a password to log in with.", a.help)
	}

	userName := args[0]
	pass := args[1]

	sessionModel, err := a.userService.Login(ctx, userName, pass)
	if err != nil {
		a.display.Exit("Login failed.", err)
	}

	a.display.Println("Session started:", sessionModel.Name)
}

// CheckPassword checks if a password matches the password hash stored for a user.
func (a *App) CheckPassword(ctx context.Context, args ...string) {
	if len(args) < 1 {
		a.display.ExitWithHelp("Please provide a user name and a password to check", a.help)
	}

	userName := args[0]
	pass := args[1]

	err := a.userService.CheckPassword(ctx, userName, pass)
	if err != nil {
		a.display.Exit("Password received does not match the user password.", err)
	}

	a.display.Println("Password matches.")
}

// CheckPasswordHash checks if a password matches a password hash.
func (a *App) CheckPasswordHash(ctx context.Context, args ...string) {
	if len(args) < 2 {
		a.display.ExitWithHelp("Please provide a password and a hashPassword to check.", a.help)
	}

	password := args[0]
	hash := args[1]

	err := a.userService.CheckPasswordHash(ctx, password, hash)
	if err != nil {
		a.display.Exit("Password does not match the hash received.", err)
	}

	a.display.Println("Password matches the hash received.")
}

// CreateUser creates a user.
func (a *App) CreateUser(ctx context.Context, args ...string) {
	if len(args) < 4 {
		a.display.ExitWithHelp("Please provide at least name, email, password to create a user.", a.help)
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

	userModel, err := a.userService.Create(ctx, name, email, password, isAdmin, access)
	if err != nil {
		a.display.Exit("User creation failed.", err)
	}

	a.display.Println("User created:", userModel.Name)
}

// Upload uploads a file.
func (a *App) Upload(ctx context.Context, args ...string) {
	if len(args) < 1 {
		a.display.ExitWithHelp("Please provide the path of the file to store and at least one access label.", a.help)
	}

	filePath := args[0]

	stats, err := os.Stat(filePath)
	if err != nil {
		a.display.Exit("File could not be found.", err)
	}

	var access []string
	if len(args) >= 1 {
		access = args[1:]
	}

	// Note: Testing this case is very tricky (if not impossible) due to the os.Stat having to pass.
	data, err := os.ReadFile(filePath)
	if err != nil {
		a.display.Exit("File could not be read.", err)
	}

	fileModel, err := a.fileService.Upload(ctx, stats.Name(), data, access)
	if err != nil {
		a.display.Exit("File could not be stored.", err)
	}

	a.display.Println("File stored:", fileModel.Name)
}

// Size retrieves a file and displays its size.
func (a *App) Size(ctx context.Context, args ...string) {
	if len(args) < 1 {
		a.display.ExitWithHelp("Please provide the path of the file to check the size of.", a.help)
	}

	filePath := args[0]

	var access []string
	if len(args) >= 1 {
		access = args[1:]
	}

	data, err := a.fileService.Retrieve(ctx, filePath, access)
	if err != nil {
		a.display.Exit("File could not be read: "+filePath+", err:", err)
	}

	fileSize := util.FileSizeFromSize(len(data))

	a.display.Println("File size:", fileSize.String())
}

// CookieKey generates a new cookie key.
func (a *App) CookieKey(args ...string) {
	length := 32

	if len(args) > 0 {
		l, err := strconv.Atoi(args[0])
		if err != nil {
			a.display.Exit("Invalid length:", err)

			return
		}

		if l > 0 {
			length = l
		}
	}

	key := securecookie.GenerateRandomKey(length)

	a.display.Println("Key generated:", hex.EncodeToString(key))
}
