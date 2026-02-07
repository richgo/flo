# MCP Server Documentation

## Overview

Flo exposes tools to AI agents via the Model Context Protocol (MCP). This enables Claude Code, Copilot, Codex, Gemini, and other MCP-compatible agents to interact with the Flo workspace.

**Key Features:**
- **Multi-Backend Support**: Works with any MCP-compatible AI backend
- **Provider Switching**: Automatic routing based on task type
- **Unified Tool Interface**: Same tools across all backends
- **TDD Enforcement**: All backends must pass tests before completing tasks

## Multi-Backend Orchestration

Flo orchestrates multiple AI backends through a unified MCP interface. Each backend implements the same tool protocol but can be optimized for different task types.

### Backend Architecture

```
┌─────────────────────────────────────────┐
│          Flo Orchestrator               │
│                                         │
│  ┌──────────────────────────────────┐  │
│  │    Task Registry & Scheduler     │  │
│  └──────────────────────────────────┘  │
│                 │                       │
│                 ▼                       │
│  ┌──────────────────────────────────┐  │
│  │   Backend Selection by Type      │  │
│  │                                  │  │
│  │  • feature  → Claude             │  │
│  │  • test     → Copilot            │  │
│  │  • docs     → Gemini             │  │
│  │  • refactor → Codex              │  │
│  └──────────────────────────────────┘  │
│                 │                       │
└─────────────────┼───────────────────────┘
                  │
      ┌───────────┴───────────┐
      │                       │
      ▼                       ▼
┌──────────┐           ┌──────────┐
│  Claude  │           │ Copilot  │
│  Backend │   MCP     │  Backend │
│          │◄─────────►│          │
└──────────┘           └──────────┘
      ▲                       ▲
      │                       │
      └───────────┬───────────┘
                  │
      ┌───────────┴───────────┐
      │                       │
      ▼                       ▼
┌──────────┐           ┌──────────┐
│  Codex   │           │  Gemini  │
│  Backend │           │  Backend │
└──────────┘           └──────────┘
```

### Backend Implementation

All backends implement the `agent.Backend` interface:

```go
type Backend interface {
    Name() string
    Start(ctx context.Context) error
    Stop() error
    CreateSession(ctx context.Context, task *task.Task, worktree string) (Session, error)
}
```

**Available Backends:**
- **Claude**: CLI-based via `claude` command with `--output-format stream-json`
- **Copilot**: SDK-based via GitHub Copilot Go SDK
- **Codex**: CLI-based via `codex` command with stream-json output
- **Gemini**: CLI-based via `gemini` command with stream-json output

### Provider Switching

Flo automatically switches providers based on task configuration:

```yaml
# .flo/config.yaml
backend: claude  # Default backend

task_types:
  feature:
    backend: claude      # Complex features
    model: claude-3-opus
  test:
    backend: copilot     # Test generation
  docs:
    backend: gemini      # Documentation
    model: gemini-pro
  refactor:
    backend: codex       # Code refactoring
```

When you create a task with `--type`, Flo:
1. Looks up the task type in config
2. Selects the configured backend
3. Creates a session with that backend
4. Routes all MCP tool calls to that backend

### Quota Management

Flo tracks usage per backend to prevent quota exhaustion:

```bash
# View current usage
flo quota

# Example output:
BACKEND   REQUESTS  TOKENS   STATUS       LAST REQUEST   WINDOW
-------   --------  ------   ------       ------------   ------
claude    45        124500   ✓ OK         5 mins ago     2.5h
copilot   12        8900     ✓ OK         1 hour ago     3.1h
gemini    3         2100     ✓ OK         2 hours ago    2.8h
```

When a backend reaches its quota:
- Flo marks it as exhausted
- Switches to fallback backend if configured
- Resumes after the retry window

## Starting the MCP Server

```bash
flo mcp serve
```

The server runs on stdio, suitable for integration with AI agent frameworks.

## Available Tools

All tools are backend-agnostic - they work the same regardless of which AI backend is processing the request.

| Tool | Description | Parameters |
|------|-------------|------------|
| `flo_task_get` | Get task details by ID | `id: string` |
| `flo_task_list` | List all tasks with optional filters | `status?: string, repo?: string` |
| `flo_task_claim` | Claim a task for work | `id: string` |
| `flo_run_tests` | Run tests for the current workspace | `repo?: string` |
| `flo_task_complete` | Mark a task as complete (runs tests) | `id: string` |
| `flo_spec_read` | Read the feature specification | none |

## Agent Discovery

Agents discover Flo via the MCP configuration file (`.flo/mcp.json`), which is auto-generated during workspace initialization:

```json
{
  "mcpServers": {
    "flo": {
      "command": "flo",
      "args": ["mcp", "serve"],
      "cwd": "/path/to/workspace",
      "env": {
        "FLO_BACKEND": "claude"
      }
    }
  }
}
```

**Backend-Specific Configuration:**

For Claude:
```json
{
  "mcpServers": {
    "flo": {
      "command": "flo",
      "args": ["mcp", "serve"],
      "env": {
        "FLO_BACKEND": "claude",
        "CLAUDE_API_KEY": "${CLAUDE_API_KEY}"
      }
    }
  }
}
```

For Copilot SDK:
```json
{
  "mcpServers": {
    "flo": {
      "command": "flo",
      "args": ["mcp", "serve"],
      "env": {
        "FLO_BACKEND": "copilot",
        "COPILOT_TOKEN": "${COPILOT_TOKEN}"
      }
    }
  }
}
```

## Concurrent Sessions

The MCP server is stateless - each agent session gets isolated access. Workspace state is protected by file locking (see pkg/task/registry.go).

## Protocol

Flo implements MCP 1.0. See https://modelcontextprotocol.io for the full specification.

## Backend Registration

To add a new backend, implement the `agent.Backend` interface and register it:

```go
// Register your backend
agent.RegisterBackend("mybackend", func(config any) agent.Backend {
    if cfg, ok := config.(*MyBackendConfig); ok {
        return NewMyBackend(*cfg)
    }
    return NewMyBackend(MyBackendConfig{})
})
```

Then configure it in `.flo/config.yaml`:

```yaml
backend: mybackend

backends:
  mybackend:
    cli_path: /path/to/mycli
    model: my-model-v1
```

## Testing

All backends share the same test suite to ensure consistent behavior:

```bash
# Test a specific backend
FLO_BACKEND=claude make test
FLO_BACKEND=copilot make test
FLO_BACKEND=codex make test
FLO_BACKEND=gemini make test
```
