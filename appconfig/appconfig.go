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
	CookieHashKey       string `env:"COOKIE_HASH_KEY"       envDefault:"0dd6cd4813db6b708e91c381c4551ac50dc57e486432d01b52220c7aa77083fa"`
	CookieBlockKey      string `env:"COOKIE_BLOCK_KEY"      envDefault:"1dad12d8b9a34a397dc6b6fdf193a868b2a709dbb0646f43bd96db79155818eb"`
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

func (c *Config) Validate() *Config {
	if c.CookieHashKey == "0dd6cd4813db6b708e91c381c4551ac50dc57e486432d01b52220c7aa77083fa" {
		panic("COOKIE_HASH_KEY is required and can't be default.")
	}

	if c.CookieBlockKey == "1dad12d8b9a34a397dc6b6fdf193a868b2a709dbb0646f43bd96db79155818eb" {
		panic("COOKIE_BLOCK_KEY is required and can't be default.")
	}

	return c
}
