.PHONY: help build test lint clean run docker-build docker-push proto

# Variables
notification-service := notification-service
DOCKER_REGISTRY ?= ghcr.io/longvhv
VERSION ?= $(shell git describe --tags --always --dirty)
GO_VERSION := 1.24

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the service
	@echo "Building $(notification-service)..."
	@go build -o bin/$(notification-service) ./cmd/main.go
	@echo "Build complete!"

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	@go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linters
	@echo "Running linters..."
	@golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@gofmt -s -w .

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/ dist/ coverage.* *.out
	@go clean -testcache
	@echo "Clean complete!"

run: ## Run the service locally
	@echo "Running $(notification-service)..."
	@go run ./cmd/main.go

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

proto: ## Generate protobuf files (if applicable)
	@if [ -d "proto" ]; then \
		echo "Generating protobuf files..."; \
		protoc --go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			proto/*.proto; \
	fi

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_REGISTRY)/$(notification-service):$(VERSION) .
	@docker tag $(DOCKER_REGISTRY)/$(notification-service):$(VERSION) $(DOCKER_REGISTRY)/$(notification-service):latest
	@echo "Docker image built: $(DOCKER_REGISTRY)/$(notification-service):$(VERSION)"

docker-push: docker-build ## Push Docker image
	@echo "Pushing Docker image..."
	@docker push $(DOCKER_REGISTRY)/$(notification-service):$(VERSION)
	@docker push $(DOCKER_REGISTRY)/$(notification-service):latest
	@echo "Docker image pushed!"

docker-run: ## Run Docker container locally
	@echo "Running Docker container..."
	@docker run --rm -p 8080:8080 -p 50051:50051 \
		--name $(notification-service) \
		$(DOCKER_REGISTRY)/$(notification-service):latest

install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@if [ -d "proto" ]; then \
		go install google.golang.org/protobuf/cmd/protoc-gen-go@latest; \
		go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest; \
	fi
	@echo "Tools installed!"

.DEFAULT_GOAL := help
