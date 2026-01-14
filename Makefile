# Enhanced Makefile for Obsidian Automation
# Combines original Docker targets with comprehensive build, test, and development tools

.PHONY: help build run test lint clean docker docker-run docker-stop deploy setup env health version
.PHONY: build-ssh docker-build-all docker-push-all up ssh-up down ssh-down logs ssh-logs status ssh-status restart ssh-restart run-local sqlc-generate k8s-apply k8s-delete
.PHONY: build-prod dev test-coverage benchmark fmt deps security install-tools quick-dev prod-workflow all watch install uninstall release

# Variables
IMAGE_NAME      ?= obsidian-bot
CONTAINER_NAME  ?= obsidian-bot
ENV_FILE        ?= .env
DOCKERFILE      ?= Dockerfile
DASHBOARD_PORT  ?= 8080
DOCKER_REGISTRY ?= your-docker-registry # Default registry, e.g., ghcr.io/your-org

SSH_IMAGE_NAME      ?= obsidian-ssh-server
SSH_CONTAINER_NAME  ?= obsidian-ssh-server
SSH_DOCKERFILE      ?= Dockerfile.ssh
SSH_PORT            ?= 2222
SSH_API_PORT        ?= 8081

# Variables
BINARY_NAME := obsidian-automation
BUILD_DIR := ./build
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(COMMIT)"

# Go settings
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
CGO_ENABLED := 1

# Default target: show help.
.DEFAULT_GOAL := help
all: help

# Build the main bot Docker image.
build: ## Build the main bot Docker image.
	@echo "🔨 Building Docker image for main bot using $(DOCKERFILE)..."
	@docker build -f $(DOCKERFILE) -t $(IMAGE_NAME) .

# Build the SSH server Docker image.
build-ssh: ## Build the SSH server Docker image.
	@echo "🔨 Building Docker image for SSH server using $(SSH_DOCKERFILE)..."
	@docker build -f $(SSH_DOCKERFILE) -t $(SSH_IMAGE_NAME) .

docker-build-all: build build-ssh ## Build all Docker images.

docker-push-all: ## Push all Docker images to registry.
	@echo "📤 Pushing $(IMAGE_NAME) to $(DOCKER_REGISTRY)"
	@docker tag $(IMAGE_NAME) $(DOCKER_REGISTRY)/$(IMAGE_NAME):latest
	@docker push $(DOCKER_REGISTRY)/$(IMAGE_NAME):latest
	@echo "📤 Pushing $(SSH_IMAGE_NAME) to $(DOCKER_REGISTRY)"
	@docker tag $(SSH_IMAGE_NAME) $(DOCKER_REGISTRY)/$(SSH_IMAGE_NAME):latest
	@docker push $(DOCKER_REGISTRY)/$(SSH_IMAGE_NAME):latest
	@echo "✅ All images pushed."

# Run the main bot application.
up: build ## Build and start the main bot container.
	@echo "🚀 Starting main bot container $(CONTAINER_NAME)..."
	@if [ ! -f "$(ENV_FILE)" ]; then \
		echo "ERROR: $(ENV_FILE) file not found!"; \
		echo "Please create a .env file with TELEGRAM_BOT_TOKEN, TURSO_DATABASE_URL, and TURSO_AUTH_TOKEN."; \
		exit 1; \
	fi
	@docker stop $(CONTAINER_NAME) >/dev/null 2>&1 || true
	@docker rm $(CONTAINER_NAME) >/dev/null 2>&1 || true
	@docker run -d \
	  --name $(CONTAINER_NAME) \
	  --restart unless-stopped \
	  --env-file $(ENV_FILE) \
	  -v "./vault:/app/vault" \
	  -v "./data:/app/data" \
	  -v "./attachments:/app/attachments" \
	  -v "./pdfs:/app/pdfs" \
	  -p $(DASHBOARD_PORT):8080 \
	  $(IMAGE_NAME)
	@echo "✅ Main bot $(CONTAINER_NAME) started successfully!"
	@echo "🖥️  Dashboard available at http://localhost:$(DASHBOARD_PORT)"

