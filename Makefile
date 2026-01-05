.PHONY: all ci build test clean install lint fmt help install-lint

# Build variables
BINARY_NAME := mcp-tidy
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"
GO := go

# golangci-lint version
GOLANGCI_LINT_VERSION := v2.1.6

# Default target
all: test lint build

## ci: Run CI checks (install-lint + lint + test)
ci: install-lint lint test

# Build the binary
build:
	$(GO) build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/mcp-tidy

# Run tests
test:
	$(GO) test -v ./...

# Run tests with coverage
test-cover:
	$(GO) test -cover ./...

# Generate coverage report
coverage:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Install to GOPATH/bin
install:
	$(GO) install $(LDFLAGS) ./cmd/mcp-tidy

# Install golangci-lint
install-lint:
	$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

# Lint code
lint:
	golangci-lint run

# Format code
fmt:
	$(GO) fmt ./...
