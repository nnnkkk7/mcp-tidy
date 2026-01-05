.PHONY: all ci build test clean install lint fmt help

# Build variables
BINARY_NAME := mcp-tidy
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"
GO := go

# Default target
all: test build

## ci: Run CI checks (lint + test)
ci: lint test

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

# Lint code
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, running go vet instead"; \
		$(GO) vet ./...; \
	fi

# Format code
fmt:
	$(GO) fmt ./...
