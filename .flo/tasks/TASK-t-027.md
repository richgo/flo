---
id: t-027
status: pending
model: claude/sonnet
type: build
priority: 1
---

# Loop mode CLI option - iterate through all outstanding tasks automatically

## TDD Requirements

**This task MUST follow Test-Driven Development:**

1. **Write tests first** - Before implementing any feature, write failing tests
2. **Red → Green → Refactor** - Follow the TDD cycle strictly
3. **Commit on green** - After each test passes, commit immediately
4. **Run tests continuously** - Use `flo test` or `make test` after each change
5. **No implementation without tests** - Every new function/method needs test coverage
6. **Tests must pass before completion** - Task cannot be marked complete with failing tests

### Workflow
```
1. Write failing test     → git add -A
2. Write minimal code     → tests pass? → git commit -m "feat: ..."
3. Refactor if needed     → tests pass? → git commit -m "refactor: ..."
4. Repeat
```

### Completion Checklist
- [ ] Tests written for new functionality
- [ ] All tests passing
- [ ] Atomic commits for each green state
- [ ] Coverage maintained or improved
- [ ] No regressions introduced
