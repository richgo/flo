# GitHub Copilot SDK - Research Summary

## Overview

**Status:** Technical Preview
**Go Package:** `github.com/github/copilot-sdk/go`

The official Copilot SDK exposes the same engine behind Copilot CLI as a programmable interface. No need to build your own orchestration—you define agent behavior, Copilot handles planning, tool invocation, file edits.

## Architecture

```
Your Application (Go)
        ↓
   SDK Client
        ↓ JSON-RPC
Copilot CLI (server mode)
        ↓
   LLM Backend
```

The SDK manages the CLI process lifecycle automatically. Can also connect to external CLI server.

## Installation

```bash
# Prerequisites
# 1. Copilot CLI installed and authenticated
copilot --version

# 2. Install Go SDK
go get github.com/github/copilot-sdk/go
```

## Basic Usage (Go)

### Simple Request
```go
package main

import (
    "context"
    "fmt"
    "log"

    copilot "github.com/github/copilot-sdk/go"
)

func main() {
    ctx := context.Background()
    client := copilot.NewClient(nil)
    if err := client.Start(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Stop()

    session, err := client.CreateSession(ctx, &copilot.SessionConfig{
        Model: "gpt-4.1",
    })
    if err != nil {
        log.Fatal(err)
    }

    response, err := session.SendAndWait(ctx, copilot.MessageOptions{
        Prompt: "What is 2 + 2?",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(*response.Data.Content)
}
```

### Streaming Responses
```go
session, _ := client.CreateSession(ctx, &copilot.SessionConfig{
    Model:     "gpt-4.1",
    Streaming: true,
})

// Listen for response chunks
session.On(func(event copilot.SessionEvent) {
    if event.Type == "assistant.message_delta" {
        fmt.Print(*event.Data.DeltaContent)
    }
    if event.Type == "session.idle" {
        fmt.Println()
    }
})

_, err = session.SendAndWait(ctx, copilot.MessageOptions{
    Prompt: "Tell me a short joke",
})
```

## Authentication Options

| Method | Description |
|--------|-------------|
| GitHub signed-in user | Uses stored OAuth credentials from `copilot` CLI login |
| OAuth GitHub App | Pass user tokens from your GitHub OAuth app |
| Environment variables | `COPILOT_GITHUB_TOKEN`, `GH_TOKEN`, `GITHUB_TOKEN` |
| BYOK | Bring Your Own Key (OpenAI, Azure, Anthropic) |

### BYOK (Bring Your Own Key)

Use the SDK without GitHub auth by providing your own API keys:
- OpenAI
- Azure AI Foundry  
- Anthropic

This could be useful for enterprise deployments with existing LLM contracts.

## Key Features

### Sessions
- Create independent conversation contexts
- Persist and resume sessions across restarts
- Multiple simultaneous sessions (perfect for parallel agents)

### Tools/Function Calling
- Define custom tools that Copilot can invoke
- SDK handles planning and tool orchestration
- Perfect for TDD loop: define "run_tests" tool

### Events
```go
// Event types
"assistant.message"       // Complete message
"assistant.message_delta" // Streaming chunk
"session.idle"           // Session finished processing
"tool.call"              // Tool invocation requested
"tool.result"            // Tool returned result
```

## Recipes Available (Go)

From the cookbook:
- **Error Handling**: Connection failures, timeouts, cleanup
- **Multiple Sessions**: Manage parallel conversations
- **Managing Local Files**: AI-powered file operations
- **PR Visualization**: GitHub MCP Server integration
- **Persisting Sessions**: Save/resume across restarts

## Integration with Our Architecture

### Agent Implementation
```go
type Agent struct {
    ID       string
    Task     *Task
    Worktree *Worktree
    Session  *copilot.Session  // Copilot session
    Client   *copilot.Client
}

func (a *Agent) Execute(ctx context.Context) (*TaskResult, error) {
    // 1. Create session with context
    session, _ := a.Client.CreateSession(ctx, &copilot.SessionConfig{
        Model:     "gpt-4.1",
        Streaming: true,
    })
    
    // 2. Load task context
    spec := a.loadSpec()
    testFile := a.loadTestFile()
    
    // 3. TDD loop via Copilot
    prompt := fmt.Sprintf(`
You are implementing a task in a TDD workflow.

## Task
%s

## Test File (must pass)
%s

## Instructions
1. Read the test file
2. Implement code to make tests pass
3. Run tests using the run_tests tool
4. Iterate until all tests pass
`, a.Task.Description, testFile)
    
    response, _ := session.SendAndWait(ctx, copilot.MessageOptions{
        Prompt: prompt,
    })
    
    return a.evaluateResult(response)
}
```

### Parallel Orchestration
```go
func (o *Orchestrator) RunParallel(ctx context.Context, tasks []*Task, maxWorkers int) {
    sem := make(chan struct{}, maxWorkers)
    var wg sync.WaitGroup
    
    for _, task := range tasks {
        if !task.IsReady() {
            continue
        }
        
        sem <- struct{}{} // Acquire
        wg.Add(1)
        
        go func(t *Task) {
            defer wg.Done()
            defer func() { <-sem }() // Release
            
            agent := o.spawnAgent(t)
            result, err := agent.Execute(ctx)
            o.handleResult(t, result, err)
        }(task)
    }
    
    wg.Wait()
}
```

## Billing

- Same model as Copilot CLI
- Each prompt counts toward premium request quota
- Free tier available with limited usage
- BYOK bypasses GitHub billing (use your own LLM costs)

## Considerations for Enterprise

### Pros
- ✅ Official SDK, production-tested
- ✅ Go support
- ✅ Session management built-in
- ✅ Tool/function calling
- ✅ BYOK option for existing LLM contracts
- ✅ Streaming support

### Cons
- ⚠️ Technical Preview (API may change)
- ⚠️ Requires Copilot CLI as backend process
- ⚠️ Billing per request (unless BYOK)

### Recommendations

1. **Use BYOK for enterprise** - leverage existing OpenAI/Azure contracts
2. **Session per worktree** - natural isolation
3. **Custom tools for TDD** - `run_tests`, `run_lint`, `commit`
4. **Streaming for progress** - real-time feedback to orchestrator

## References

- [copilot-sdk repo](https://github.com/github/copilot-sdk)
- [Go cookbook](https://github.com/github/awesome-copilot/tree/main/cookbook/copilot-sdk/go)
- [Getting Started Guide](https://github.com/github/copilot-sdk/blob/main/docs/getting-started.md)
- [BYOK Documentation](https://github.com/github/copilot-sdk/blob/main/docs/auth/byok.md)
