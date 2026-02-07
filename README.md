# Flo AI SDLC

AI-powered spec-driven, test-driven development for individuals, teams & parallel agents.

[![CI](https://github.com/richgo/flo/actions/workflows/ci.yml/badge.svg)](https://github.com/richgo/flo/actions/workflows/ci.yml)
[![Tests](https://img.shields.io/badge/tests-140%20passing-brightgreen)]()
[![Go](https://img.shields.io/badge/go-1.24-blue)]()

## Overview

Flo orchestrates AI agents for structured development workflows:
- **Spec-Driven**: Start with SPEC.md, break into tasks
- **Test-Driven**: Agents must pass tests before completing tasks
- **Multi-Backend**: Claude, Copilot, Codex, Gemini - bring your own AI
- **Git-Native**: All state stored in `.flo/` directory
- **Task Types**: Different backends for different work (feature, test, docs, refactor)

## Installation

```bash
go install github.com/richgo/flo/cmd/eas@latest
```

## Quick Start

```bash
# Initialize a feature workspace
flo init user-auth --backend claude

# Edit the specification
vim .flo/SPEC.md

# Create tasks
flo task create "Implement OAuth" --repo android
flo task create "Add token storage" --repo android --deps t-001
flo task create "iOS OAuth" --repo ios

# Check status
flo status

# Start agent work on a task
flo work t-001

# Work on specific task types
flo task create "Build auth API" --type feature    # Heavy lifting
flo task create "Add unit tests" --type test       # Test generation
flo task create "Update README" --type docs        # Documentation
flo task create "Extract helpers" --type refactor  # Code cleanup
```

## Commands

| Command | Description |
|---------|-------------|
| `flo init <feature>` | Initialize workspace |
| `flo task list` | List all tasks |
| `flo task create <title>` | Create a task |
| `flo task get <id>` | Get task details |
| `flo status` | Show workspace status |
| `flo work <task-id>` | Run agent on task |
| `flo spec validate [path]` | Validate SPEC.md format |
| `flo config show` | Show configuration and secrets (masked) |
| `flo quota` | Show backend usage and quota status |
| `flo mcp serve` | Start MCP server |

## Architecture

```
.flo/
├── config.yaml       # Feature configuration
├── SPEC.md           # Feature specification
├── tasks/
│   └── manifest.json # Task DAG
└── mcp.json          # Auto-generated MCP config
```

### Multi-Provider Support (BYO AI)

Flo supports multiple AI backends with automatic provider switching:

**Available Backends:**
- **Claude**: Claude Code CLI with stream-json output
- **Copilot**: GitHub Copilot SDK (Go)
- **Codex**: OpenAI Codex CLI
- **Gemini**: Google Gemini CLI

**Task Types & Backend Selection:**

Different task types can use different backends based on their strengths:

```yaml
# Example .flo/config.yaml
backend: claude  # Default backend

task_types:
  feature:
    backend: claude      # Complex features → Claude
  test:
    backend: copilot     # Test generation → Copilot
  docs:
    backend: gemini      # Documentation → Gemini
  refactor:
    backend: codex       # Code refactoring → Codex
```

**Backend Configuration:**

```bash
# Initialize with specific backend
flo init my-feature --backend claude

# Create task with specific type (uses configured backend)
flo task create "Add auth" --type feature  # Uses claude
flo task create "Add tests" --type test    # Uses copilot
flo task create "Update docs" --type docs  # Uses gemini

# Check backend usage and quotas
flo quota
```

All backends share the same MCP tool definitions and TDD enforcement.

## Tools (MCP)

EAS exposes these tools to agents:

| Tool | Description |
|------|-------------|
| `flo_task_list` | List tasks with filters |
| `flo_task_get` | Get task details |
| `flo_task_claim` | Claim a task |
| `flo_task_complete` | Complete task (runs tests) |
| `flo_run_tests` | Run tests for task |
| `flo_spec_read` | Read SPEC.md |

## Development

### Environment Variables

Flo supports the following environment variables for configuration:

| Variable | Description | Required |
|----------|-------------|----------|
| `CLAUDE_API_KEY` | API key for Claude backend | Yes (if using Claude) |
| `COPILOT_TOKEN` | GitHub Copilot token | Yes (if using Copilot) |
| `OPENAI_API_KEY` | API key for Codex backend | Yes (if using Codex) |
| `GEMINI_API_KEY` | API key for Gemini backend | Yes (if using Gemini) |
| `FLO_BACKEND` | Default backend (claude/copilot/codex/gemini) | No (defaults to claude) |
| `FLO_MODEL` | Default model to use | No |

You can set these variables in:
- System environment variables
- `.env` file in project root
- `.flo/.env` file in the workspace

Example `.env` file:
```bash
# Choose your backend(s)
CLAUDE_API_KEY=sk-ant-api-xxxxx
COPILOT_TOKEN=ghp_xxxxx
OPENAI_API_KEY=sk-xxxxx
GEMINI_API_KEY=xxxxx

# Set default backend
FLO_BACKEND=claude
```

View current configuration with: `flo config show`

### Building from Source

```bash
# Using Make (recommended)
make build      # Build binary to bin/flo
make test       # Run tests
make lint       # Run linter
make all        # Run lint, test, and build
make install    # Install to GOPATH/bin
make clean      # Remove build artifacts

# Or using Go directly
go build -o bin/flo ./cmd/flo
go test ./...
```

### Testing Locally

```bash
# Build and test
make all

# Test the binary
./bin/flo init test-feature
./bin/flo task create "Test task"
./bin/flo status
```

## Documentation

- [Architecture](ARCHITECTURE.md)
- [CLI Approaches](CLI-APPROACHES.md)
- [Dual Backend Design](DUAL-BACKEND.md)
- [MCP & Multi-Backend Orchestration](docs/MCP.md)
- [Copilot SDK Deep Dive](COPILOT-SDK.md)
- [Research Notes](RESEARCH.md)

## License

MIT
