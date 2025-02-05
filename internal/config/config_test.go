package config

import (
	"os"
	"path/filepath"
	"testing"
)

// createTempConfigFile creates a temporary directory and writes the given
// content to a config file. It returns the temp directory and the config file path.
func createTempConfigFile(t *testing.T, content string) (string, string) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "test-config")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	configPath := filepath.Join(tempDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}
	return tempDir, configPath
}

func TestLoadConfig_ValidFile(t *testing.T) {
	validTOML := `
[database]
host = "localhost"
port = 5432
user = "minerva_user"
password = "secure_password"
name = "minerva"
`
	tempDir, configPath := createTempConfigFile(t, validTOML)
	defer os.RemoveAll(tempDir)

	conf, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig returned an error: %v", err)
	}

	if conf.Database.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got %q", conf.Database.Host)
	}
	if conf.Database.Port != 5432 {
		t.Errorf("Expected port 5432, got %d", conf.Database.Port)
	}
	if conf.Database.User != "minerva_user" {
		t.Errorf("Expected user 'minerva_user', got %q", conf.Database.User)
	}
	if conf.Database.Password != "secure_password" {
		t.Errorf("Expected password 'secure_password', got %q", conf.Database.Password)
	}
	if conf.Database.Name != "minerva" {
		t.Errorf("Expected name 'minerva', got %q", conf.Database.Name)
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("/path/that/does/not/exist.toml")
	if err == nil {
		t.Fatal("Expected an error for a non-existent file, but got nil")
	}
}

func TestLoadConfig_InvalidTOML(t *testing.T) {
	invalidTOML := `
[database
host = "localhost"
port = 5432
`
	tempDir, configPath := createTempConfigFile(t, invalidTOML)
	defer os.RemoveAll(tempDir)

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Fatal("Expected an error for invalid TOML, but got nil")
	}
}
