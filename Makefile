# Makefile for Obsidian Automation Bot

.PHONY: all build up down logs status restart help 

# Variables
IMAGE_NAME      ?= obsidian-bot
CONTAINER_NAME  ?= obsidian-bot
ENV_FILE        ?= .env
DOCKERFILE      ?= Dockerfile
DASHBOARD_PORT  ?= 8080

# Default target: show help.
all: help

# Build the Docker image.
build: ## Build the Docker image.
	@echo "🔨 Building Docker image using $(DOCKERFILE)..."
	@docker build -f $(DOCKERFILE) -t $(IMAGE_NAME) .

# Run the application.
up: build ## Build and start the container.
	@echo "🚀 Starting container $(CONTAINER_NAME)..."
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
	@echo "✅ Bot $(CONTAINER_NAME) started successfully!"
	@echo "🖥️  Dashboard available at http://localhost:$(DASHBOARD_PORT)"

# Stop and remove the container.
down: ## Stop and remove the container.
	@echo "🛑 Stopping and removing container $(CONTAINER_NAME)..."
	@docker stop $(CONTAINER_NAME) >/dev/null 2>&1 || true
	@docker rm $(CONTAINER_NAME) >/dev/null 2>&1 || true
	@echo "✅ Bot $(CONTAINER_NAME) stopped and container removed."

# View container logs.
logs: ## View container logs.
	@docker logs -f $(CONTAINER_NAME)

# Show container status.
status: ## Show container status.
	@docker ps --filter "name=^/$(CONTAINER_NAME)$$"

# Restart the container.
restart: down up ## Restart the container.

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
