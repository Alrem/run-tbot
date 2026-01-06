# Makefile for run-tbot - Educational Telegram Bot
# Provides convenient commands for development, testing, and building

# =============================================================================
# Configuration
# =============================================================================

# Binary name for the compiled application
BINARY_NAME=run-tbot

# Build output directory
BUILD_DIR=bin

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOFMT=gofmt
GOVET=$(GOCMD) vet

# Build flags for optimized binary
# -ldflags "-w -s" strips debug info and symbol table (smaller binary)
# -trimpath removes file system paths from binary (reproducible builds)
LDFLAGS=-ldflags "-w -s"
BUILD_FLAGS=-trimpath $(LDFLAGS)

# Test flags
# -v: verbose output
# -race: enable race detector (finds concurrent access bugs)
# -coverprofile: generate coverage report
# -covermode=atomic: coverage mode for race detector
TEST_FLAGS=-v -race -coverprofile=coverage.out -covermode=atomic

# Coverage output file
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Docker image name (for local testing)
DOCKER_IMAGE=run-tbot:local

# =============================================================================
# PHONY Targets
# =============================================================================
# .PHONY declares targets that don't create files
# Without this, make might think "test" is a file and skip the command

.PHONY: help test test-unit test-integration build run lint fmt clean coverage docker-build docker-run all

# =============================================================================
# Default Target
# =============================================================================

# Default target when you just run "make"
# Shows help message with all available commands
help: ## Show this help message
	@echo "run-tbot - Educational Telegram Bot"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

# =============================================================================
# Testing Targets
# =============================================================================

test: ## Run all tests with coverage and race detector
	@echo "Running all tests with coverage..."
	$(GOTEST) $(TEST_FLAGS) ./...
	@echo ""
	@echo "Coverage report generated: $(COVERAGE_FILE)"
	@echo "Run 'make coverage' to view in browser"

test-unit: ## Run only unit tests (exclude integration tests)
	@echo "Running unit tests..."
	$(GOTEST) -v -race ./... -run '^Test[^R]' -short
	@echo "Unit tests completed"

test-integration: ## Run only integration tests
	@echo "Running integration tests..."
	$(GOTEST) -v -race ./handlers -run '^TestRouteUpdate'
	@echo "Integration tests completed"

coverage: ## Generate and open coverage report in browser
	@echo "Generating HTML coverage report..."
	$(GOTEST) $(TEST_FLAGS) ./...
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Opening coverage report in browser..."
	@if command -v open > /dev/null 2>&1; then \
		open $(COVERAGE_HTML); \
	elif command -v xdg-open > /dev/null 2>&1; then \
		xdg-open $(COVERAGE_HTML); \
	else \
		echo "Coverage report saved to $(COVERAGE_HTML)"; \
	fi

# =============================================================================
# Build Targets
# =============================================================================

build: ## Build the binary for current platform
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

build-linux: ## Build Linux binary (for deployment)
	@echo "Building $(BINARY_NAME) for Linux..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux .
	@echo "Linux binary built: $(BUILD_DIR)/$(BINARY_NAME)-linux"

# =============================================================================
# Development Targets
# =============================================================================

run: ## Run the bot locally (requires .env file)
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found"; \
		echo "Create .env file with:"; \
		echo "  BOT_TOKEN=your_token_here"; \
		echo "  PORT=8080"; \
		echo "  ENVIRONMENT=development"; \
		echo "  ALLOWED_USERS=your_user_id"; \
		echo ""; \
		echo "Or copy from example: cp .env.example .env"; \
		exit 1; \
	fi
	@echo "Loading .env and running bot..."
	@export $$(cat .env | xargs) && $(GOCMD) run .

fmt: ## Format code with gofmt
	@echo "Formatting code..."
	$(GOFMT) -w -s .
	@echo "Code formatted"

lint: fmt ## Run linters (gofmt + go vet)
	@echo "Running go vet..."
	$(GOVET) ./...
	@echo "Linting completed"

# =============================================================================
# Cleanup Targets
# =============================================================================

clean: ## Remove build artifacts and test coverage files
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@echo "Cleanup completed"

# =============================================================================
# Docker Targets
# =============================================================================

docker-build: ## Build Docker image locally
	@echo "Building Docker image: $(DOCKER_IMAGE)..."
	docker build -t $(DOCKER_IMAGE) .
	@echo "Docker image built: $(DOCKER_IMAGE)"

docker-run: docker-build ## Build and run Docker container locally
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found"; \
		echo "Create .env file before running Docker container"; \
		exit 1; \
	fi
	@echo "Running Docker container..."
	docker run --rm -it \
		--env-file .env \
		-p 8080:8080 \
		$(DOCKER_IMAGE)

# =============================================================================
# CI/CD Targets
# =============================================================================

ci: lint test build ## Run CI checks locally (lint, test, build)
	@echo ""
	@echo "✅ CI checks passed!"
	@echo ""

all: clean lint test build ## Clean, lint, test, and build
	@echo ""
	@echo "✅ All tasks completed!"
	@echo ""

# =============================================================================
# Educational Notes
# =============================================================================
#
# Makefile syntax:
#   target: dependencies ## help text
#       command
#
# Special variables:
#   $@ - target name
#   $< - first dependency
#   $^ - all dependencies
#
# @ prefix: suppress command echo (only show output)
# - prefix: ignore errors
#
# Common patterns:
#   .PHONY: target - declare non-file target
#   target: dep1 dep2 - target depends on dep1 and dep2
#   $(VAR) - use variable
#   $$(cmd) - shell command substitution
#
# Tips:
#   - Use tabs (not spaces) for indentation
#   - Variables make Makefile maintainable
#   - Comments with ## appear in help
#   - Chain commands with && for error handling
#
# =============================================================================
