# Simple Makefile for managing the Obsidian Bot Docker container.

.PHONY: all build up stop rm clean logs status restart health help backup

# Customizable variables
IMAGE_NAME      ?= obsidian-bot
CONTAINER_NAME  ?= obsidian-bot
DOCKER_BUILD_ARGS ?= "" # Example: make build DOCKER_BUILD_ARGS="--no-cache"
ENV_FILE        ?= .env

# Default target: show help.
all: help

build: ## Build the Docker image.
	@echo "🔨 Building Docker image..."
	@docker build $(DOCKER_BUILD_ARGS) -t $(IMAGE_NAME) .

up: clean build ## Run the application. Stops and removes any existing container.
	@echo "🚀 Starting new container..."
	@if [ ! -f "$(ENV_FILE)" ]; then \
		echo "ERROR: $(ENV_FILE) file not found!"; \
		exit 1; \
	fi
	@set -a; source $(ENV_FILE); set +a; \
	if [ -z "$$TELEGRAM_BOT_TOKEN" ]; then \
		echo "ERROR: TELEGRAM_BOT_TOKEN not set in $(ENV_FILE)"; \
		exit 1; \
	fi; \
	if [ -z "$$GEMINI_API_KEY" ]; then \
		echo "WARNING: GEMINI_API_KEY not set in $(ENV_FILE). Gemini AI will be unavailable."; \
	fi; \
	docker run -d \
	  --name $(CONTAINER_NAME) \
	  --restart unless-stopped \
	  -e TELEGRAM_BOT_TOKEN \
	  -e GEMINI_API_KEY \
	  -e OLLAMA_HOST \
	  -e TZ=Africa/Tunis \
	  -v "$(shell pwd)/vault:/app/vault" \
	  -v "$(shell pwd)/attachments:/app/attachments" \
	  -v "$(shell pwd)/stats.json:/app/stats.json" \
	  -p 8080:8080 \
	  --memory="512m" \
	  --log-driver json-file \
	  --log-opt max-size=10m \
	  --log-opt max-file=3 \
	  $(IMAGE_NAME)
	@echo ""
	@echo "✅ Bot started successfully!"

stop: ## Stop the running container.
	@echo "🛑 Stopping container..."
	@docker stop $(CONTAINER_NAME) >/dev/null 2>&1 || true

rm: ## Remove the stopped container.
	@echo "🗑️ Removing container..."
	@docker rm $(CONTAINER_NAME) >/dev/null 2>&1 || true

clean: stop rm ## Stop and remove the container.

restart: ## Restart the container.
	@echo "🔄 Restarting container..."
	@docker restart $(CONTAINER_NAME)

logs: ## View container logs.
	@docker logs -f $(CONTAINER_NAME)

status: ## Show container status.
	@docker ps --filter "name=^/$(CONTAINER_NAME)$$"

health: ## Check the application's health.
	@echo "Checking health endpoint..."
	@curl --fail -s http://localhost:8080/health && echo "OK" || (echo "Health check failed!" && exit 1)

backup: ## Backup the vault.
	@if [ -f ./backup-vault.sh ]; then \
		echo "Backing up vault..."; \
		./backup-vault.sh; \
	else \
		echo "backup-vault.sh not found."; \
	fi

help: ## Show this help message.
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%%-20s\033[0m %%s\n", $$1, $$2}'