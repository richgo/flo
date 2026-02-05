// Package task provides the core Task data model for EAS.
package task

import (
	"fmt"
	"time"
)

// Status represents the current state of a task.
type Status string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusComplete   Status = "complete"
	StatusFailed     Status = "failed"
)

// IsValid returns true if the status is a known valid status.
func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusInProgress, StatusComplete, StatusFailed:
		return true
	default:
		return false
	}
}

// Task represents a unit of work within a feature.
type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Status      Status    `json:"status"`
	Priority    int       `json:"priority,omitempty"`
	Repo        string    `json:"repo,omitempty"`
	Deps        []string  `json:"deps,omitempty"`
	SpecRef     string    `json:"spec_ref,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// New creates a new Task with the given ID and title.
// Status defaults to pending, timestamps are set automatically.
func New(id, title string) *Task {
	now := time.Now()
	return &Task{
		ID:        id,
		Title:     title,
		Status:    StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Validate checks if the task has valid required fields.
func (t *Task) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}
	if t.Title == "" {
		return fmt.Errorf("task title cannot be empty")
	}
	if t.Status != "" && !t.Status.IsValid() {
		return fmt.Errorf("invalid status: %s", t.Status)
	}
	return nil
}

// validTransitions defines allowed status transitions.
// Key is current status, value is set of allowed next statuses.
var validTransitions = map[Status]map[Status]bool{
	StatusPending: {
		StatusInProgress: true,
	},
	StatusInProgress: {
		StatusComplete: true,
		StatusFailed:   true,
	},
	StatusComplete: {
		// Terminal state - no transitions allowed
	},
	StatusFailed: {
		StatusPending: true, // Allow retry
	},
}

// SetStatus changes the task status if the transition is valid.
// Returns an error if the transition is not allowed.
func (t *Task) SetStatus(newStatus Status) error {
	if t.Status == newStatus {
		return nil // No change
	}

	allowed, ok := validTransitions[t.Status]
	if !ok {
		return fmt.Errorf("unknown current status: %s", t.Status)
	}

	if !allowed[newStatus] {
		return fmt.Errorf("invalid status transition: %s -> %s", t.Status, newStatus)
	}

	t.Status = newStatus
	t.UpdatedAt = time.Now()
	return nil
}

// IsReady returns true if the task is pending and could be started.
// Note: This doesn't check dependencies - use Registry.IsReady() for that.
func (t *Task) IsReady() bool {
	return t.Status == StatusPending
}

// IsComplete returns true if the task has completed successfully.
func (t *Task) IsComplete() bool {
	return t.Status == StatusComplete
}

// IsTerminal returns true if the task is in a terminal state (complete or failed).
func (t *Task) IsTerminal() bool {
	return t.Status == StatusComplete || t.Status == StatusFailed
}
