# Flo v3: Design Improvements

**Status**: PROPOSED  
**Date**: 2026-02-11  
**Author**: Jambo  
**Context**: Review of modern AI development tools and Flo's current capabilities

## Executive Summary

Flo v2 established multi-backend support and quota management. v3 focuses on **developer experience**, **team collaboration**, and **production readiness**. Key themes:

1. **Visual Interfaces** - TUI and web dashboard alongside CLI
2. **Team Features** - Collaboration, review workflows, knowledge sharing
3. **Observability** - Rich logging, tracing, debugging tools
4. **Performance** - Caching, incremental builds, cost optimization
5. **Integration Ecosystem** - Git platforms, IDEs, CI/CD, issue trackers

## 1. Interactive TUI (Terminal UI)

**Problem**: CLI is powerful but lacks real-time visibility into ongoing work.

**Solution**: Add a `flo ui` command that launches a Bubble Tea TUI.

### Features

```go
// Main views
type View int
const (
    ViewTaskBoard    View = iota  // Kanban board of tasks
    ViewTaskDetail                 // Detailed task view with logs
    ViewQuotaUsage                 // Real-time quota dashboard
    ViewAgentActivity              // Live agent work stream
    ViewConfig                     // Interactive config editor
)
```

**Task Board View:**
```
┌─ Flo: user-auth ─────────────────────────────────────────────────────┐
│ ⚡ 3 running  ✓ 12 complete  ○ 8 pending  ⚠ 1 blocked               │
├──────────────────────────────────────────────────────────────────────┤
│ TODO (8)          IN PROGRESS (3)    COMPLETE (12)     BLOCKED (1)   │
│                                                                       │
│ ○ t-028          ⚡ t-027           ✓ t-026            ⚠ t-015       │
│   Add OAuth        Loop mode          MCP docs          Needs deps   │
│   [build]          [claude/sonnet]    [docs]            [test]       │
│   Priority: 1      75% complete       2h ago            Dep: t-014   │
│                                                                       │
│ ○ t-029          ⚡ t-030           ✓ t-025                          │
│   iOS impl         Test suite        README             │
│   [build]          [test]            [docs]                          │
│   Priority: 2      45% complete      3h ago                          │
│                                                                       │
│ Press 't' to create task  'w' to work  'q' to quit  '?' for help     │
└──────────────────────────────────────────────────────────────────────┘
```

**Agent Activity View:**
```
┌─ Agent Activity (Live) ──────────────────────────────────────────────┐
│ t-027: Loop mode implementation [claude/sonnet]                      │
│ ├─ 14:30:15 │ Reading task.md                                        │
│ ├─ 14:30:18 │ Tool: flo_spec_read → .flo/SPEC.md                     │
│ ├─ 14:30:22 │ Analyzing codebase: cmd/flo/cmd/loop.go                │
│ ├─ 14:30:35 │ Writing code: pkg/workflow/loop.go                     │
│ ├─ 14:30:48 │ Tool: flo_run_tests → 5/5 passing ✓                   │
│ └─ 14:30:51 │ Committing changes...                                  │
│                                                                       │
│ t-030: Test suite expansion [copilot/gpt-4]                          │
│ ├─ 14:29:45 │ Tool: flo_task_get → t-030                             │
│ ├─ 14:30:02 │ Generating test cases...                               │
│ └─ 14:30:30 │ Writing: pkg/workflow/loop_test.go                     │
│                                                                       │
│ Quota: Claude 42/50 req/hr  Copilot 8/∞  Tokens: 125k               │
└──────────────────────────────────────────────────────────────────────┘
```

