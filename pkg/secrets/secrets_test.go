package secrets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManager_LoadEnvFile(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantVars    map[string]string
		wantError   bool
	}{
		{
			name: "valid env file",
			content: `CLAUDE_API_KEY=sk-test-123
COPILOT_TOKEN=ghp-test-456
FLO_BACKEND=claude`,
			wantVars: map[string]string{
				"CLAUDE_API_KEY": "sk-test-123",
				"COPILOT_TOKEN":  "ghp-test-456",
				"FLO_BACKEND":    "claude",
			},
			wantError: false,
		},
		{
			name: "env file with comments",
			content: `# API Keys
CLAUDE_API_KEY=sk-test-123
# Token
COPILOT_TOKEN=ghp-test-456`,
			wantVars: map[string]string{
				"CLAUDE_API_KEY": "sk-test-123",
				"COPILOT_TOKEN":  "ghp-test-456",
			},
			wantError: false,
		},
		{
			name: "env file with quotes",
			content: `CLAUDE_API_KEY="sk-test-123"
COPILOT_TOKEN='ghp-test-456'`,
			wantVars: map[string]string{
				"CLAUDE_API_KEY": "sk-test-123",
				"COPILOT_TOKEN":  "ghp-test-456",
			},
			wantError: false,
		},
		{
			name: "env file with empty lines",
			content: `
CLAUDE_API_KEY=sk-test-123

COPILOT_TOKEN=ghp-test-456

`,
			wantVars: map[string]string{
				"CLAUDE_API_KEY": "sk-test-123",
				"COPILOT_TOKEN":  "ghp-test-456",
			},
			wantError: false,
		},
		{
			name:      "invalid format",
			content:   `INVALID_LINE_WITHOUT_EQUALS`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary .env file
			tmpDir := t.TempDir()
			envPath := filepath.Join(tmpDir, ".env")
			
			if err := os.WriteFile(envPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Clear environment variables before test
			for key := range tt.wantVars {
				os.Unsetenv(key)
			}

			m := NewManager()
			err := m.LoadEnvFile(envPath)

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check loaded variables
			for key, wantValue := range tt.wantVars {
				gotValue := m.Get(key)
				if gotValue != wantValue {
					t.Errorf("Get(%q) = %q, want %q", key, gotValue, wantValue)
				}
			}

			// Clean up environment
			for key := range tt.wantVars {
				os.Unsetenv(key)
			}
		})
	}
}

func TestManager_Get(t *testing.T) {
	m := NewManager()

	// Test loaded value
	m.Set("TEST_KEY", "test_value")
	if got := m.Get("TEST_KEY"); got != "test_value" {
		t.Errorf("Get(TEST_KEY) = %q, want %q", got, "test_value")
	}

	// Test environment variable takes precedence
	os.Setenv("TEST_KEY", "env_value")
	defer os.Unsetenv("TEST_KEY")

	if got := m.Get("TEST_KEY"); got != "env_value" {
		t.Errorf("Get(TEST_KEY) = %q, want %q (env should take precedence)", got, "env_value")
	}

	// Test non-existent key
	if got := m.Get("NONEXISTENT"); got != "" {
		t.Errorf("Get(NONEXISTENT) = %q, want empty string", got)
	}
}

func TestManager_GetRequired(t *testing.T) {
	m := NewManager()
	m.Set("EXISTING_KEY", "value")

	// Test existing key
	value, err := m.GetRequired("EXISTING_KEY")
	if err != nil {
		t.Errorf("GetRequired(EXISTING_KEY) unexpected error: %v", err)
	}
	if value != "value" {
		t.Errorf("GetRequired(EXISTING_KEY) = %q, want %q", value, "value")
	}

	// Test non-existent key
	_, err = m.GetRequired("NONEXISTENT_KEY")
	if err == nil {
		t.Error("GetRequired(NONEXISTENT_KEY) expected error, got nil")
	}
}

func TestManager_List(t *testing.T) {
	m := NewManager()
	m.Set("KEY1", "value1")
	m.Set("KEY2", "value2")
	m.Set("KEY3", "value3")

	keys := m.List()
	if len(keys) != 3 {
		t.Errorf("List() returned %d keys, want 3", len(keys))
	}

	// Check all keys are present
	keyMap := make(map[string]bool)
	for _, key := range keys {
		keyMap[key] = true
	}

	for _, expectedKey := range []string{"KEY1", "KEY2", "KEY3"} {
		if !keyMap[expectedKey] {
			t.Errorf("List() missing key %q", expectedKey)
		}
	}
}

func TestMask(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "empty value",
			value: "",
			want:  "(not set)",
		},
		{
			name:  "short value",
			value: "12345",
			want:  "****",
		},
		{
			name:  "medium value",
			value: "123456789",
			want:  "1234****6789",
		},
		{
			name:  "long value",
			value: "sk-test-1234567890abcdef",
			want:  "sk-t****cdef",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Mask(tt.value)
			if got != tt.want {
				t.Errorf("Mask(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestLoadDefault(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Create .env file
	envContent := `CLAUDE_API_KEY=sk-test-123
COPILOT_TOKEN=ghp-test-456`
	if err := os.WriteFile(".env", []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}

	// Load default
	m, err := LoadDefault()
	if err != nil {
		t.Errorf("LoadDefault() unexpected error: %v", err)
	}

	// Verify loaded variables
	if got := m.Get("CLAUDE_API_KEY"); got != "sk-test-123" {
		t.Errorf("CLAUDE_API_KEY = %q, want %q", got, "sk-test-123")
	}

	// Clean up
	os.Unsetenv("CLAUDE_API_KEY")
	os.Unsetenv("COPILOT_TOKEN")
}

func TestWellKnownKeys(t *testing.T) {
	expectedKeys := []string{
		"CLAUDE_API_KEY",
		"COPILOT_TOKEN",
		"FLO_BACKEND",
		"FLO_MODEL",
	}

	if len(WellKnownKeys) != len(expectedKeys) {
		t.Errorf("WellKnownKeys has %d keys, want %d", len(WellKnownKeys), len(expectedKeys))
	}

	keyMap := make(map[string]bool)
	for _, key := range WellKnownKeys {
		keyMap[key] = true
	}

	for _, expected := range expectedKeys {
		if !keyMap[expected] {
			t.Errorf("WellKnownKeys missing %q", expected)
		}
	}
}
