package main

import (
	"context"
	"os"
	"time"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"

	"github.com/peteraba/cloudy-files/appconfig"
	"github.com/peteraba/cloudy-files/cli"
	"github.com/peteraba/cloudy-files/compose"
)

const storeEnvKey = "STORE"

const storeTypeS3 = "s3"

func main() {
	start := time.Now()

	ctx := context.Background()

	factory := compose.NewFactory(appconfig.NewConfigFromFile())
	logger := factory.GetLogger()

	if os.Getenv(storeEnvKey) == storeTypeS3 {
		cfg, err := awsConfig.LoadDefaultConfig(ctx)
		if err != nil {
			logger.Error().Err(err).Msg("unable to load SDK config")

			os.Exit(1)
		}

		factory.SetAWS(cfg)
		logger.Info().Msg("AWS SDK config loaded")
	}

	cliApp := cli.NewApp(factory)

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
		logger.Error().Str("command", os.Args[1]).Msg("Unknown command")
	}

	logger.Info().Dur("duration", time.Since(start)).Msg("Execution time")
}
