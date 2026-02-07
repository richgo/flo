// Package quota manages usage tracking and quota limits for AI backends.
package quota

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Usage tracks usage metrics for a backend.
type Usage struct {
	Backend      string    `json:"backend"`
	Requests     int       `json:"requests"`
	Tokens       int       `json:"tokens"`
	LastRequest  time.Time `json:"last_request"`
	WindowStart  time.Time `json:"window_start"`
	IsExhausted  bool      `json:"is_exhausted"`
	RetryAfter   time.Time `json:"retry_after,omitempty"`
}

// Tracker manages quota tracking for multiple backends.
type Tracker struct {
	mu      sync.RWMutex
	usage   map[string]*Usage
	path    string
	limits  map[string]int // Backend -> requests per window
	window  time.Duration  // Time window for limits
}

// New creates a new quota tracker.
func New(dataPath string) *Tracker {
	return &Tracker{
		usage:  make(map[string]*Usage),
		path:   dataPath,
		limits: make(map[string]int),
		window: time.Hour, // Default 1 hour window
	}
}

// SetLimit sets the request limit for a backend.
func (t *Tracker) SetLimit(backend string, requests int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.limits[backend] = requests
}

// SetWindow sets the time window for quota tracking.
func (t *Tracker) SetWindow(d time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.window = d
}

// Record records a request and token usage for a backend.
func (t *Tracker) Record(backend string, tokens int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	
	usage, ok := t.usage[backend]
	if !ok {
		usage = &Usage{
			Backend:     backend,
			WindowStart: now,
		}
		t.usage[backend] = usage
	}

	// Reset window if expired
	if now.Sub(usage.WindowStart) > t.window {
		usage.Requests = 0
		usage.Tokens = 0
		usage.WindowStart = now
		usage.IsExhausted = false
	}

	usage.Requests++
	usage.Tokens += tokens
	usage.LastRequest = now

	// Check if exhausted
	if limit, ok := t.limits[backend]; ok {
		if usage.Requests >= limit {
			usage.IsExhausted = true
			usage.RetryAfter = usage.WindowStart.Add(t.window)
		}
	}

	return t.save()
}

// RecordError records a rate limit error for a backend.
func (t *Tracker) RecordError(backend string, retryAfter time.Duration) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	
	usage, ok := t.usage[backend]
	if !ok {
		usage = &Usage{
			Backend:     backend,
			WindowStart: now,
		}
		t.usage[backend] = usage
	}

	usage.IsExhausted = true
	if retryAfter > 0 {
		usage.RetryAfter = now.Add(retryAfter)
	} else {
		usage.RetryAfter = now.Add(time.Hour) // Default 1 hour
	}

	return t.save()
}

// GetUsage returns the usage for a backend.
func (t *Tracker) GetUsage(backend string) (*Usage, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	usage, ok := t.usage[backend]
	if !ok {
		return nil, false
	}

	// Return a copy to prevent external modification
	copy := *usage
	return &copy, true
}

// IsExhausted returns true if the backend has exhausted its quota.
func (t *Tracker) IsExhausted(backend string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	usage, ok := t.usage[backend]
	if !ok {
		return false
	}

	// Check if exhausted and retry time has passed
	if usage.IsExhausted && time.Now().After(usage.RetryAfter) {
		// Reset exhausted state
		t.mu.RUnlock()
		t.mu.Lock()
		usage.IsExhausted = false
		usage.Requests = 0
		usage.Tokens = 0
		usage.WindowStart = time.Now()
		t.save()
		t.mu.Unlock()
		t.mu.RLock()
		return false
	}

	return usage.IsExhausted
}

// ListUsage returns usage for all backends.
func (t *Tracker) ListUsage() map[string]*Usage {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make(map[string]*Usage)
	for k, v := range t.usage {
		copy := *v
		result[k] = &copy
	}
	return result
}

// Reset clears usage for a backend.
func (t *Tracker) Reset(backend string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, ok := t.usage[backend]; ok {
		delete(t.usage, backend)
		return t.save()
	}
	return nil
}

// ResetAll clears all usage data.
func (t *Tracker) ResetAll() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.usage = make(map[string]*Usage)
	return t.save()
}

// Load loads usage data from disk.
func (t *Tracker) Load() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	data, err := os.ReadFile(t.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No file yet, start fresh
		}
		return fmt.Errorf("failed to read quota file: %w", err)
	}

	var usage map[string]*Usage
	if err := json.Unmarshal(data, &usage); err != nil {
		return fmt.Errorf("failed to parse quota file: %w", err)
	}

	t.usage = usage
	return nil
}

// save persists usage data to disk (must be called with lock held).
func (t *Tracker) save() error {
	// Create directory if needed
	dir := filepath.Dir(t.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(t.usage, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize usage: %w", err)
	}

	if err := os.WriteFile(t.path, data, 0644); err != nil {
		return fmt.Errorf("failed to write quota file: %w", err)
	}

	return nil
}
