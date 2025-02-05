package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Database DatabaseConfig `toml:"database"`
}

type DatabaseConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Name     string `toml:"name"`
}

func LoadConfig(path string) (*Config, error) {
	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found at: %s", path)
	}

	var conf Config
	if _, err := toml.DecodeFile(path, &conf); err != nil {
		return nil, fmt.Errorf("unable to decode config file: %v", err)
	}

	return &conf, nil
}
