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

	setupAws(ctx, factory)

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
		factory.GetLogger().Error().Str("command", os.Args[1]).Msg("Unknown command")
	}

	factory.GetLogger().Info().Dur("duration", time.Since(start)).Msg("Execution time")
}

func setupAws(ctx context.Context, factory *compose.Factory) {
	if os.Getenv(storeEnvKey) != storeTypeS3 {
		return
	}

	cfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		factory.GetLogger().Error().Err(err).Msg("unable to load SDK config")

		os.Exit(1)
	}

	factory.SetAWS(cfg)
	factory.GetLogger().Info().Msg("AWS SDK config loaded")
}
