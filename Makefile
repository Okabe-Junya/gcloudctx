.PHONY: help build install test test-coverage lint fmt clean run

BINARY_NAME=gcloudctx
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X github.com/Okabe-Junya/gcloudctx/cmd.Version=$(VERSION)"

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary into bin/
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) .

install: ## Install the binary with go install
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) .

test: ## Run tests with the race detector and coverage
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Generate an HTML coverage report
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run golangci-lint
	@echo "Running linters..."
	golangci-lint run ./...

fmt: ## Format and simplify Go sources
	@echo "Formatting code..."
	gofmt -s -w .

clean: ## Remove build and coverage artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf dist/
	rm -f coverage.out coverage.html

run: build ## Build and run the binary
	./bin/$(BINARY_NAME)

.DEFAULT_GOAL := help
