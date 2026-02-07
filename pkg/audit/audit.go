// Package audit provides structured logging and audit trail functionality.
package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Level represents the severity level of an audit event.
type Level string

const (
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// Event represents an audit trail event.
type Event struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     Level                  `json:"level"`
	Operation string                 `json:"operation"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// Logger handles audit event logging.
type Logger struct {
	mu       sync.Mutex
	filePath string
	file     *os.File
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// Init initializes the global audit logger with the given workspace root.
// It creates the audit log file at .flo/audit.log.
func Init(workspaceRoot string) error {
	var err error
	once.Do(func() {
		auditPath := filepath.Join(workspaceRoot, ".flo", "audit.log")
		
		// Ensure directory exists
		dir := filepath.Dir(auditPath)
		if mkdirErr := os.MkdirAll(dir, 0755); mkdirErr != nil {
			err = fmt.Errorf("failed to create audit directory: %w", mkdirErr)
			return
		}
		
		// Open file in append mode
		file, openErr := os.OpenFile(auditPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if openErr != nil {
			err = fmt.Errorf("failed to open audit log: %w", openErr)
			return
		}
		
		defaultLogger = &Logger{
			filePath: auditPath,
			file:     file,
		}
	})
	return err
}

// Close closes the audit logger.
func Close() error {
	if defaultLogger != nil && defaultLogger.file != nil {
		return defaultLogger.file.Close()
	}
	return nil
}

// Log writes an audit event to the log file.
func Log(level Level, operation, message string, details map[string]interface{}) {
	if defaultLogger == nil {
		// If not initialized, skip logging silently
		return
	}
	
	event := Event{
		Timestamp: time.Now(),
		Level:     level,
		Operation: operation,
		Message:   message,
		Details:   details,
	}
	
	defaultLogger.writeEvent(event)
}

// Info logs an informational audit event.
func Info(operation, message string, details map[string]interface{}) {
	Log(LevelInfo, operation, message, details)
}

// Warn logs a warning audit event.
func Warn(operation, message string, details map[string]interface{}) {
	Log(LevelWarn, operation, message, details)
}

// Error logs an error audit event.
func Error(operation, message string, details map[string]interface{}) {
	Log(LevelError, operation, message, details)
}

func (l *Logger) writeEvent(event Event) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.file == nil {
		return
	}
	
	data, err := json.Marshal(event)
	if err != nil {
		// Can't log an error about logging, so just return
		return
	}
	
	// Write event as JSON line
	l.file.Write(data)
	l.file.Write([]byte("\n"))
}
