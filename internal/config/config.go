package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Config represents the application configuration loaded from a TOML file.
type Config struct {
	Database DatabaseConfig `toml:"database"`
}

// DatabaseConfig holds the database connection parameters.
type DatabaseConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Name     string `toml:"name"`
}

// LoadConfig loads and parses the configuration from the specified file path.
func LoadConfig(path string) (*Config, error) {
	// Check if the config file exists and return a wrapped error if not.
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found at %s", path)
		}
		return nil, fmt.Errorf("error checking config file: %w", err)
	}

	var conf Config
	if _, err := toml.DecodeFile(path, &conf); err != nil {
		return nil, fmt.Errorf("unable to decode config file: %w", err)
	}

	return &conf, nil
}
