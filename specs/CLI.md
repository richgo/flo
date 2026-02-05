# CLI Component Specification

## Overview

The `eas` CLI is the main entry point for engineers. It provides commands for initializing features, managing tasks, and running agents.

## Commands

### eas init
Initialize a new feature workspace.

```bash
eas init <feature-name> [--backend claude|copilot]
```

Creates:
- `.eas/config.yaml`
- `.eas/SPEC.md` (template)
- `.eas/tasks/manifest.json`

### eas task
Task management subcommands.

```bash
eas task list [--status pending|in_progress|complete] [--repo <name>]
eas task create <title> [--repo <name>] [--deps <id,...>] [--priority <n>]
eas task get <id>
eas task claim <id>
eas task complete <id>
```

### eas status
Show feature status overview.

```bash
eas status
```

Displays:
- Feature name and backend
- Task counts by status
- Ready tasks (can be started)
- Recent activity

### eas work
Start agent work on a task.

```bash
eas work <task-id> [--backend claude|copilot]
```

### eas mcp
MCP server for Claude integration.

```bash
eas mcp serve [--port <n>]
```

## Acceptance Criteria

### eas init
- [ ] Creates .eas directory structure
- [ ] Writes default config.yaml
- [ ] Creates empty SPEC.md template
- [ ] Initializes empty task manifest
- [ ] Returns error if already initialized

### eas task list
- [ ] Lists all tasks by default
- [ ] Filters by status
- [ ] Filters by repo
- [ ] Shows task ID, title, status, deps

### eas task create
- [ ] Creates task with unique ID
- [ ] Validates deps exist
- [ ] Saves to manifest
- [ ] Returns new task ID

### eas status
- [ ] Shows feature name
- [ ] Shows task counts
- [ ] Shows ready tasks
- [ ] Works without tasks
