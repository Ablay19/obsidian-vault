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
	  -p 8080:8080 \
	  --memory="512m" \
	  --log-driver json-file \
	  --log-opt max-size=10m \
	  --log-opt max-file=3 \
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
