# Task Component Specification

## Overview

A Task represents a unit of work within a feature. Tasks form a directed acyclic graph (DAG) via dependencies.

## Data Model

```go
type Task struct {
    ID          string    // Unique identifier (e.g., "ua-001")
    Title       string    // Human-readable title
    Description string    // Detailed description
    Status      Status    // pending | in_progress | complete | failed
    Priority    int       // 0 = highest
    Repo        string    // Target repository name
    Deps        []string  // Task IDs this depends on
    SpecRef     string    // Reference to SPEC.md section
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type Status string

const (
    StatusPending    Status = "pending"
    StatusInProgress Status = "in_progress"
    StatusComplete   Status = "complete"
    StatusFailed     Status = "failed"
)
```

## Acceptance Criteria

### Task Creation
- [ ] Can create a task with required fields (ID, Title)
- [ ] ID must be non-empty and unique within registry
- [ ] Status defaults to "pending" if not specified
- [ ] CreatedAt/UpdatedAt set automatically

### Task Validation
- [ ] Returns error if ID is empty
- [ ] Returns error if Title is empty
- [ ] Returns error if Status is invalid
- [ ] Deps must reference valid task IDs (validated by registry)

### Status Transitions
- [ ] pending → in_progress: allowed
- [ ] pending → complete: NOT allowed (must go through in_progress)
- [ ] in_progress → complete: allowed (only if deps complete)
- [ ] in_progress → failed: allowed
- [ ] complete → *: NOT allowed (terminal state)
- [ ] failed → pending: allowed (retry)

### JSON Serialization
- [ ] Can marshal Task to JSON
- [ ] Can unmarshal JSON to Task
- [ ] Unknown fields ignored (forward compatibility)
