package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/peteraba/cloudy-files/compose"
	"github.com/peteraba/cloudy-files/util"
)

// App represents the command line interface.
type App struct{}

// NewApp creates a new App instance.
func NewApp() *App {
	return &App{}
}

// Help prints the help message.
func (a *App) Help() {
	fmt.Println("TODO...")
}

// HashPassword hashes a password.
func (a *App) HashPassword() {
	if len(os.Args) <= 2 {
		a.ExitWithHelp("Please provide a password to hashPassword")
	}

	password := os.Args[2]

	userService := compose.CreateUserService()

	hash, err := userService.HashPassword(password)
	if err != nil {
		a.Exit("Failed to hash password", err)
	}

	fmt.Println("Password hash:", hash)
}

// Login logs in a user.
func (a *App) Login() {
	if len(os.Args) <= 3 {
		a.ExitWithHelp("Please provide a user name and a password to log in with")
	}

	userName := os.Args[2]
	pass := os.Args[3]

	userService := compose.CreateUserService()

	hash, err := userService.Login(userName, pass)
	if err != nil {
		a.Exit("Login failed", err)
	}

	fmt.Println("Session started:", hash)
}

// CheckPassword checks if a password matches the password hash stored for a user.
func (a *App) CheckPassword() {
	if len(os.Args) <= 3 {
		a.ExitWithHelp("Please provide a user name and a password to check")
	}

	userName := os.Args[2]
	pass := os.Args[3]

	userService := compose.CreateUserService()

	err := userService.CheckPassword(userName, pass)
	if err != nil {
		a.Exit("Password does not match", err)
	}

	fmt.Println("Password matches")
}

// CheckPasswordHash checks if a password matches a password hash.
func (a *App) CheckPasswordHash() {
	if len(os.Args) <= 3 {
		a.ExitWithHelp("Please provide a password and a hashPassword to check")
	}

	password := os.Args[2]
	hash := os.Args[3]

	userService := compose.CreateUserService()

	err := userService.CheckHash(password, hash)
	if err != nil {
		a.Exit("Password does not match", err)
	}

	fmt.Println("Password matches")
}

// StartSession starts a session for a user.
func (a *App) StartSession() {
	if len(os.Args) <= 2 {
		a.ExitWithHelp("Please provide a name to start a session for")
	}

	sessionService := compose.CreateSessionService()

	hash, err := sessionService.Start(os.Args[2])
	if err != nil {
		a.Exit("Session could not be started", err)
	}

	fmt.Println(hash)
}

// CheckSession checks if a session exists.
func (a *App) CheckSession() {
	if len(os.Args) <= 3 {
		a.ExitWithHelp("Please provide a name and a hashPassword to check")
	}

	sessionService := compose.CreateSessionService()

	ok, err := sessionService.Check(os.Args[2], os.Args[3])
	if err != nil {
		a.Exit("Session does not exist", err)
	}

	fmt.Println(ok)
}

// CleanUp cleans up all sessions.
func (a *App) CleanUp() {
	sessionService := compose.CreateSessionService()

	err := sessionService.CleanUp()
	if err != nil {
		a.Exit("Session cleanup failed", err)
	}

	fmt.Println("Cleaned up")
}

// CreateUser creates a user.
func (a *App) CreateUser() {
	if len(os.Args) <= 4 {
		a.ExitWithHelp("Please provide at least name, email, password to create a user")
	}

	userService := compose.CreateUserService()

	name := os.Args[2]
	email := os.Args[3]
	password := os.Args[4]
	isAdmin := false

	switch strings.ToLower(os.Args[5]) {
	case "true", "1", "yes", "y":
		isAdmin = true
	}

	var access []string
	if len(os.Args) >= 6 {
		access = os.Args[6:]
	}

	err := userService.Create(name, email, password, isAdmin, access)
	if err != nil {
		a.Exit("Session cleanup failed", err)
	}

	fmt.Println("Cleaned up")
}

// Upload uploads a file.
func (a *App) Upload() {
	if len(os.Args) <= 3 {
		a.ExitWithHelp("Please provide the path of the file to store and at least one access label")
	}

	filePath := os.Args[2]

	stats, err := os.Stat(filePath)
	if err != nil {
		a.Exit("File could not be found", err)
	}

	var access []string
	if len(os.Args) >= 3 {
		access = os.Args[3:]
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		a.Exit("File could not be read", err)
	}

	fileService := compose.CreateFileService()

	err = fileService.Upload(stats.Name(), data, access)
	if err != nil {
		a.Exit("File could not be stored", err)
	}

	fmt.Println("File stored")
}

// Size retrieves a file and displays its size.
func (a *App) Size() {
	if len(os.Args) <= 2 {
		a.ExitWithHelp("Please provide the path of the file to check the size of")
	}

	filePath := os.Args[2]

	stats, err := os.Stat(filePath)
	if err != nil {
		a.Exit("File could not be found", err)
	}

	var access []string
	if len(os.Args) >= 3 {
		access = os.Args[3:]
	}

	fileService := compose.CreateFileService()

	data, err := fileService.Retrieve(stats.Name(), access)
	if err != nil {
		a.Exit("File could not be read", err)
	}

	fileSize := util.FileSizeFromSize(len(data))

	fmt.Println("File size:", fileSize.String())
}

// ExitWithHelp exits the application with a help message.
func (a *App) ExitWithHelp(msg string) {
	fmt.Println("Result:", msg)
	a.Help()
	os.Exit(1)
}

// Exit exits the application after displaying a message and an error.
func (a *App) Exit(msg string, err error) {
	fmt.Println("Result:", msg)
	fmt.Println("Error:", err)
	os.Exit(1)
}