**Implementation:**
- Use [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework
- WebSocket connection to flo daemon for live updates
- Keyboard shortcuts for all actions (vim-style)
- Mouse support optional

### CLI Integration

```bash
# Launch TUI
flo ui

# Launch specific view
flo ui --view quota
flo ui --view tasks

# Daemon mode (background process)
flo daemon start
flo daemon stop
flo daemon status
```

**Priority**: High  
**Effort**: 2-3 weeks  
**Dependencies**: None

---

## 2. Web Dashboard

**Problem**: TUI is great for terminal users, but teams need shareable visualization and PM/stakeholder visibility.

**Solution**: Web-based dashboard with real-time updates.

### Architecture

```
┌──────────────┐         ┌──────────────┐         ┌──────────────┐
│   Browser    │◄────────┤  flo serve   │◄────────┤  .flo/       │
│  (React/Vue) │  HTTP   │  (Go server) │  FS     │  workspace   │
│              │  WS     │              │         │              │
└──────────────┘         └──────────────┘         └──────────────┘
```

### Features

**Dashboard Views:**
1. **Overview** - Metrics, progress, quota usage
2. **Task Board** - Drag-and-drop Kanban (like Linear, GitHub Projects)
3. **Agent Activity** - Live stream of agent actions
4. **Code Review** - Diff viewer for AI-generated changes
5. **Quota Analytics** - Cost tracking, usage trends
6. **Team Activity** - Who's working on what

**Example UI (Overview):**
```
┌─────────────────────────────────────────────────────────────────┐
│ Flo Dashboard: user-auth                                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Progress: ████████████░░░░░ 75% (18/24 tasks complete)        │
│                                                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │ In Progress │  │  Complete   │  │   Blocked   │             │
│  │     3       │  │     18      │  │     1       │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
│                                                                  │
│  Recent Activity:                                                │
│  ✓ t-027: Loop mode - completed by claude/sonnet (2m ago)      │
│  ⚡ t-030: Test suite - in progress by copilot/gpt-4            │
│  ✓ t-026: MCP docs - completed by claude/haiku (3h ago)        │
│                                                                  │
│  Quota Usage (last 24h):                                         │
│  • Claude:  42/50 req/hr  │  125k tokens  │  $2.45             │
│  • Copilot:  8/∞ req      │   45k tokens  │  $0.00             │
│  • Total Cost: $2.45                                             │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Implementation

**Backend:**
```go
// cmd/flo/cmd/serve.go
type Server struct {
    workspace *workspace.Workspace
    hub       *websocket.Hub  // Broadcast updates
    router    *chi.Router
}

func (s *Server) Start(port int) error {
    // REST API for task CRUD
    s.router.Get("/api/tasks", s.handleGetTasks)
    s.router.Post("/api/tasks", s.handleCreateTask)
    
    // WebSocket for live updates
    s.router.Get("/ws", s.handleWebSocket)
    
    // Static assets (embedded React build)
    s.router.Get("/*", s.handleStatic)
    
    return http.ListenAndServe(fmt.Sprintf(":%d", port), s.router)
}
```

**Frontend:**
- React or Vue.js
- Real-time updates via WebSocket
- Tailwind CSS for styling
- Charts.js for analytics
- CodeMirror for diff viewing

### CLI Integration

```bash
# Start web server
flo serve --port 3000

# Open in browser
flo serve --open

# Production deployment
flo serve --host 0.0.0.0 --port 8080 --tls
```

**Priority**: Medium  
**Effort**: 4-6 weeks  
**Dependencies**: None

---

## 3. Team Collaboration Features

**Problem**: Flo is single-user focused. Teams need coordination, reviews, and knowledge sharing.

### 3.1 Task Assignment & Ownership

```yaml
# .flo/tasks/TASK-t-028.md frontmatter
---
id: t-028
title: Add OAuth implementation
assignee: alice@example.com
reviewers: [bob@example.com, charlie@example.com]
status: in_review
type: build
model: claude/sonnet
---
```

**CLI:**
```bash
# Assign task
flo task assign t-028 alice@example.com

# Request review
flo task review t-028 --reviewers bob@example.com,charlie@example.com

# Approve changes
flo review approve t-028 --comment "LGTM!"

# Request changes
flo review reject t-028 --comment "Needs error handling"
```

### 3.2 Code Review Workflow

**Problem**: AI generates code, but humans should review before merge.

**Solution**: Built-in review workflow similar to GitHub PRs.

```bash
# AI completes task → creates draft
flo work t-028  # Status: in_review (not complete)

# Human reviews diff
flo review show t-028

# Diff view (TUI or web)
┌─ Review: t-028 Add OAuth implementation ─────────────────────────────┐
│ Files changed: 3  Additions: +245  Deletions: -12                   │
├──────────────────────────────────────────────────────────────────────┤
│ pkg/auth/oauth.go                                                    │
│ + func NewOAuthHandler(config *Config) *OAuthHandler {              │
│ +     return &OAuthHandler{                                          │
│ +         config: config,                                            │
│ +         client: &http.Client{Timeout: 10 * time.Second},          │
│ +     }                                                              │
│ + }                                                                  │
│                                                                      │
│ Comments:                                                            │
│ alice: "Should we add retry logic here?"                            │
│ bob: "Looks good, but add context.Context parameter"                │
│                                                                      │
│ [a]pprove  [r]equest changes  [c]omment  [q]uit                     │
└──────────────────────────────────────────────────────────────────────┘
```

### 3.3 Knowledge Base

**Problem**: Agents don't learn from past mistakes or team patterns.

**Solution**: `.flo/knowledge/` directory with team conventions, past issues, architectural decisions.

```
.flo/
├── knowledge/
│   ├── decisions/           # ADRs (Architecture Decision Records)
│   │   ├── 001-use-postgres.md
│   │   └── 002-auth-strategy.md
│   ├── patterns/            # Code patterns, best practices
│   │   ├── error-handling.md
│   │   └── testing-patterns.md
│   ├── gotchas/             # Common mistakes to avoid
│   │   └── auth-pitfalls.md
│   └── context.md           # High-level project context
```

**Agent Integration:**
Agents automatically read `.flo/knowledge/` before starting tasks:

```go
// pkg/agent/context.go
func (a *Agent) LoadKnowledge() ([]string, error) {
    docs := []string{}
    
    // Load ADRs
    decisions, _ := os.ReadDir(".flo/knowledge/decisions")
    for _, d := range decisions {
        content, _ := os.ReadFile(filepath.Join(".flo/knowledge/decisions", d.Name()))
        docs = append(docs, string(content))
    }
    
    // Load patterns relevant to task type
    // ...
    
    return docs, nil
}
```

**CLI:**
```bash
# Add decision
flo knowledge add decision "Use PostgreSQL for auth storage"

# Add pattern
flo knowledge add pattern error-handling.md

# Search knowledge base
flo knowledge search "authentication"

# Show what agents see for a task type
flo knowledge show --type build
```

**Priority**: High (3.1), Medium (3.2, 3.3)  
**Effort**: 3-4 weeks total

---

## 4. Enhanced Observability

**Problem**: When agents fail or produce unexpected output, debugging is opaque.

### 4.1 Structured Logging

```go
// pkg/agent/logger.go
type AgentLogger struct {
    taskID   string
    backend  string
    model    string
    logger   *slog.Logger
}

func (l *AgentLogger) LogThought(thought string) {
    l.logger.Info("agent_thought",
        "task_id", l.taskID,
        "backend", l.backend,
        "model", l.model,
        "thought", thought,
    )
}

func (l *AgentLogger) LogToolCall(tool string, args map[string]interface{}, result interface{}) {
    l.logger.Info("tool_call",
        "task_id", l.taskID,
        "tool", tool,
        "args", args,
        "result", result,
        "duration_ms", durationMs,
    )
}
```

**Output formats:**
```bash
# Human-readable (default)
flo work t-028 --verbose

# JSON for parsing
flo work t-028 --log-format json > task-028.jsonl

# Send to observability platform
flo work t-028 --log-export datadog://api-key
```

### 4.2 Trace Viewer

**Problem**: Complex tasks involve many tool calls and agent decisions. Need timeline view.

**Solution**: OpenTelemetry-compatible traces.

```bash
# Export traces
flo work t-028 --trace

# View in Jaeger/Grafana
flo trace view t-028

# CLI trace viewer (TUI)
flo trace show t-028
```

**Trace View:**
```
┌─ Trace: t-028 Add OAuth (Duration: 2m 34s) ─────────────────────────┐
│                                                                       │
│ ▓▓▓▓▓░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ 2m 34s                    │
│                                                                       │
│ Timeline:                                                             │
│ 0s    ├─ flo_task_get (50ms)                                         │
│ 0.1s  ├─ flo_spec_read (120ms)                                       │
│ 0.5s  ├─ claude_request (45s) ◄─ SLOWEST                            │
│       │  ├─ Analyzing requirements (5s)                              │
│       │  ├─ Planning implementation (10s)                            │
│       │  └─ Generating code (30s)                                    │
│ 45s   ├─ Write pkg/auth/oauth.go (200ms)                             │
│ 46s   ├─ flo_run_tests (15s)                                         │
│       │  ├─ TestOAuthHandler (2s) ✓                                  │
│       │  └─ TestTokenRefresh (13s) ✓                                 │
│ 61s   └─ git commit (1s)                                             │
│                                                                       │
│ Spans: 12  Errors: 0  Warnings: 2                                    │
└───────────────────────────────────────────────────────────────────────┘
```

### 4.3 Agent Replay & Debugging

**Problem**: Agents sometimes fail midway. Hard to debug or resume.

**Solution**: Checkpoint system with replay capability.

```bash
# Resume failed task
flo work t-028 --resume

# Replay with different model
flo replay t-028 --model copilot/gpt-4

# Step-by-step debugging
flo debug t-028
> Next tool: flo_run_tests
> [s]tep  [c]ontinue  [i]nspect  [q]uit
```

**Priority**: High  
**Effort**: 3-4 weeks

---

## 5. Performance & Cost Optimization

### 5.1 Intelligent Caching

**Problem**: Agents repeatedly analyze same code, run same tests, ask same questions.

**Solution**: Multi-level cache system.

```
┌─────────────────────────────────────────────────────────┐
│ Cache Levels                                             │
├─────────────────────────────────────────────────────────┤
│ L1: In-Memory    │ Recent responses     │ TTL: 5m       │
│ L2: Disk (.flo/) │ Embeddings, analyses │ TTL: 24h      │
│ L3: Object Store │ Shared team cache    │ TTL: 7d       │
└─────────────────────────────────────────────────────────┘
```

**Implementation:**
```go
// pkg/cache/cache.go
type Cache interface {
    Get(key string) ([]byte, error)
    Set(key string, value []byte, ttl time.Duration) error
    Invalidate(pattern string) error
}

// Semantic cache for LLM responses
type SemanticCache struct {
    embeddings *vector.Store  // Embeddings DB
    threshold  float64        // Similarity threshold
}

func (c *SemanticCache) GetSimilar(query string) (string, bool) {
    embedding := c.embed(query)
    results := c.embeddings.Search(embedding, c.threshold)
    if len(results) > 0 {
        return results[0].Response, true
    }
    return "", false
}
```

**Cache Strategies:**
1. **Exact Match** - Identical queries (tool calls, code analysis)
2. **Semantic Match** - Similar questions using embeddings
3. **Content-Based** - File hashes for unchanged files
4. **Test Results** - Cache passing test runs

**CLI:**
```bash
# Show cache stats
flo cache stats

# Clear cache
flo cache clear --type embeddings

# Configure cache
flo config set cache.enabled true
flo config set cache.ttl 24h
```

### 5.2 Incremental Task Execution

**Problem**: Re-running entire task when only one file changed.

**Solution**: Track dependencies and only re-execute affected parts.

```go
// pkg/task/dependencies.go
type DependencyGraph struct {
    nodes map[string]*Node
    edges map[string][]string
}

func (g *DependencyGraph) Affected(changedFiles []string) []string {
    // Return list of tasks/tests affected by file changes
}
```

### 5.3 Cost Tracking & Budgets

**Problem**: AI costs can spiral. Need visibility and controls.

**Solution**: Per-task, per-user, per-feature budgets.

```yaml
# .flo/config.yaml
budgets:
  feature: 
    total: $100.00
    remaining: $42.35
  daily:
    limit: $20.00
    alert_threshold: 0.8  # Alert at 80%
  per_task:
    max: $5.00
```

**CLI:**
```bash
# Show costs
flo cost show

# Set budget
flo budget set feature --limit 100.00

# Cost report
flo cost report --range last-week --export csv
```

**Dashboard widget:**
```
┌─ Cost Tracking ──────────────────────────────────┐
│ Feature Budget: $57.65 / $100.00  ████████░░░░  │
│ Daily Spend:    $12.34 / $20.00   ██████░░░░░░  │
│                                                   │
│ Breakdown by Backend:                             │
│ • Claude:  $45.20  (78%)  █████████████░░░      │
│ • Copilot:  $8.15  (14%)  ███░░░░░░░░░░░░░      │
│ • Gemini:   $4.30  ( 8%)  ██░░░░░░░░░░░░░░      │
│                                                   │
│ ⚠ Approaching daily limit! (62%)                │
└───────────────────────────────────────────────────┘
```

**Priority**: Medium (5.1), High (5.2), Medium (5.3)  
**Effort**: 4-5 weeks total

---

## 6. Integration Ecosystem

### 6.1 IDE Extensions

**VS Code Extension:**
- Task list in sidebar
- Inline AI suggestions
- One-click task creation from TODOs
- Agent activity notifications

**IntelliJ Plugin:**
- Similar features for JetBrains IDEs

### 6.2 Git Platform Integration

**GitHub App:**
```yaml
# .github/flo.yml
on:
  issue:
    types: [labeled]
    label: "flo:auto"
  
actions:
  - create_task:
      type: build
      title: "{{ issue.title }}"
      description: "{{ issue.body }}"
      
  - on_task_complete:
      create_pr: true
      request_review: true
```

**GitLab Integration:**
- Merge request integration
- Pipeline triggers

### 6.3 Issue Tracker Sync

**Jira:**
```bash
# Sync Jira issues as tasks
flo integrate jira sync --project AUTH

# Update Jira on task completion
flo integrate jira enable --auto-update
```

**Linear:**
```bash
flo integrate linear sync
```

### 6.4 CI/CD Integration

**GitHub Actions:**
```yaml
# .github/workflows/flo.yml
name: Flo AI Tasks
on: [push]
jobs:
  work:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: flo-ai/flo-action@v1
        with:
          command: work --loop --max-iterations 5
          api-key: ${{ secrets.FLO_API_KEY }}
```

**Jenkins Plugin:**
- Trigger Flo tasks from Jenkins pipelines
- Report results back to Jenkins

### 6.5 Notification Channels

**Slack Bot:**
```
@flo create task "Add rate limiting" --type build
```

**Discord Bot:**
- Similar functionality

**Email/Webhook:**
- Digest emails (daily/weekly)
- Webhook notifications for external systems

**Priority**: Low-Medium  
**Effort**: 6-8 weeks (phased rollout)

---

## 7. Enterprise Features

### 7.1 Multi-Workspace Management

**Problem**: Large orgs have many features/projects.

**Solution**: Workspace registry and cross-workspace operations.

```bash
# List workspaces
flo workspace list

# Switch workspace
flo workspace use auth-service

# Create workspace
flo workspace create payment-service --template microservice

# Cross-workspace operations
flo search "authentication" --workspaces all
```

### 7.2 Access Control & Audit

```yaml
# .flo/access.yaml
roles:
  developer:
    permissions: [read_tasks, create_tasks, work_tasks]
  reviewer:
    permissions: [read_tasks, review_tasks, approve_tasks]
  admin:
    permissions: [all]

users:
  alice@example.com: [developer, reviewer]
  bob@example.com: [developer]
  charlie@example.com: [admin]
```

**Audit Log:**
```bash
# View audit trail
flo audit log --user alice@example.com --action work_task

# Export for compliance
flo audit export --format json --range 2026-01
```

### 7.3 SSO / SAML Integration

```bash
# Configure SSO
flo auth configure-sso \
  --provider okta \
  --domain example.okta.com \
  --client-id abc123

# Login via SSO
flo auth login --sso
```

**Priority**: Low (7.1), Low (7.2), Low (7.3)  
**Effort**: 4-6 weeks

---

## 8. Advanced Agent Capabilities

### 8.1 Multi-Agent Orchestration

**Problem**: Some tasks require coordination between multiple agents.

**Solution**: Agent mesh with task decomposition.

```bash
# Create parent task that spawns sub-tasks
flo task create "Implement full auth system" \
  --type architecture \
  --auto-decompose

# Flo creates sub-tasks:
# - t-101: Design auth API (claude/opus)
# - t-102: Implement OAuth (claude/sonnet)
# - t-103: Add tests (copilot/gpt-4)
# - t-104: Update docs (claude/haiku)

# Agents work in parallel
flo work --parallel --max-agents 4
```

### 8.2 Human-in-the-Loop

**Problem**: Critical decisions need human approval.

**Solution**: Approval gates in task flow.

```yaml
# .flo/tasks/TASK-t-030.md
---
id: t-030
approvals:
  - stage: before_start
    required: true
    approvers: [alice@example.com]
  - stage: before_merge
    required: true
    approvers: [bob@example.com, charlie@example.com]
---
```

### 8.3 Agent Learning & Feedback

**Problem**: Agents don't improve from mistakes.

**Solution**: Feedback loop with reinforcement learning.

```bash
# Rate agent output
flo feedback t-030 --rating 4 --comment "Good but needs optimization"

# Agents use ratings to improve
# (Store in .flo/knowledge/feedback/)
```

**Priority**: Medium (8.1), High (8.2), Low (8.3)  
**Effort**: 4-5 weeks

---

## 9. Documentation & Onboarding

### 9.1 Interactive Tutorials

```bash
# Guided tutorial
flo tutorial start

# Learn by doing
flo tutorial create-your-first-task
flo tutorial work-with-agents
flo tutorial review-workflow
```

### 9.2 Example Templates

```bash
# List templates
flo template list

# Create from template
flo init my-api --template rest-api

# Community templates
flo template search "microservice"
flo template install awesome-microservice
```

### 9.3 AI-Powered Help

```bash
# Ask Flo for help
flo help "how do I assign a task to someone?"
flo help "what's the best model for refactoring?"

# Context-aware suggestions
flo suggest
# Output: "You have 3 blocked tasks. Run 'flo task unblock' to see why."
```

**Priority**: Medium  
**Effort**: 2-3 weeks

---

## Implementation Roadmap

### Phase 1: Foundation (v3.0) - 8 weeks
- **Week 1-3**: TUI (`flo ui`)
- **Week 4-6**: Enhanced observability (logging, tracing)
- **Week 7-8**: Intelligent caching

### Phase 2: Collaboration (v3.1) - 6 weeks
- **Week 1-2**: Task assignment & ownership
- **Week 3-4**: Code review workflow
- **Week 5-6**: Knowledge base

### Phase 3: Performance (v3.2) - 4 weeks
- **Week 1-2**: Incremental execution
- **Week 3-4**: Cost tracking & budgets

### Phase 4: Integration (v3.3) - 8 weeks
- **Week 1-3**: IDE extensions (VS Code, IntelliJ)
- **Week 4-5**: GitHub/GitLab integration
- **Week 6-8**: Issue tracker sync (Jira, Linear)

### Phase 5: Scale (v3.4) - 6 weeks
- **Week 1-2**: Web dashboard
- **Week 3-4**: Multi-agent orchestration
- **Week 5-6**: Multi-workspace management

### Phase 6: Polish (v3.5) - 4 weeks
- **Week 1-2**: Interactive tutorials
- **Week 3**: Template library
- **Week 4**: AI-powered help

**Total Effort**: ~36 weeks (~9 months)

---

## Success Metrics

**User Experience:**
- Time to first task completion < 5 minutes
- Task review turnaround < 2 hours
- Developer satisfaction score > 4.5/5

**Performance:**
- 50% reduction in redundant API calls (via caching)
- 30% cost savings (via optimization)
- 40% faster task completion (via incremental execution)

**Adoption:**
- 1000+ active users in 6 months
- 50+ enterprise customers in 12 months
- 10+ community contributions per month

**Business:**
- Break-even on infrastructure costs by month 9
- $1M ARR by month 12

---

## Open Questions

1. **Pricing Model**: How to monetize? (Open-source core + paid enterprise features? SaaS?)
2. **Deployment**: Self-hosted vs. cloud? Both?
3. **Security**: How to handle secrets in multi-user environments?
4. **Scalability**: How many concurrent agents per workspace?
5. **Compliance**: GDPR, SOC2, ISO compliance requirements?

---

## Appendix: Competitive Analysis

**Similar Tools:**
- **Cursor**: IDE-first, single-user, no task management
- **Replit Agent**: Web-only, no local development
- **Devin**: Closed-source, expensive, limited availability
- **Aider**: CLI-only, no team features, basic task tracking

**Flo's Differentiators:**
- Open-source with enterprise options
- Multi-backend (BYO AI)
- Team collaboration features
- TDD enforcement
- Git-native architecture
- Rich observability

---

## Next Steps

1. **Validate** with current users (survey, interviews)
2. **Prioritize** features based on feedback
3. **Prototype** TUI (quick win, high impact)
4. **Design** web dashboard mockups
5. **RFC** for critical features (caching, multi-agent)
6. **Recruit** contributors for help with integrations

---

**END OF SPEC**
