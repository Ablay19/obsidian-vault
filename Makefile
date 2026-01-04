# Makefile for Obsidian Automation Bot

.PHONY: all build build-ssh docker-build-all docker-push-all up ssh-up down ssh-down logs ssh-logs status ssh-status restart ssh-restart run-local sqlc-generate k8s-apply k8s-delete help

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

# Default target: show help.
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

# Run locally (clearing CGO flags to avoid onnxruntime issues)
run-local: ## Run the bot locally.
	@CGO_LDFLAGS="" CGO_CFLAGS="" go run ./cmd/bot/main.go

sqlc-generate: ## Generate SQLC code from queries.
	@echo "Generating SQLC code..."
	@sqlc generate
	@echo "✅ SQLC code generated successfully."

# Show this help message.
help: ## Show this help message.
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Required Environment Variables (in .env file):"
	@echo "  TELEGRAM_BOT_TOKEN    Your Telegram bot token."
	@echo "  TURSO_DATABASE_URL    URL for your Turso database."
	@echo "  TURSO_AUTH_TOKEN      Auth token for your Turso database."
	@echo "  GEMINI_API_KEYS       (Optional) Comma-separated list of Gemini API keys."
	@echo "  GROQ_API_KEY          (Optional) Your Groq API key."
	@echo ""
	@echo "Optional Variables:"
	@echo "  DASHBOARD_PORT        Port for the web dashboard (default: 8080)."
	@echo "  SSH_PORT              Port for the SSH server (default: 2222)."
	@echo "  SSH_API_PORT          Port for the SSH server's API (default: 8081)."
	@echo "  DOCKER_REGISTRY       Docker registry to push images to (default: your-docker-registry)."
