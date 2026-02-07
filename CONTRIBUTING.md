# Contributing to Flo

Thank you for your interest in contributing to Flo! This document provides guidelines for contributing to the project.

## Table of Contents

- [Development Setup](#development-setup)
- [Code Style](#code-style)
- [Pull Request Process](#pull-request-process)
- [Testing Requirements](#testing-requirements)
- [Reporting Bugs](#reporting-bugs)
- [Suggesting Features](#suggesting-features)

## Development Setup

### Prerequisites

- Go 1.24+ installed
- Git for version control
- golangci-lint for linting (optional but recommended)

### Setting Up Your Environment

1. **Fork and Clone**
   ```bash
   git clone https://github.com/YOUR_USERNAME/flo.git
   cd flo
   ```

2. **Build the Project**
   ```bash
   make build
   ```

3. **Run Tests**
   ```bash
   make test
   ```

4. **Run Linter**
   ```bash
   make lint
   ```

### Project Structure

```
flo/
├── cmd/flo/          # Main application entry point
├── pkg/              # Core packages
│   ├── config/       # Configuration management
│   ├── task/         # Task data model and registry
│   ├── workspace/    # Workspace management
│   └── ...
├── specs/            # Feature specifications
└── .flo/             # EAS workspace (created at runtime)
```

## Code Style

This project follows standard Go conventions and uses golangci-lint for enforcement.

### Style Guidelines

- **Formatting**: Use `gofmt` and `goimports` (enforced by `.golangci.yml`)
- **Naming**: Follow Go naming conventions
  - Use camelCase for unexported identifiers
  - Use PascalCase for exported identifiers
  - Use descriptive names; avoid single-letter variables except in short scopes
- **Documentation**: 
  - All exported functions, types, and constants must have doc comments
  - Doc comments should be complete sentences starting with the item name
- **Error Handling**: 
  - Always check and handle errors
  - Wrap errors with context using `fmt.Errorf("context: %w", err)`
- **Testing**:
  - Write table-driven tests where appropriate
  - Use descriptive test names: `TestFunctionName_Scenario`
  - Aim for high test coverage of critical paths

### Linter Configuration

The project uses `.golangci.yml` with the following key linters enabled:

- `errcheck` - Ensures errors are checked
- `gosimple` - Suggests code simplifications
- `govet` - Reports suspicious constructs
- `staticcheck` - Advanced static analysis
- `gofmt` / `goimports` - Code formatting
- `gosec` - Security checks
- `gocritic` - Comprehensive checks

Run the linter before submitting:
```bash
make lint
```

## Pull Request Process

1. **Create a Feature Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make Your Changes**
   - Write clean, tested code
   - Follow the code style guidelines
   - Add or update tests as needed
   - Update documentation if applicable

3. **Run Tests and Linter**
   ```bash
   make all  # Runs lint, test, and build
   ```

4. **Commit Your Changes**
   - Write clear, descriptive commit messages
   - Use conventional commit format when possible:
     ```
     feat: add new feature
     fix: resolve bug in X
     docs: update README
     test: add tests for Y
     refactor: improve Z implementation
     ```

5. **Push and Create PR**
   ```bash
   git push origin feature/your-feature-name
   ```
   - Open a pull request on GitHub
   - Fill out the PR template completely
   - Link any related issues

6. **Code Review**
   - Address reviewer feedback promptly
   - Keep discussions constructive and professional
   - Be open to suggestions and improvements

7. **Merge**
   - PRs require at least one approval
   - All CI checks must pass
   - Maintainers will merge once approved

## Testing Requirements

### Test Coverage

- All new features must include tests
- Bug fixes should include regression tests
- Aim for at least 80% coverage for new code

### Writing Tests

```go
func TestFeatureName(t *testing.T) {
    // Use t.TempDir() for file operations
    tmpDir := t.TempDir()
    
    // Arrange
    input := setupTestData()
    
    // Act
    result, err := functionUnderTest(input)
    
    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result != expected {
        t.Errorf("expected %v, got %v", expected, result)
    }
}
```

### Running Tests

```bash
# Run all tests
make test

# Run specific package tests
go test ./pkg/task/...

# Run with coverage
go test -cover ./...

# Run with verbose output
go test -v ./...
```

## Reporting Bugs

Please use the [Bug Report template](.github/ISSUE_TEMPLATE/bug_report.md) when reporting bugs.

Include:
- Clear description of the issue
- Steps to reproduce
- Expected vs actual behavior
- Environment details (OS, Go version)
- Relevant logs or error messages

## Suggesting Features

Please use the [Feature Request template](.github/ISSUE_TEMPLATE/feature_request.md) when suggesting features.

Include:
- Problem statement or use case
- Proposed solution
- Alternative solutions considered
- Impact on existing functionality

## Code of Conduct

This project adheres to the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## Questions?

- Open a [GitHub Discussion](https://github.com/richgo/flo/discussions)
- Check existing [Issues](https://github.com/richgo/flo/issues)
- Review the [README](README.md) and documentation

Thank you for contributing to Flo! 🚀
