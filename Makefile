<<<<<<< HEAD
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
=======
# Makefile for Obsidian Automation Bot

.PHONY: all build start stop restart logs status update backup stats clean test

# Variables
BOT_NAME=obsidian-bot
BOT_IMAGE=obsidian-bot
ENV_FILE=.env
SRC_FILES=main.go processor.go health.go stats.go dedup.go ai_gemini.go organizer.go

all: build

# Build the Go application
build:
	@echo "🔨 Building Go application..."
	@go build -o $(BOT_NAME) $(SRC_FILES)

# Build the Docker image
build-docker:
	@echo "🔨 Building Docker image..."
	@docker build -t $(BOT_IMAGE) .

# Start the bot
start: build-docker
	@echo "🚀 Starting new container..."
	@docker run -d \
	  --name $(BOT_NAME) \
	  --restart unless-stopped \
	  -e TELEGRAM_BOT_TOKEN=$$(cat $(ENV_FILE) | grep TELEGRAM_BOT_TOKEN | cut -d '=' -f2) \
	  -e TZ=Africa/Tunis \
	  -v "$$(pwd)/vault:/app/vault" \
	  -v "$$(pwd)/attachments:/app/attachments" \
	  -v "$$(pwd)/stats.json:/app/stats.json" \
>>>>>>> main
	  -p 8080:8080 \
	  --memory="512m" \
	  --log-driver json-file \
	  --log-opt max-size=10m \
	  --log-opt max-file=3 \
<<<<<<< HEAD
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
=======
	  $(BOT_IMAGE)
	@echo ""
	@echo "✅ Bot started successfully!"

# Stop the bot
stop:
	@echo "🛑 Stopping and removing container..."
	@docker stop $(BOT_NAME) 2>/dev/null || true
	@docker rm $(BOT_NAME) 2>/dev/null || true
	@echo "✅ Bot stopped."

# Restart the bot
restart:
	@echo "🔄 Restarting bot..."
	@docker restart $(BOT_NAME)
	@sleep 2
	@make status

# View bot logs
logs:
	@docker logs -f --tail=50 $(BOT_NAME)

# Check bot status
status:
	@echo "=== Obsidian Bot Status ==="
	@echo ""
	@if docker ps | grep -q $(BOT_NAME); then \
		echo "✅ Bot is RUNNING"
		echo ""
		docker ps --format "table {{.Names}}\t{{.Status}}" | grep $(BOT_NAME)
		echo ""
		if curl -sf http://localhost:8080/health > /dev/null 2>&1; then \
			echo "✅ Health check: PASSING"
			curl -s http://localhost:8080/health
		else \
			echo "❌ Health check: FAILING"
		fi
		echo ""
		echo "Recent logs:"
		docker logs $(BOT_NAME) --tail=5
		echo ""
		if [ -f stats.json ]; then \
			echo "Statistics:"
			cat stats.json
		fi
	else \
		echo "❌ Bot is NOT running"
		echo "Start with: make start"
	fi

# Update the bot
update: stop start

# Backup the vault
backup:
	@echo "💾 Backing up vault..."
	@cd vault && \
	if [[ -n $$(git status -s 2>/dev/null) ]]; then \
		git add . && \
		git commit -m "Auto-backup: $$(date '+%Y-%m-%d %H:%M:%S')" && \
		echo "✅ Vault backed up" && \
		if git remote | grep -q origin; then \
			git push origin main 2>/dev/null && echo "✅ Pushed to GitHub"; \
		fi;
	else \
		echo "No changes to backup"; \
	fi

# View stats
stats:
	@echo "=== Processing Statistics ==="
	@echo ""
	@if [ -f stats.json ]; then \
		cat stats.json
	else \
		echo "No statistics yet"
	fi
	@echo ""
	@echo "=== Recent Files ==="
	@echo ""
	@echo "Inbox notes:"
	@ls -lht vault/Inbox/ 2>/dev/null | head -10 || echo "No notes yet"
	@echo ""
	@echo "Attachments:"
	@ls -lht attachments/ 2>/dev/null | head -10 || echo "No attachments yet"

# Test the bot
test:
	@echo "Testing Telegram API..."
	@curl -s "https://api.telegram.org/bot$$(cat $(ENV_FILE) | grep TELEGRAM_BOT_TOKEN | cut -d '=' -f2)/getMe" | jq .
	@echo ""
	@echo "Testing Go bot..."
	@go run $(SRC_FILES)

# Clean up old scripts
clean:
	@echo "🧹 Cleaning up old scripts..."
	@rm -f backup-vault.sh logs.sh quick-start.sh restart.sh setup-all.sh status.sh stop-bot.sh test-bot.sh update.sh view-stats.sh
	@echo "✅ Scripts removed."
>>>>>>> main
