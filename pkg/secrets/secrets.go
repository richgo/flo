// Package secrets handles credential management and environment variable loading.
package secrets

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Manager manages secrets and environment variables.
type Manager struct {
	envVars map[string]string
}

// NewManager creates a new secrets manager.
func NewManager() *Manager {
	return &Manager{
		envVars: make(map[string]string),
	}
}

// LoadEnvFile loads environment variables from a .env file.
func (m *Manager) LoadEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // .env file is optional
		}
		return fmt.Errorf("failed to open .env file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid format at line %d: %s", lineNum, line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, `"'`)

		m.envVars[key] = value

		// Also set in environment if not already set
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading .env file: %w", err)
	}

	return nil
}

// Get retrieves a secret by key, checking environment first, then loaded .env.
func (m *Manager) Get(key string) string {
	// Check environment first
	if value := os.Getenv(key); value != "" {
		return value
	}

	// Fall back to loaded .env
	return m.envVars[key]
}

// GetRequired retrieves a required secret, returning an error if not found.
func (m *Manager) GetRequired(key string) (string, error) {
	value := m.Get(key)
	if value == "" {
		return "", fmt.Errorf("required secret %q not found", key)
	}
	return value, nil
}

// Set sets a secret value (for testing purposes).
func (m *Manager) Set(key, value string) {
	m.envVars[key] = value
}

// List returns all loaded secret keys (not values).
func (m *Manager) List() []string {
	keys := make([]string, 0, len(m.envVars))
	for key := range m.envVars {
		keys = append(keys, key)
	}
	return keys
}

// Mask returns a masked version of a secret value for display.
func Mask(value string) string {
	if value == "" {
		return "(not set)"
	}
	if len(value) <= 8 {
		return "****"
	}
	return value[:4] + "****" + value[len(value)-4:]
}

// LoadDefault loads .env from the current directory or specified locations.
func LoadDefault() (*Manager, error) {
	m := NewManager()

	// Try to load from current directory
	if err := m.LoadEnvFile(".env"); err != nil {
		return nil, err
	}

	// Also try .flo/.env
	floEnv := filepath.Join(".flo", ".env")
	if err := m.LoadEnvFile(floEnv); err != nil {
		return nil, err
	}

	return m, nil
}

// WellKnownKeys are the environment variables used by Flo.
var WellKnownKeys = []string{
	"CLAUDE_API_KEY",
	"COPILOT_TOKEN",
	"FLO_BACKEND",
	"FLO_MODEL",
}
