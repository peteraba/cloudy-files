package main

import (
	"fmt"
	"os"
	"time"

	"github.com/peteraba/cloudy-files/cli"
	"github.com/peteraba/cloudy-files/compose"
)

func main() {
	start := time.Now()

	cliApp := cli.NewApp(compose.NewFactory())

	if len(os.Args) <= 1 {
		cliApp.ExitWithHelp("Please provide a command.")
	}

	switch os.Args[1] {
	case "createUser":
		cliApp.CreateUser()
	case "hashPassword":
		cliApp.HashPassword()
	case "login":
		cliApp.Login()
	case "checkPassword":
		cliApp.CheckPassword()
	case "checkPasswordHash":
		cliApp.CheckPasswordHash()
	case "checkSession":
		cliApp.CheckSession()
	case "cleanUp":
		cliApp.CleanUp()
	case "upload":
		cliApp.Upload()
	case "size":
		cliApp.Size()
	default:
		fmt.Println("Unknown command:", os.Args[1])
	}

	fmt.Println("Execution time:", time.Since(start))
}
