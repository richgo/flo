package task

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewTask(t *testing.T) {
	task := New("ua-001", "Implement OAuth")

	if task.ID != "ua-001" {
		t.Errorf("expected ID 'ua-001', got '%s'", task.ID)
	}
	if task.Title != "Implement OAuth" {
		t.Errorf("expected Title 'Implement OAuth', got '%s'", task.Title)
	}
	if task.Status != StatusPending {
		t.Errorf("expected Status 'pending', got '%s'", task.Status)
	}
	if task.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if task.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestTaskValidation(t *testing.T) {
	tests := []struct {
		name    string
		task    *Task
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid task",
			task:    New("ua-001", "Valid Task"),
			wantErr: false,
		},
		{
			name:    "empty ID",
			task:    &Task{ID: "", Title: "No ID"},
			wantErr: true,
			errMsg:  "task ID cannot be empty",
		},
		{
			name:    "empty title",
			task:    &Task{ID: "ua-001", Title: ""},
			wantErr: true,
			errMsg:  "task title cannot be empty",
		},
		{
			name:    "invalid status",
			task:    &Task{ID: "ua-001", Title: "Test", Status: "invalid"},
			wantErr: true,
			errMsg:  "invalid status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errMsg)
				} else if !contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestStatusTransitions(t *testing.T) {
	tests := []struct {
		name     string
		from     Status
		to       Status
		wantErr  bool
	}{
		{"pending to in_progress", StatusPending, StatusInProgress, false},
		{"pending to complete", StatusPending, StatusComplete, true},
		{"in_progress to complete", StatusInProgress, StatusComplete, false},
		{"in_progress to failed", StatusInProgress, StatusFailed, false},
		{"complete to pending", StatusComplete, StatusPending, true},
		{"complete to in_progress", StatusComplete, StatusInProgress, true},
		{"failed to pending", StatusFailed, StatusPending, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{
				ID:     "test-001",
				Title:  "Test Task",
				Status: tt.from,
			}
			err := task.SetStatus(tt.to)
			if tt.wantErr && err == nil {
				t.Errorf("expected error for transition %s -> %s", tt.from, tt.to)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error for transition %s -> %s: %v", tt.from, tt.to, err)
			}
			if !tt.wantErr && task.Status != tt.to {
				t.Errorf("expected status %s, got %s", tt.to, task.Status)
			}
		})
	}
}

func TestTaskJSONSerialization(t *testing.T) {
	original := New("ua-001", "Implement OAuth")
	original.Description = "OAuth2 with Google"
	original.Priority = 1
	original.Repo = "android"
	original.Deps = []string{"ua-000"}
	original.SpecRef = "SPEC.md#oauth"

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Unmarshal
	var restored Task
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify fields
	if restored.ID != original.ID {
		t.Errorf("ID mismatch: %s vs %s", restored.ID, original.ID)
	}
	if restored.Title != original.Title {
		t.Errorf("Title mismatch: %s vs %s", restored.Title, original.Title)
	}
	if restored.Description != original.Description {
		t.Errorf("Description mismatch")
	}
	if restored.Status != original.Status {
		t.Errorf("Status mismatch: %s vs %s", restored.Status, original.Status)
	}
	if restored.Priority != original.Priority {
		t.Errorf("Priority mismatch: %d vs %d", restored.Priority, original.Priority)
	}
	if restored.Repo != original.Repo {
		t.Errorf("Repo mismatch: %s vs %s", restored.Repo, original.Repo)
	}
	if len(restored.Deps) != len(original.Deps) || restored.Deps[0] != original.Deps[0] {
		t.Errorf("Deps mismatch")
	}
	if restored.SpecRef != original.SpecRef {
		t.Errorf("SpecRef mismatch")
	}
}

func TestTaskJSONForwardCompatibility(t *testing.T) {
	// JSON with unknown field
	jsonData := `{
		"id": "ua-001",
		"title": "Test",
		"status": "pending",
		"unknown_field": "should be ignored",
		"created_at": "2026-02-05T22:00:00Z",
		"updated_at": "2026-02-05T22:00:00Z"
	}`

	var task Task
	err := json.Unmarshal([]byte(jsonData), &task)
	if err != nil {
		t.Fatalf("should ignore unknown fields: %v", err)
	}
	if task.ID != "ua-001" {
		t.Errorf("expected ID 'ua-001', got '%s'", task.ID)
	}
}

func TestStatusIsValid(t *testing.T) {
	validStatuses := []Status{StatusPending, StatusInProgress, StatusComplete, StatusFailed}
	for _, s := range validStatuses {
		if !s.IsValid() {
			t.Errorf("expected %s to be valid", s)
		}
	}

	invalid := Status("bogus")
	if invalid.IsValid() {
		t.Error("expected 'bogus' to be invalid")
	}
}

func TestTaskUpdateTimestamp(t *testing.T) {
	task := New("ua-001", "Test")
	originalUpdated := task.UpdatedAt

	time.Sleep(10 * time.Millisecond)
	task.SetStatus(StatusInProgress)

	if !task.UpdatedAt.After(originalUpdated) {
		t.Error("expected UpdatedAt to be updated after status change")
	}
}

// Helper
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
