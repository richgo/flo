# Agent Backend Specification

## Overview

Agent backends execute tasks using AI. The abstraction allows swapping between Claude Code and Copilot SDK.

## Interface

```go
type Backend interface {
    Name() string
    Start(ctx context.Context) error
    Stop() error
    CreateSession(ctx context.Context, task *task.Task, worktree string) (Session, error)
}

type Session interface {
    Run(ctx context.Context, prompt string) (*Result, error)
    Events() <-chan Event
    Destroy(ctx context.Context) error
}

type Result struct {
    Success bool
    Output  string
    Error   string
}

type Event struct {
    Type    string  // "message", "tool_call", "complete", "error"
    Content string
}
```

## Backends

### ClaudeBackend
- Executes claude CLI with MCP config
- Uses `--print` for non-interactive
- Streams via `--output-format stream-json`

### CopilotBackend
- Uses Copilot SDK (Go)
- Native tool registration
- Pre-tool hooks for TDD

## Acceptance Criteria

### Backend Interface
- [ ] Name() returns backend identifier
- [ ] Start() initializes the backend
- [ ] Stop() cleans up resources
- [ ] CreateSession() returns a new session

### Session Interface
- [ ] Run() executes prompt and returns result
- [ ] Events() streams events during execution
- [ ] Destroy() cleans up session

### ClaudeBackend
- [ ] Builds correct CLI command
- [ ] Passes MCP config path
- [ ] Parses stream-json output
- [ ] Returns structured result

### Mock Backend (for testing)
- [ ] Configurable responses
- [ ] Records calls for verification
