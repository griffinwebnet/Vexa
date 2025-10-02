package config

import (
	"os"
)

// Config holds application configuration
type Config struct {
	Environment string
	Debug       bool
	DevMode     bool
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	return &Config{
		Environment: env,
		Debug:       env == "development",
		DevMode:     env == "development",
	}
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.DevMode
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}
