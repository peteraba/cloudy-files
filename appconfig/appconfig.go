package appconfig

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	StoreAwsBucket      string `env:"STORE_AWS_BUCKET"`
	StoreLocalPath      string `env:"STORE_LOCAL_PATH"      envDefault:"./data"`
	FileSystemAwsBucket string `env:"FILESYSTEM_AWS_BUCKET"`
	FileSystemLocalPath string `env:"FILESYSTEM_LOCAL_PATH" envDefault:"./files"`
}

func NewConfigFromFile(filenames ...string) *Config {
	err := godotenv.Load(filenames...)
	if err != nil {
		panic(fmt.Errorf("error loading .env file: %w", err))
	}

	return NewConfig()
}

func NewConfig() *Config {
	var cfg Config

	err := env.Parse(&cfg)
	if err != nil {
		panic(fmt.Errorf("error parsing configuration: %w", err))
	}

	return &cfg
}
