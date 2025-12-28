# AI-Powered Obsidian Automation Bot

A powerful, AI-enhanced Telegram bot to automate your note-taking workflow with Obsidian. Send images, PDFs, or just chat with the bot, and it will intelligently process the content, create organized notes, and even stream responses back to you in real-time.

## ✨ Features

-   **AI-Powered Content Analysis**: Uses Google's Gemini Pro to summarize text, answer questions, and categorize content.
-   **Streaming Responses**: Get real-time answers from the AI, just like a modern chatbot.
-   **Chatbot Mode**: Chat directly with the bot for quick questions and answers.
-   **Multi-Language Support**: Defaults to French, but you can change the language at any time with the `/lang` command.
-   **File Processing**: Extracts text from images (via Tesseract OCR) and PDFs.
-   **Intelligent Categorization**: Automatically categorizes content into topics like `physics`, `math`, `chemistry`, etc.
-   **Duplicate Detection**: Prevents processing the same file twice.
-   **Interactive Commands**: Manage the bot and your notes with a rich set of slash commands.
-   **Dockerized**: Easy to set up and run in a containerized environment.

## 🚀 Getting Started

### Prerequisites

-   Docker
-   `make`
-   Go (for development)

### 1. Configuration

Create a `.env` file in the root of the project and add your credentials:

```dotenv
# Your Telegram Bot Token (Required)
TELEGRAM_BOT_TOKEN=your-token-goes-here

# Comma-separated list of your Gemini API Keys (Required for AI features)
# The bot will automatically rotate keys if one hits its quota.
GEMINI_API_KEYS=key-1,key-2,key-3

# Host for Ollama if you have it (Optional, for future fallback use)
OLLAMA_HOST=http://localhost:11434
```

### 2. Running the Bot

The project is managed with a `Makefile` for simplicity.

```bash
# Build and start the bot in the background
make up

# View the bot's logs
make logs

# Stop the bot
make down
```

## 🤖 Usage

### Sending Files

Simply send an image or a PDF file to the bot in your Telegram chat. The bot will process it, stream a summary back to you, and create a new note in your Obsidian vault.

### Chatting with the Bot

Send any text message to the bot to start a conversation. It will respond using the AI in the language you've configured.

### Bot Commands

-   `/start` or `/help`: Shows the welcome message and list of commands.
-   `/stats`: Displays usage statistics.
-   `/lang <language>`: Sets the AI's response language (e.g., `/lang English`).
-   `/switchkey`: Manually switch to the next Gemini API key.
-   `/last`: Shows the path of the last note created.
-   `/reprocess`: Reprocesses the last file you sent.

## 🛠️ Project Management

Use `make` to manage the bot's lifecycle.

```bash
# Show all available commands
make help

# Build the Docker image
make build

# Start the application (builds first if needed)
make up

# Stop and remove the container
make down

# View live logs
make logs

# Check the status of the container
make status

# Restart the container
make restart

# Run the vault backup script
make backup
```

## ⚙️ Technical Details

### Architecture

```
Telegram User ↔ Telegram Bot API ↔ Go Application (Docker) ↔ Gemini API
                                        │
                                        └─► Obsidian Vault
```

### File Structure

```
obsidian-automation/
├── main.go              # Main bot logic and command handler
├── processor.go         # File processing and AI orchestration
├── ai_service.go        # Gemini AI service
├── stats.go             # Statistics tracking
├── health.go            # Health check endpoint
├── dedup.go             # Duplicate file detection
├── Dockerfile           # Container definition
├── Makefile             # Project management commands
├── .env                 # Environment variables (gitignored)
├── vault/               # Your Obsidian vault
└── attachments/         # Raw files received by the bot
```

