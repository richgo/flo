# Enterprise AI SDLC (EAS)

AI-powered spec-driven, test-driven development for enterprise teams.

[![Tests](https://img.shields.io/badge/tests-78%20passing-brightgreen)]()
[![Go](https://img.shields.io/badge/go-1.24-blue)]()

## Overview

EAS orchestrates AI agents for structured development workflows:
- **Spec-Driven**: Start with SPEC.md, break into tasks
- **Test-Driven**: Agents must pass tests before completing tasks
- **Multi-Backend**: Claude Code or GitHub Copilot SDK
- **Git-Native**: All state stored in `.eas/` directory

## Installation

```bash
go install github.com/richgo/enterprise-ai-sdlc/cmd/eas@latest
```

## Quick Start

```bash
# Initialize a feature workspace
eas init user-auth --backend claude

# Edit the specification
vim .eas/SPEC.md

# Create tasks
eas task create "Implement OAuth" --repo android
eas task create "Add token storage" --repo android --deps t-001
eas task create "iOS OAuth" --repo ios

# Check status
eas status

# Start agent work on a task
eas work t-001
```

## Commands

| Command | Description |
|---------|-------------|
| `eas init <feature>` | Initialize workspace |
| `eas task list` | List all tasks |
| `eas task create <title>` | Create a task |
| `eas task get <id>` | Get task details |
| `eas status` | Show workspace status |
| `eas work <task-id>` | Run agent on task |
| `eas mcp serve` | Start MCP server |

## Architecture

```
.eas/
├── config.yaml       # Feature configuration
├── SPEC.md           # Feature specification
├── tasks/
│   └── manifest.json # Task DAG
└── mcp.json          # Auto-generated MCP config
```

### Backends

- **Claude**: Uses Claude Code CLI via MCP
- **Copilot**: Uses GitHub Copilot SDK (Go)

Both backends share the same tool definitions and TDD enforcement.

## Tools (MCP)

EAS exposes these tools to agents:

| Tool | Description |
|------|-------------|
| `eas_task_list` | List tasks with filters |
| `eas_task_get` | Get task details |
| `eas_task_claim` | Claim a task |
| `eas_task_complete` | Complete task (runs tests) |
| `eas_run_tests` | Run tests for task |
| `eas_spec_read` | Read SPEC.md |

## Development

```bash
# Run tests
go test ./...

# Build
go build -o eas ./cmd/eas

# Test locally
./eas init test-feature
./eas task create "Test task"
./eas status
```

## Documentation

- [Architecture](ARCHITECTURE.md)
- [CLI Approaches](CLI-APPROACHES.md)
- [Dual Backend Design](DUAL-BACKEND.md)
- [Copilot SDK Deep Dive](COPILOT-SDK.md)
- [Research Notes](RESEARCH.md)

## License

MIT
