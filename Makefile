.PHONY: all build test lint clean install help

# Default target
all: lint test build

# Build the project
build:
	@echo "Building flo..."
	@mkdir -p bin
	@go build -o bin/flo ./cmd/flo
	@echo "Build complete: bin/flo"

# Run tests
test:
	@echo "Running tests..."
	@go test ./... -v

# Run linter
lint:
	@echo "Running golangci-lint..."
	@golangci-lint run

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@echo "Clean complete"

# Install the binary
install:
	@echo "Installing flo..."
	@go install ./cmd/flo
	@echo "Install complete"

# Show help
help:
	@echo "Flo Makefile targets:"
	@echo "  all      - Run lint, test, and build (default)"
	@echo "  build    - Build the flo binary to bin/flo"
	@echo "  test     - Run all tests"
	@echo "  lint     - Run golangci-lint"
	@echo "  clean    - Remove build artifacts"
	@echo "  install  - Install flo to GOPATH/bin"
	@echo "  help     - Show this help message"
