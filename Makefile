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

## lint: Run golangci-lint (install if needed)
lint: install-lint-deps
	$(GOLANGCI_LINT) run

## lint-fix: Run golangci-lint with auto-fix (install if needed)
lint-fix: install-lint-deps
	$(GOLANGCI_LINT) run --fix

## install-lint-deps: Install golangci-lint if not present
install-lint-deps:
	@if ! command -v $(GOLANGCI_LINT) &> /dev/null; then \
		echo "golangci-lint not found. Installing $(GOLANGCI_LINT_VERSION)..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION); \
	fi

# Format code
fmt:
	$(GO) fmt ./...
