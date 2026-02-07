package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestAuditInit(t *testing.T) {
	tmpDir := t.TempDir()
	
	err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Close()
	
	// Verify audit log file was created
	auditPath := filepath.Join(tmpDir, ".flo", "audit.log")
	if _, err := os.Stat(auditPath); os.IsNotExist(err) {
		t.Errorf("Audit log file was not created at %s", auditPath)
	}
}

func TestAuditLog(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Reset for testing - create new once and clear logger
	once = sync.Once{}
	defaultLogger = nil
	
	err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Close()
	
	// Log some events
	Info("test.operation", "Test info message", map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	})
	
	Warn("test.operation", "Test warning message", nil)
	
	Error("test.operation", "Test error message", map[string]interface{}{
		"error": "something went wrong",
	})
	
	// Close to flush
	if err := Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	
	// Read and verify events
	auditPath := filepath.Join(tmpDir, ".flo", "audit.log")
	data, err := os.ReadFile(auditPath)
	if err != nil {
		t.Fatalf("Failed to read audit log: %v", err)
	}
	
	// Parse events
	lines := []byte{}
	events := []Event{}
	for _, b := range data {
		if b == '\n' {
			if len(lines) > 0 {
				var event Event
				if err := json.Unmarshal(lines, &event); err != nil {
					t.Errorf("Failed to parse event: %v", err)
				}
				events = append(events, event)
				lines = []byte{}
			}
		} else {
			lines = append(lines, b)
		}
	}
	
	if len(events) != 3 {
		t.Errorf("Expected 3 events, got %d", len(events))
	}
	
	// Verify first event
	if events[0].Level != LevelInfo {
		t.Errorf("Expected INFO level, got %s", events[0].Level)
	}
	if events[0].Operation != "test.operation" {
		t.Errorf("Expected operation 'test.operation', got %s", events[0].Operation)
	}
	if events[0].Message != "Test info message" {
		t.Errorf("Expected message 'Test info message', got %s", events[0].Message)
	}
	if events[0].Details["key1"] != "value1" {
		t.Errorf("Expected detail key1='value1', got %v", events[0].Details["key1"])
	}
	
	// Verify second event
	if events[1].Level != LevelWarn {
		t.Errorf("Expected WARN level, got %s", events[1].Level)
	}
	
	// Verify third event
	if events[2].Level != LevelError {
		t.Errorf("Expected ERROR level, got %s", events[2].Level)
	}
}

func TestAuditLogWithoutInit(t *testing.T) {
	// Reset logger
	once = sync.Once{}
	defaultLogger = nil
	
	// Should not panic when logging without init
	Info("test", "message", nil)
	Warn("test", "message", nil)
	Error("test", "message", nil)
}

func TestAuditEventTimestamp(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Reset for testing
	once = sync.Once{}
	defaultLogger = nil
	
	err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Close()
	
	beforeTime := time.Now()
	Info("test", "message", nil)
	afterTime := time.Now()
	
	// Close to flush
	Close()
	
	// Read event
	auditPath := filepath.Join(tmpDir, ".flo", "audit.log")
	data, err := os.ReadFile(auditPath)
	if err != nil {
		t.Fatalf("Failed to read audit log: %v", err)
	}
	
	var event Event
	if err := json.Unmarshal(data[:len(data)-1], &event); err != nil {
		t.Fatalf("Failed to parse event: %v", err)
	}
	
	// Verify timestamp is within range
	if event.Timestamp.Before(beforeTime) || event.Timestamp.After(afterTime) {
		t.Errorf("Event timestamp %v is not within expected range [%v, %v]", 
			event.Timestamp, beforeTime, afterTime)
	}
}