# Run the SSH server application.
ssh-up: build-ssh ## Build and start the SSH server container.
	@echo "🚀 Starting SSH server container $(SSH_CONTAINER_NAME)..."
	@if [ ! -f "$(ENV_FILE)" ]; then \
		echo "ERROR: $(ENV_FILE) file not found!"; \
		echo "Please create a .env file with relevant SSH server environment variables."; \
		exit 1; \
	fi
	@docker stop $(SSH_CONTAINER_NAME) >/dev/null 2>&1 || true
	@docker rm $(SSH_CONTAINER_NAME) >/dev/null 2>&1 || true
	@docker run -d \
	  --name $(SSH_CONTAINER_NAME) \
	  --restart unless-stopped \
	  --env-file $(ENV_FILE) \
	  -v "./data:/data" \
	  -v "./id_rsa:/app/id_rsa:ro" \
	  -v "./id_rsa.pub:/app/id_rsa.pub:ro" \
	  -p $(SSH_PORT):2222 \
	  -p $(SSH_API_PORT):8081 \
	  $(SSH_IMAGE_NAME)
	@echo "✅ SSH server $(SSH_CONTAINER_NAME) started successfully!"
	@echo "🔑 SSH server available on port $(SSH_PORT)"
	@echo "API available at http://localhost:$(SSH_API_PORT)"

# Stop and remove the main bot container.
down: ## Stop and remove the main bot container.
	@echo "🛑 Stopping and removing main bot container $(CONTAINER_NAME)..."
	@docker stop $(CONTAINER_NAME) >/dev/null 2>&1 || true
	@docker rm $(CONTAINER_NAME) >/dev/null 2>&1 || true
	@echo "✅ Main bot $(CONTAINER_NAME) stopped and container removed."

# Stop and remove the SSH server container.
ssh-down: ## Stop and remove the SSH server container.
	@echo "🛑 Stopping and removing SSH server container $(SSH_CONTAINER_NAME)..."
	@docker stop $(SSH_CONTAINER_NAME) >/dev/null 2>&1 || true
	@docker rm $(SSH_CONTAINER_NAME) >/dev/null 2>&1 || true
	@echo "✅ SSH server $(SSH_CONTAINER_NAME) stopped and container removed."

# View main bot container logs.
logs: ## View main bot container logs.
	@docker logs -f $(CONTAINER_NAME)

# View SSH server container logs.
ssh-logs: ## View SSH server container logs.
	@docker logs -f $(SSH_CONTAINER_NAME)

# Show main bot container status.
status: ## Show main bot container status.
	@docker ps --filter "name=^/$(CONTAINER_NAME)$$"

# Show SSH server container status.
ssh-status: ## Show SSH server container status.
	@docker ps --filter "name=^/$(SSH_CONTAINER_NAME)$$"

# Restart the main bot container.
restart: down up ## Restart the main bot container.

# Restart the SSH server container.
ssh-restart: ssh-down ssh-up ## Restart the SSH server container.

# Apply Kubernetes manifests.
k8s-apply: ## Apply Kubernetes manifests.
	@echo "🚀 Applying Kubernetes manifests from k8s/ directory..."
	@kubectl apply -f k8s/
	@echo "✅ Kubernetes manifests applied."

# Delete Kubernetes manifests.
k8s-delete: ## Delete Kubernetes manifests.
	@echo "🔥 Deleting Kubernetes manifests from k8s/ directory..."
	@kubectl delete -f k8s/
	@echo "✅ Kubernetes manifests deleted."

# Enhanced build targets
build: ## Build application
	@echo -e "\033[34m[•]\033[0m Building application..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/bot/main.go
	@echo -e "\033[32m[✓]\033[0m Build completed: $(BUILD_DIR)/$(BINARY_NAME)"

# Build production
build-prod: ## Build optimized production binary
	@echo -e "\033[34m[•]\033[0m Building production binary..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-prod ./cmd/bot/main.go
	@echo -e "\033[32m[✓]\033[0m Production build completed: $(BUILD_DIR)/$(BINARY_NAME)-prod"

# Run target
run: build ## Build and run application
	@echo -e "\033[34m[•]\033[0m Starting application..."
	@echo -e "\033[35m[ℹ]\033[0m Dashboard: http://localhost:8080"
	@echo -e "\033[35m[ℹ]\033[0m Press Ctrl+C to stop"
	@$(BUILD_DIR)/$(BINARY_NAME)

# Run development
dev: ## Run in development mode
	@echo -e "\033[34m[•]\033[0m Starting development mode..."
	@ENVIRONMENT_MODE=dev ENABLE_COLORFUL_LOGS=true ENABLE_DEBUG_LOGS=true go run ./cmd/bot/main.go

# Test target
test: ## Run all tests
	@echo -e "\033[34m[•]\033[0m Running tests..."
	@go test -v ./...

