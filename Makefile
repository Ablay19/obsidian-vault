# Makefile for Obsidian Automation Bot

.PHONY: all build up down logs status restart help

# Variables
IMAGE_NAME      ?= obsidian-bot
CONTAINER_NAME  ?= obsidian-bot
ENV_FILE        ?= .env

# Default target: show help.
all: help

# Build the Docker image.
build:
	@echo "🔨 Building Docker image..."
	@docker build -t $(IMAGE_NAME) .

# Run the application.
up: build
	@echo "🚀 Starting container..."
	@if [ ! -f "$(ENV_FILE)" ]; then \
		echo "ERROR: $(ENV_FILE) file not found!"; \
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
	@echo "✅ Bot started successfully!"

# Stop and remove the container.
down:
	@echo "🛑 Stopping and removing container..."
	@docker stop $(CONTAINER_NAME) >/dev/null 2>&1 || true
	@docker rm $(CONTAINER_NAME) >/dev/null 2>&1 || true
	@echo "✅ Bot stopped and container removed."

# View container logs.
logs:
	@docker logs -f $(CONTAINER_NAME)

# Show container status.
status:
	@docker ps --filter "name=^/$(CONTAINER_NAME)$$"

# Restart the container.
restart: down up

# Show this help message.
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: model-info
model-info:
	@echo "Testing model info endpoint..."
	curl -s http://localhost:8080/info | jq .

.PHONY: test-external-apis
test-external-apis:
	@echo "Testing external AI APIs..."
	./test_external_apis.sh

.PHONY: test-endpoint
test-endpoint:
	 @echo "Testing local endpoint..."
	curl -X POST http://localhost:8080/test-process \
		-H "Content-Type: application/json" \
		-d '{"text":"Test message","language":"English"}'

.PHONY: test-file
test-file:
	 @echo "Testing file processing..."
	curl -X POST http://localhost:8081/test-process \
		-H "Content-Type: application/json" \
		-d '{"file_path":"./test.pdf","language":"French"}'
