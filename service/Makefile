.PHONY: build run clean migrate test docker-build docker-run help

# Variables
APP_NAME=rvcs
BUILD_DIR=./build
DOCKER_IMAGE=$(APP_NAME):latest

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(APP_NAME) cmd/server/main.go
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

run: ## Run the application
	@go run cmd/server/main.go

clean: ## Clean build files
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

migrate: ## Run database migrations
	@go run cmd/server/main.go migrate

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

tidy: ## Tidy dependencies
	@echo "Tidying dependencies..."
	@go mod tidy

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -f deployments/docker/Dockerfile -t $(DOCKER_IMAGE) .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run -p 8080:8080 $(DOCKER_IMAGE)

docker-compose-up: ## Start services with Docker Compose
	@echo "Starting services..."
	@docker-compose up -d

docker-compose-down: ## Stop services with Docker Compose
	@echo "Stopping services..."
	@docker-compose down

install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