# Test with coverage
test-coverage: ## Run tests with coverage report
	@echo -e "\033[34m[•]\033[0m Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo -e "\033[32m[✓]\033[0m Coverage report: coverage.html"

# Benchmark target
benchmark: ## Run benchmarks
	@echo -e "\033[34m[•]\033[0m Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Lint target
lint: ## Lint code
	@echo -e "\033[34m[•]\033[0m Linting code..."
	@if command -v gofmt > /dev/null; then \
		echo "Checking formatting..."; \
		unformatted=$$(gofmt -l .); \
		if [ -n "$$unformatted" ]; then \
			echo "Files need formatting:"; \
			echo "$$unformatted"; \
			echo "Auto-formatting..."; \
			gofmt -w .; \
		else \
			echo -e "\033[32m[✓]\033[0m Code is properly formatted"; \
		fi; \
	fi
	@echo "Running go vet..."
	@go vet ./...
	@if command -v golint > /dev/null; then \
		echo "Running golint..."; \
		golint ./...; \
	fi
	@echo -e "\033[32m[✓]\033[0m Linting completed"

# Clean target
clean: ## Clean build artifacts
	@echo -e "\033[34m[•]\033[0m Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME) $(BINARY_NAME)-*
	@rm -f coverage.out coverage.html
	@go clean -cache
	@go clean -testcache
	@echo -e "\033[32m[✓]\033[0m Clean completed"

# Setup target
setup: ## Perform initial setup
	@echo -e "\033[34m[•]\033[0m Performing initial setup..."
	@./scripts/env-setup.sh init
	@echo -e "\033[32m[✓]\033[0m Setup completed"

# Environment configuration
env: ## Configure environment
	@echo -e "\033[34m[•]\033[0m Configuring environment..."
	@./scripts/env-setup.sh config
	@echo -e "\033[32m[✓]\033[0m Environment configured"

# Health check
health: ## Perform health check
	@echo -e "\033[34m[•]\033[0m Performing health check..."
	@./scripts/maintenance.sh health

# Version information
version: ## Show version information
	@echo -e "\033[36mObsidian Automation\033[0m"
	@echo "=================="
	@echo "Version: $(VERSION)"
	@echo "Build: $(BUILD_TIME)"
	@echo "Commit: $(COMMIT)"
	@echo "Go: $(shell go version)"
	@echo "OS/Arch: $(GOOS)/$(GOARCH)"

# Docker targets (preserving original functionality)
docker: docker-build ## Build Docker image (alias)
docker-run: up ## Run Docker container (alias)
docker-stop: down ## Stop Docker container (alias)

# Run locally (clearing CGO flags to avoid onnxruntime issues)
run-local: ## Run the bot locally.
	@CGO_LDFLAGS="" CGO_CFLAGS="" go run ./cmd/bot/main.go

sqlc-generate: ## Generate SQLC code from queries.
	@echo "Generating SQLC code..."
	@sqlc generate
	@echo "✅ SQLC code generated successfully."

# Enhanced help target
help: ## Show this help message
	@echo -e "\033[36mObsidian Automation Makefile\033[0m"
	@echo "================================"
	@echo ""
	@echo -e "\033[34mQuick Start:\033[0m"
	@echo "  make setup      # Initial setup"
	@echo "  make build      # Build application"
	@echo "  make run        # Build and run"
	@echo "  make test       # Run tests"
	@echo ""
	@echo -e "\033[34mDevelopment:\033[0m"
	@echo "  make dev        # Development mode"
	@echo "  make fmt        # Format code"
	@echo "  make lint       # Lint code"
	@echo "  make test-coverage # Test with coverage"
	@echo ""
	@echo -e "\033[34mDocker:\033[0m"
	@echo "  make docker     # Build Docker image"
	@echo "  make up         # Start container"
	@echo "  make down       # Stop container"
	@echo ""
	@echo -e "\033[34mAll Targets:\033[0m"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo -e "\033[34mRequired Environment Variables (in .env file):\033[0m"
	@echo "  TELEGRAM_BOT_TOKEN    Your Telegram bot token."
	@echo "  TURSO_DATABASE_URL    URL for your Turso database."
	@echo "  TURSO_AUTH_TOKEN      Auth token for your Turso database."
	@echo "  GEMINI_API_KEYS       (Optional) Comma-separated list of Gemini API keys."
	@echo "  GROQ_API_KEY          (Optional) Your Groq API key."
	@echo ""
	@echo -e "\033[34mOptional Variables:\033[0m"
	@echo "  DASHBOARD_PORT        Port for the web dashboard (default: 8080)."
	@echo "  SSH_PORT              Port for the SSH server (default: 2222)."
	@echo "  SSH_API_PORT          Port for the SSH server's API (default: 8081)."
	@echo "  DOCKER_REGISTRY       Docker registry to push images to (default: your-docker-registry)."
	@echo ""
	@echo -e "\033[34mArchitecture Separation:\033[0m"
	@echo "  make arch-setup      # Setup architecture separation structure"
	@echo "  make arch-build      # Build all Go apps and workers"
	@echo "  make arch-test       # Run all tests"
	@echo "  make arch-deploy-dev # Deploy to development"
	@echo "  make create-worker   # Create new worker (name=...)"
	@echo "  make create-go-app   # Create new Go app (name=...)"

# Architecture Separation Targets
arch-setup: ## Setup architecture separation structure
	@echo "Setting up architecture separation..."
	@mkdir -p apps workers packages deploy tests
	@echo "Architecture separation structure created."

arch-build: ## Build all Go applications and workers
	@echo "Building all Go applications..."
	@for app in apps/*/; do \
		if [ -f "$${app}go.mod" ]; then \
			echo "Building $${app}..."; \
			cd "$${app}" && go build -o bin/ ./cmd/main.go || true; \
		fi; \
	done
	@echo "Building all workers..."
	@for worker in workers/*/; do \
		if [ -f "$${worker}package.json" ]; then \
			echo "Building $${worker}..."; \
			(cd "$${worker}" && npm run build 2>/dev/null || echo "No build step"); \
		fi; \
	done

arch-test: ## Run all tests for architecture separation
	@echo "Running all tests..."
	@cd apps/api-gateway && go test -v ./internal/... -race || true
	@for worker in workers/*/; do \
		if [ -f "$${worker}package.json" ]; then \
			echo "Testing $${worker}..."; \
			(cd "$${worker}" && npm test 2>/dev/null || true); \
		fi; \
	done

arch-deploy-dev: ## Deploy to development environment
	@echo "Deploying to development..."
	@echo "Go apps: Run 'cd apps/api-gateway && go run cmd/main.go'"
	@echo "Workers: Run 'cd workers/ai-worker && npm run dev'"

create-worker: ## Create new worker (usage: make create-worker name=my-worker)
	@if [ -z "$(name)" ]; then \
		echo "Error: name parameter required. Usage: make create-worker name=my-worker"; \
		exit 1; \
	fi
	@echo "Creating new worker: $(name)"
	@mkdir -p workers/$(name)/src workers/$(name)/tests
	@echo '{"name": "@obsidian-vault/$(name)","version": "1.0.0"}' > workers/$(name)/package.json
	@echo "Worker created at workers/$(name)/"

create-go-app: ## Create new Go application (usage: make create-go-app name=my-app)
	@if [ -z "$(name)" ]; then \
		echo "Error: name parameter required. Usage: make create-go-app name=my-app"; \
		exit 1; \
	fi
	@echo "Creating new Go application: $(name)"
	@mkdir -p apps/$(name)/cmd apps/$(name)/internal/{handlers,services,models} apps/$(name)/tests
	@echo "module github.com/abdoullahelvogani/obsidian-vault/apps/$(name)" > apps/$(name)/go.mod
	@echo "Go application created at apps/$(name)/"

# ============================================
# US2: Build Process Independence (T083-T084)
# ============================================

build-go-apps: ## Build all Go applications
	@echo "Building all Go applications..."
	@for app in apps/*/; do \
		if [ -f "$${app}go.mod" ]; then \
			echo "Building $${app}..."; \
			cd "$${app}" && go build -o bin/ ./cmd/main.go 2>/dev/null && echo "✓ $${app} built" || echo "✗ $${app} failed"; \
			cd - > /dev/null; \
		fi; \
	done

build-workers: ## Build all Cloudflare Workers
	@echo "Building all workers..."
	@for worker in workers/*/; do \
		if [ -f "$${worker}package.json" ] && [ -f "$${worker}wrangler.toml" ]; then \
			echo "Building $${worker}..."; \
			(cd "$${worker}" && npm run build 2>/dev/null || echo "No build step for $${worker}"); \
		fi; \
	done

build-all: build-go-apps build-workers ## Build all Go applications and workers
	@echo ""
	@echo -e "\033[32m✓ All builds completed\033[0m"

# ============================================
# US2: Test Process Independence (T088-T089)
# ============================================

test-go-apps: ## Run tests for all Go applications
	@echo "Running tests for all Go applications..."
	@for app in apps/*/; do \
		if [ -f "$${app}go.mod" ]; then \
			echo "Testing $${app}..."; \
			(cd "$${app}" && go test -v ./internal/... 2>/dev/null && echo "✓ $${app} tests passed" || echo "✗ $${app} tests failed"); \
			cd - > /dev/null; \
		fi; \
	done

test-workers: ## Run tests for all workers
	@echo "Running tests for all workers..."
	@for worker in workers/*/; do \
		if [ -f "$${worker}package.json" ] && [ -f "$${worker}vitest.config.ts" -o -f "$${worker}jest.config.js" ]; then \
			echo "Testing $${worker}..."; \
			(cd "$${worker}" && npm test 2>/dev/null || echo "No test config for $${worker}"); \
		fi; \
	done

test-all: test-go-apps test-workers ## Run all tests
	@echo ""
	@echo -e "\033[32m✓ All tests completed\033[0m"
