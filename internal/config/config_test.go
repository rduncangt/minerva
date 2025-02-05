package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadConfig_ValidFile checks that a proper config TOML file is parsed without error.
func TestLoadConfig_ValidFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "test-config-valid")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Define a valid TOML
	validTOML := `
[database]
host = "localhost"
port = 5432
user = "minerva_user"
password = "secure_password"
name = "minerva"
`

	// Write it to a temp file
	configPath := filepath.Join(tempDir, "minerva_config.toml")
	if err := os.WriteFile(configPath, []byte(validTOML), 0600); err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	// Attempt to load the config
	conf, loadErr := LoadConfig(configPath)
	if loadErr != nil {
		t.Fatalf("LoadConfig returned an error: %v", loadErr)
	}

	// Verify fields are loaded
	if conf.Database.Host != "localhost" {
		t.Errorf("Expected host=%q, got %q", "localhost", conf.Database.Host)
	}
	if conf.Database.Port != 5432 {
		t.Errorf("Expected port=%d, got %d", 5432, conf.Database.Port)
	}
	if conf.Database.User != "minerva_user" {
		t.Errorf("Expected user=%q, got %q", "minerva_user", conf.Database.User)
	}
	if conf.Database.Password != "secure_password" {
		t.Errorf("Expected password=%q, got %q", "secure_password", conf.Database.Password)
	}
	if conf.Database.Name != "minerva" {
		t.Errorf("Expected name=%q, got %q", "minerva", conf.Database.Name)
	}
}

// TestLoadConfig_FileNotFound checks that an error is returned if the config file is missing.
func TestLoadConfig_FileNotFound(t *testing.T) {
	// Provide a path that doesn’t exist
	_, err := LoadConfig("/path/that/does/not/exist.toml")
	if err == nil {
		t.Fatal("Expected an error for a non-existent file, but got nil")
	}
}

// TestLoadConfig_InvalidTOML checks behavior when TOML is malformed.
func TestLoadConfig_InvalidTOML(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-config-invalid")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Malformed TOML (missing closing bracket, invalid syntax, etc.)
	invalidTOML := `
[database
host = "localhost"
port = 5432
`

	configPath := filepath.Join(tempDir, "bad_config.toml")
	if err := os.WriteFile(configPath, []byte(invalidTOML), 0600); err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	_, loadErr := LoadConfig(configPath)
	if loadErr == nil {
		t.Fatal("Expected an error for invalid TOML, but got nil")
	}
}
