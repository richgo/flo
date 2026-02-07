package quota

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewTracker(t *testing.T) {
	tracker := New("/tmp/test-quota.json")
	if tracker == nil {
		t.Fatal("Expected tracker, got nil")
	}
	if tracker.usage == nil {
		t.Error("Expected usage map to be initialized")
	}
}

func TestRecordUsage(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "quota.json")
	
	tracker := New(path)
	
	// Record some usage
	if err := tracker.Record("claude", 1000); err != nil {
		t.Fatalf("Failed to record usage: %v", err)
	}
	
	usage, ok := tracker.GetUsage("claude")
	if !ok {
		t.Fatal("Expected usage for claude")
	}
	
	if usage.Requests != 1 {
		t.Errorf("Expected 1 request, got %d", usage.Requests)
	}
	if usage.Tokens != 1000 {
		t.Errorf("Expected 1000 tokens, got %d", usage.Tokens)
	}
}

func TestRecordMultipleRequests(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "quota.json")
	
	tracker := New(path)
	
	// Record multiple requests
	tracker.Record("claude", 500)
	tracker.Record("claude", 750)
	tracker.Record("claude", 250)
	
	usage, ok := tracker.GetUsage("claude")
	if !ok {
		t.Fatal("Expected usage for claude")
	}
	
	if usage.Requests != 3 {
		t.Errorf("Expected 3 requests, got %d", usage.Requests)
	}
	if usage.Tokens != 1500 {
		t.Errorf("Expected 1500 tokens, got %d", usage.Tokens)
	}
}

func TestIsExhausted(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "quota.json")
	
	tracker := New(path)
	tracker.SetLimit("claude", 3)
	
	// Should not be exhausted initially
	if tracker.IsExhausted("claude") {
		t.Error("Should not be exhausted initially")
	}
	
	// Record requests up to limit
	tracker.Record("claude", 100)
	tracker.Record("claude", 100)
	
	if tracker.IsExhausted("claude") {
		t.Error("Should not be exhausted at 2/3 requests")
	}
	
	tracker.Record("claude", 100)
	
	// Should be exhausted now
	if !tracker.IsExhausted("claude") {
		t.Error("Should be exhausted at 3/3 requests")
	}
}

func TestRecordError(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "quota.json")
	
	tracker := New(path)
	
	// Record an error
	if err := tracker.RecordError("copilot", 30*time.Minute); err != nil {
		t.Fatalf("Failed to record error: %v", err)
	}
	
	// Should be exhausted
	if !tracker.IsExhausted("copilot") {
		t.Error("Should be exhausted after error")
	}
	
	usage, ok := tracker.GetUsage("copilot")
	if !ok {
		t.Fatal("Expected usage for copilot")
	}
	
	if !usage.IsExhausted {
		t.Error("Expected IsExhausted to be true")
	}
}

func TestMultipleBackends(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "quota.json")
	
	tracker := New(path)
	
	tracker.Record("claude", 100)
	tracker.Record("copilot", 200)
	tracker.Record("gemini", 300)
	
	usage := tracker.ListUsage()
	if len(usage) != 3 {
		t.Errorf("Expected 3 backends, got %d", len(usage))
	}
	
	if usage["claude"].Tokens != 100 {
		t.Errorf("Expected claude to have 100 tokens, got %d", usage["claude"].Tokens)
	}
	if usage["copilot"].Tokens != 200 {
		t.Errorf("Expected copilot to have 200 tokens, got %d", usage["copilot"].Tokens)
	}
	if usage["gemini"].Tokens != 300 {
		t.Errorf("Expected gemini to have 300 tokens, got %d", usage["gemini"].Tokens)
	}
}

func TestReset(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "quota.json")
	
	tracker := New(path)
	
	tracker.Record("claude", 1000)
	
	if _, ok := tracker.GetUsage("claude"); !ok {
		t.Fatal("Expected usage for claude")
	}
	
	// Reset
	if err := tracker.Reset("claude"); err != nil {
		t.Fatalf("Failed to reset: %v", err)
	}
	
	if _, ok := tracker.GetUsage("claude"); ok {
		t.Error("Should not have usage after reset")
	}
}

func TestResetAll(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "quota.json")
	
	tracker := New(path)
	
	tracker.Record("claude", 100)
	tracker.Record("copilot", 200)
	
	if len(tracker.ListUsage()) != 2 {
		t.Fatal("Expected 2 backends")
	}
	
	// Reset all
	if err := tracker.ResetAll(); err != nil {
		t.Fatalf("Failed to reset all: %v", err)
	}
	
	if len(tracker.ListUsage()) != 0 {
		t.Error("Should have no usage after reset all")
	}
}

func TestPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "quota.json")
	
	// Create tracker and record usage
	tracker1 := New(path)
	tracker1.Record("claude", 500)
	tracker1.Record("copilot", 750)
	
	// Create new tracker and load
	tracker2 := New(path)
	if err := tracker2.Load(); err != nil {
		t.Fatalf("Failed to load: %v", err)
	}
	
	// Verify data persisted
	usage := tracker2.ListUsage()
	if len(usage) != 2 {
		t.Errorf("Expected 2 backends, got %d", len(usage))
	}
	
	if usage["claude"].Tokens != 500 {
		t.Errorf("Expected claude to have 500 tokens, got %d", usage["claude"].Tokens)
	}
	if usage["copilot"].Tokens != 750 {
		t.Errorf("Expected copilot to have 750 tokens, got %d", usage["copilot"].Tokens)
	}
}

func TestWindowReset(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "quota.json")
	
	tracker := New(path)
	tracker.SetWindow(100 * time.Millisecond)
	tracker.SetLimit("claude", 2)
	
	// Record requests
	tracker.Record("claude", 100)
	tracker.Record("claude", 100)
	
	// Should be exhausted
	if !tracker.IsExhausted("claude") {
		t.Error("Should be exhausted at limit")
	}
	
	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)
	
	// Record another request - should reset window
	tracker.Record("claude", 100)
	
	usage, _ := tracker.GetUsage("claude")
	if usage.Requests != 1 {
		t.Errorf("Expected 1 request after window reset, got %d", usage.Requests)
	}
	if usage.Tokens != 100 {
		t.Errorf("Expected 100 tokens after window reset, got %d", usage.Tokens)
	}
}

func TestGetUsageNonExistent(t *testing.T) {
	tracker := New("/tmp/test-quota.json")
	
	_, ok := tracker.GetUsage("nonexistent")
	if ok {
		t.Error("Should not have usage for nonexistent backend")
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "nonexistent.json")
	
	tracker := New(path)
	if err := tracker.Load(); err != nil {
		t.Errorf("Load should not fail for nonexistent file: %v", err)
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "invalid.json")
	
	// Write invalid JSON
	os.WriteFile(path, []byte("not json"), 0644)
	
	tracker := New(path)
	if err := tracker.Load(); err == nil {
		t.Error("Load should fail for invalid JSON")
	}
}
