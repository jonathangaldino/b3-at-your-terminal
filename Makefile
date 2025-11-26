.PHONY: help build test clean install fmt lint run dev tidy

# Variables
BINARY_NAME=b3cli
MAIN_PATH=./cmd/b3cli
GO=go
GOFLAGS=-v

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: ./$(BINARY_NAME)"

test: ## Run all tests
	@echo "Running tests..."
	$(GO) test -v ./...

test-cover: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

install: ## Install the binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(MAIN_PATH)
	@echo "Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

fmt: ## Format Go code
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "Format complete"

lint: ## Run linter (requires golangci-lint)
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

tidy: ## Tidy go modules
	@echo "Tidying modules..."
	$(GO) mod tidy
	@echo "Tidy complete"

run: build ## Build and run the application
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

dev: ## Run without building binary (useful for development)
	$(GO) run $(MAIN_PATH)

# Wallet commands (shortcuts for development)
wallet-create: build ## Create a new wallet
	./$(BINARY_NAME) wallet create

wallet-open: build ## Open existing wallet
	./$(BINARY_NAME) wallet open

parse: build ## Parse Excel file
	./$(BINARY_NAME) parse

assets: build ## Show assets
	./$(BINARY_NAME) assets

# CI/CD targets
ci-test: ## Run tests in CI mode
	$(GO) test -race -coverprofile=coverage.out -covermode=atomic ./...

ci-build: ## Build for CI
	$(GO) build -ldflags="-w -s" -o $(BINARY_NAME) $(MAIN_PATH)

# Multi-platform builds
build-all: ## Build for multiple platforms
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 $(GO) build -o $(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 $(GO) build -o $(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GO) build -o $(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 $(GO) build -o $(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Multi-platform build complete"
