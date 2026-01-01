# AI-Powered Obsidian Automation Bot

A powerful, AI-enhanced Telegram bot to automate your note-taking workflow with Obsidian. Send images, PDFs, or just chat with the bot, and it will intelligently process the content, create organized notes, and even stream responses back to you in real-time.

## ✨ Features

-   **AI-Powered Content Analysis**: Uses Google's Gemini Pro and **Groq** to summarize text, answer questions, and categorize content.
-   **Multi-Provider AI Support**: Seamlessly switch between Google Gemini and Groq AI providers, offering flexibility and optimized performance.
-   **Streaming Responses**: Get real-time answers from the AI, just like a modern chatbot.
-   **Chatbot Mode**: Chat directly with the bot for quick questions and answers.
-   **Multi-Language Support**: Defaults to French, but you can change the language at any time with the `/lang` command.
-   **File Processing**: Extracts text from images (via Tesseract OCR) and PDFs.
-   **Improved PDF Conversion**: The bot now uses a headless Chrome instance to convert Markdown notes to PDF, ensuring high-fidelity rendering of complex notes, including those with LaTeX.
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

# Your Groq API Key (Required for Groq provider)
GROQ_API_KEY=your-groq-api-key

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
-   `/setprovider <provider>`: Sets the AI provider (e.g., `/setprovider Groq`).
-   `/switchkey`: Manually switch to the next Gemini API key.
-   `/last`: Shows the path of the last note created.
-   `/reprocess`: Reprocesses the last file you sent.
-   `/pid`: Shows the process ID of the bot instance.

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

## 🧑‍💻 Development

### Running Tests

To run the unit tests, use the following command:

```bash
go test ./...
```

### Linting

This project uses `golangci-lint` to enforce a consistent code style. To run the linter locally, use the following command:

```bash
golangci-lint run
```

The linter is also run automatically as part of the CI/CD pipeline.


## ⚙️ Technical Details

### Architecture

```
Telegram User ↔ Telegram Bot API ↔ Go Application (Docker) ↔ AI Providers (Gemini/Groq)
                                        │
                                        └─► Obsidian Vault
```

### File Structure

```
obsidian-automation/
├── cmd/bot/main.go             # Main application entrypoint
├── internal/
│   ├── ai/                     # AI service and providers
│   ├── bot/                    # Core bot logic, handlers, and processor
│   ├── converter/              # File conversion utilities
│   ├── database/               # Database schema and queries
│   ├── health/                 # Health and control endpoints
│   └── pid/                    # PID lock mechanism
├── Dockerfile                  # Container definition
├── Makefile                    # Project management commands
├── .env                        # Environment variables (gitignored)
├── vault/                      # Your Obsidian vault
│   ├── attachments/            # Raw files received by the bot
├── .github/workflows/          # CI/CD pipelines
│   ├── ci.yml                  # Run tests and build
│   └── lint.yml                # Run linter
```

### Bot Instance Management

The bot uses a PID file lock mechanism to prevent multiple instances from running simultaneously, avoiding conflicts with the Telegram API. For more technical details, refer to `GEMINI.md`.

### Observability

The bot provides several endpoints for monitoring and control on port 8080:

-   `/status`: Returns a JSON object with the current status of the bot, including PID, uptime, Go version, OS, architecture, and last activity time.
-   `/pause`: Pauses the bot's listening activity.
-   `/resume`: Resumes the bot's listening activity.