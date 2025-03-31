package configuration

import (
	"fmt"
	"log/slog"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config holds the application configuration settings. The configuration is loaded from
// environment variables.
type Configuration struct {
	DynamoEndpoint string     `env:"DYNAMODB_ENDPOINT,required"`
	Host           string     `env:"HOST,required"`
	Port           string     `env:"PORT,required"`
	LogLevel       slog.Level `env:"LOG_LEVEL,required"`
	ShutdownTimout int        `env:"SHUTDOWN_TIMEOUT,required"`
}

// New loads Configuration from environment variables and a .env file, and returns a
// Config struct or error.
func New() (Configuration, error) {
	// Load values from a .env file and add them to system environment variables.
	// Discard errors coming from this function. This allows us to call this
	// function without a .env file which will by default load values directly
	// from system environment variables.
	_ = godotenv.Load()
	// Once values have been loaded into system env vars, parse those into our
	// configuration struct and validate them returning any errors.
	cfg, err := env.ParseAs[Configuration]()
	if err != nil {
		return Configuration{}, fmt.Errorf("[in configuration.New] failed to parse configuration: %w", err)
	}
	return cfg, nil
}
