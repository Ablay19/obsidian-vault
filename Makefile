# Makefile for Obsidian Automation Bot

.PHONY: all build up down logs status restart help onnx-build onnx-up

# Variables
IMAGE_NAME      ?= obsidian-bot
CONTAINER_NAME  ?= obsidian-bot
ENV_FILE        ?= .env
DOCKERFILE      ?= Dockerfile

# Default target: show help.
all: help

# Build the Docker image.
build:
	@echo "🔨 Building Docker image using $(DOCKERFILE)..."
	@docker build -f $(DOCKERFILE) -t $(IMAGE_NAME) .

# Run the application.
up: build
	@echo "🚀 Starting container $(CONTAINER_NAME)..."
	@if [ ! -f "$(ENV_FILE)" ]; then \
		echo "ERROR: $(ENV_FILE) file not found!"; \
		echo "Please create a .env file with TELEGRAM_BOT_TOKEN, TURSO_DATABASE_URL, TURSO_AUTH_TOKEN, and optionally ONNX_MODEL_PATH."; \
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
	  -p 8080:8080 \
	  $(IMAGE_NAME)
	@echo "✅ Bot $(CONTAINER_NAME) started successfully!"

# Stop and remove the container.
down:
	@echo "🛑 Stopping and removing container $(CONTAINER_NAME)..."
	@docker stop $(CONTAINER_NAME) >/dev/null 2>&1 || true
	@docker rm $(CONTAINER_NAME) >/dev/null 2>&1 || true
	@echo "✅ Bot $(CONTAINER_NAME) stopped and container removed."

# View container logs.
logs:
	@docker logs -f $(CONTAINER_NAME)

# Show container status.
status:
	@docker ps --filter "name=^/$(CONTAINER_NAME)$$"

# Restart the container.
restart: down up

# Build using Dockerfile.onnx
onnx-build: DOCKERFILE = Dockerfile.onnx
onnx-build: IMAGE_NAME = obsidian-bot-onnx
onnx-build: build ## Build the Docker image specifically for ONNX.

# Run using Dockerfile.onnx
onnx-up: DOCKERFILE = Dockerfile.onnx
onnx-up: IMAGE_NAME = obsidian-bot-onnx
onnx-up: CONTAINER_NAME = obsidian-bot-onnx
onnx-up: up ## Start the container specifically for ONNX.

# Show this help message.
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Environment Variables (in .env file or passed directly):"
	@echo "  TELEGRAM_BOT_TOKEN    (Required) Your Telegram bot token."
	@echo "  TURSO_DATABASE_URL    (Required) URL for your Turso database."
	@echo "  TURSO_AUTH_TOKEN      (Required) Auth token for your Turso database."
	@echo "  ONNX_MODEL_PATH       (Optional, Required for ONNX) Path to the ONNX model within the container (e.g., /app/models/model.onnx)."
