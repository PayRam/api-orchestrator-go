.PHONY: help build run test clean migrate fmt vet

help: ## Display this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the example application
	@echo "Building example application..."
	go build -o bin/api-orchestrator ./cmd/example/main.go

run: ## Run the example application
	@echo "Running example application..."
	go run ./cmd/example/main.go

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	go clean

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

lint: ## Run golangci-lint (requires golangci-lint installed)
	@echo "Running linter..."
	golangci-lint run

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

migrate: ## Run database migrations (requires DB to be running)
	@echo "Running migrations..."
	@echo "Not implemented yet - will be part of Phase 7"

.DEFAULT_GOAL := help
