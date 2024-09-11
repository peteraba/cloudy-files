package main

import (
	"context"
	"os"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"

	"github.com/peteraba/cloudy-files/appconfig"
	"github.com/peteraba/cloudy-files/compose"
)

const storeEnvKey = "STORE"

const storeTypeS3 = "s3"

const (
	commandCli  = "cli"
	commandHTTP = "http"
)

func main() {
	ctx := context.Background()

	factory := compose.NewFactory(appconfig.NewConfigFromFile().Validate())

	setupAws(ctx, factory)

	if len(os.Args) <= 1 {
		factory.GetLogger().Error().Msg("Please provide a command.")
		os.Exit(1)
	}

	switch os.Args[1] {
	case commandCli:
		if len(os.Args) <= 2 {
			factory.GetLogger().Error().Msg("Please provide a subcommand.")
			os.Exit(1)
		}

		args := []string{}
		if len(os.Args) > 2 {
			args = os.Args[3:]
		}

		cliApp := factory.CreateCliApp()
		cliApp.Route(ctx, os.Args[2], args...)
	case commandHTTP:
		router := factory.CreateHTTPApp()
		mux := router.Route()
		router.Start(mux)
	default:
		factory.GetLogger().Info().Str("command", os.Args[0]).Msg("Unknown command.")
		os.Exit(1)
	}
}

func setupAws(ctx context.Context, factory *compose.Factory) {
	if os.Getenv(storeEnvKey) != storeTypeS3 {
		return
	}

	cfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		factory.GetLogger().Error().Err(err).Msg("Unable to load SDK config.")

		os.Exit(1)
	}

	factory.SetAWS(cfg)
	factory.GetLogger().Info().Msg("AWS SDK config loaded.")
}
