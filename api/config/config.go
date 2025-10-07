package config

import (
	"os"
	"strings"
)

// Config holds application configuration
type Config struct {
	Environment string

	// Server configuration
	ServerHostname string
	ServerNames    []string // Alternative names the server might be known by
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	// Get server hostname from environment or use system hostname
	serverHostname := os.Getenv("SERVER_HOSTNAME")
	if serverHostname == "" {
		// Try to get system hostname
		if hostname, err := os.Hostname(); err == nil {
			serverHostname = hostname
		} else {
			serverHostname = "server" // Fallback
		}
	}

	// Get alternative server names from environment
	serverNamesStr := os.Getenv("SERVER_NAMES")
	var serverNames []string
	if serverNamesStr != "" {
		serverNames = strings.Split(serverNamesStr, ",")
		// Trim whitespace
		for i, name := range serverNames {
			serverNames[i] = strings.TrimSpace(name)
		}
	} else {
		// Default alternative names
		serverNames = []string{"server", "dc", "domain-controller"}
	}

	return &Config{
		Environment:    env,
		ServerHostname: serverHostname,
		ServerNames:    serverNames,
	}
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsServerName checks if a given name matches the server hostname or any alternative names
func (c *Config) IsServerName(name string) bool {
	if name == c.ServerHostname {
		return true
	}

	for _, serverName := range c.ServerNames {
		if name == serverName {
			return true
		}
	}

	return false
}

// GetAllServerNames returns all possible names for the server
func (c *Config) GetAllServerNames() []string {
	names := []string{c.ServerHostname}
	names = append(names, c.ServerNames...)
	return names
}
