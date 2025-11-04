.PHONY: help build test lint clean install integration-test run

# Default target
help:
	@echo "Yoink - Credential Revocation Tool"
	@echo ""
	@echo "Available targets:"
	@echo "  make build             - Build the yoink binary"
	@echo "  make test              - Run unit tests"
	@echo "  make integration-test  - Run integration tests"
	@echo "  make lint              - Run linter"
	@echo "  make clean             - Remove build artifacts"
	@echo "  make install           - Install yoink to GOPATH/bin"
	@echo "  make run               - Run yoink (usage: make run ARGS='--help')"
	@echo "  make all               - Run tests, lint, and build"

# Build the binary
build:
	@echo "Building yoink..."
	@mkdir -p bin
	@go build -o bin/yoink ./cmd/yoink
	@echo "✓ Build complete: bin/yoink"

# Run unit tests
test:
	@echo "Running unit tests..."
	@go test ./... -v -cover

# Run integration tests
integration-test: build
	@echo "Running integration tests..."
	@bash test/integration_test.sh

# Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping lint"; \
		echo "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@echo "✓ Clean complete"

# Install to GOPATH/bin
install:
	@echo "Installing yoink..."
	@go install ./cmd/yoink
	@echo "✓ Install complete"

# Run the binary
run: build
	@./bin/yoink $(ARGS)

# Run all checks
all: test lint build
	@echo "✓ All checks passed"
