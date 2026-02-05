# Enterprise AI SDLC

A Go CLI/SDK for AI-assisted software development across multiple repositories with enforced TDD and parallel agent execution.

## Vision

Build high-quality software faster using AI coding assistants in a structured, enterprise-ready workflow:

- **Spec-driven** — Define what to build before writing code
- **TDD-enforced** — Tests define done, no exceptions
- **Multi-repo** — Coordinate features across Android, iOS, Web, Services
- **Parallel execution** — 1-30 AI agents working simultaneously
- **Git-native** — Worktrees, branches, no external databases

## Status

🚧 **Research & Design Phase**

## Documentation

| Document | Description |
|----------|-------------|
| [RESEARCH.md](RESEARCH.md) | Analysis of existing tools (Ralph, Gastown, Beads, Backlog.md) |
| [ARCHITECTURE.md](ARCHITECTURE.md) | Proposed Go CLI/SDK architecture |
| [INTEGRATIONS.md](INTEGRATIONS.md) | Integration research (Conductor, Spec-Kit, OpenSpec) |
| [COPILOT-SDK.md](COPILOT-SDK.md) | GitHub Copilot SDK integration notes |

## Core Concepts

### Spec-Driven Development

```
PROPOSAL.md → SPEC.md → DESIGN.md → tasks/*.md → Implementation
```

Each feature gets:
- **PROPOSAL.md** — Why we're building this
- **SPEC.md** — What we're building (user stories, acceptance criteria)
- **DESIGN.md** — How we're building it (technical approach)
- **tasks/*.md** — Atomic implementation units with TDD scenarios

### Cross-Repo Coordination

```
feature/user-auth/
├── repos.json           # Links to Android, iOS, Web, API repos
└── tasks/
    ├── auth-001.md      # API Gateway: OAuth endpoint
    ├── auth-002.md      # Web: Login button
    ├── auth-003.md      # iOS: Sign-in flow
    └── auth-004.md      # Android: Sign-in flow
```

### Parallel Execution with Isolation

```
┌─────────────────────────────────────┐
│           Orchestrator              │
│                                     │
│  ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐    │
│  │A1 │ │A2 │ │A3 │ │...│ │A30│    │
│  └─┬─┘ └─┬─┘ └─┬─┘ └───┘ └─┬─┘    │
│    │     │     │           │       │
│  ┌─┴─┐ ┌─┴─┐ ┌─┴─┐       ┌─┴─┐    │
│  │WT1│ │WT2│ │WT3│  ...  │WT30│   │  ← Git worktrees
│  └───┘ └───┘ └───┘       └───┘    │
└─────────────────────────────────────┘
```

Each agent works in an isolated git worktree. No conflicts. Clean context.

## Planned CLI

```bash
# Initialize
eas init

# Features
eas feature new user-auth
eas feature clarify user-auth
eas feature analyze user-auth

# Tasks
eas task create "Implement OAuth" --repo api-gateway --deps auth-001
eas task list --ready
eas task graph

# Execute
eas run --parallel 10
eas status
```

## Influences

This project synthesizes ideas from:

- [Ralph Loop](https://github.com/snarktank/ralph) — Autonomous agent loop with PRD-driven tasks
- [Gastown](https://github.com/steveyegge/gastown) — Multi-agent orchestration
- [Beads](https://github.com/steveyegge/beads) — Git-backed task tracking with DAG
- [Backlog.md](https://github.com/MrLesk/Backlog.md) — Markdown-native task management
- [GitHub Spec-Kit](https://github.com/github/spec-kit) — Spec-driven development toolkit
- [OpenSpec](https://github.com/Fission-AI/OpenSpec) — Lightweight spec framework
- [Conductor](https://github.com/gemini-cli-extensions/conductor) — Context-driven development

## License

MIT
