# AI-Powered Obsidian Automation Bot

A powerful, AI-enhanced Telegram bot to automate your note-taking workflow with Obsidian. Send images, PDFs, or just chat with the bot, and it will intelligently process the content, create organized notes, and even stream responses back to you in real-time. It also features a web dashboard for monitoring and management.

## ✨ Features

-   **Web Dashboard**: A beautiful and responsive web interface for monitoring bot status, managing AI providers, and viewing system information, built with Go, Templ, and HTMX.
-   **AI-Powered Content Analysis**: Uses Google's Gemini, Groq and/or Hugging Face models (configurable via `config.yml`) to summarize text, answer questions, and categorize content.
-   **Multi-Provider AI Support**: Seamlessly switch between Google Gemini, Groq, and Hugging Face AI providers via the web dashboard or bot commands.
-   **Streaming Responses**: Get real-time answers from the AI, just like a modern chatbot.
-   **Chatbot Mode**: Chat directly with the bot for quick questions and answers; any non-command text is treated as an AI prompt.
-   **Multi-Language Support**: AI responses can be configured to default to any language on-the-fly with the `/lang` command.
-   **File Processing**: Extracts text from images (via Tesseract OCR) and PDFs (via Poppler).
-   **Intelligent Categorization**: Automatically categorizes content based on keywords defined in `config.yml`.
-   **Duplicate Detection**: Prevents processing the same file twice by checking its hash.
-   **Interactive Commands**: A rich set of slash commands, including `/setprovider`, `/reprocess`, `/modelinfo`, `/lang`, and `/stats`.
-   **Dockerized**: Easy to set up and run in a lightweight, single-stage Docker container.
-   **Database-Driven**: Uses a Turso database for persistent state (like chat history), with `sqlc` for generating type-safe Go code.

## 🚀 Getting Started

### Prerequisites

-   Docker
-   `make`
-   A Turso database instance (URL and auth token).

### 1. Configuration

Create a `.env` file in the root of the project and add your credentials:

```dotenv
# Your Telegram Bot Token (Required)
TELEGRAM_BOT_TOKEN=your-token-goes-here

# Turso Database URL and Auth Token (Required)
TURSO_DATABASE_URL=your-turso-database-url
TURSO_AUTH_TOKEN=your-turso-auth-token

# At least one AI provider key is recommended
# Comma-separated list of your Gemini API Keys
GEMINI_API_KEYS=key-1,key-2,key-3
# Your Groq API Key
GROQ_API_KEY=your-groq-api-key
# Your Hugging Face API Key
HUGGINGFACE_API_KEY=your-huggingface-api-key

# Port for the web dashboard (Optional, defaults to 8080)
DASHBOARD_PORT=8080
```

Create a `config.yml` file in the project root to define AI models and classification patterns:

```yaml
providers:
  gemini:
    model: gemini-1.5-pro-latest
  groq:
    model: llama-3.1-8b-instant

classification:
  patterns:
    physics: ["force", "energy", "mass"]
    math: ["equation", "function"]
    # ... more patterns ...

language_detection:
  french_words: ["le", "la", "de", "et"]
  # ... more language specific words ...
```

### 2. Running the Bot

Start the bot with a single command:

```sh
make up
```

Your bot is now running! You can view the dashboard at `http://localhost:8080` (or your configured port).

### 3. Available `make` Commands

-   `make up`: Build the Docker image and start the container.
-   `make down`: Stop and remove the container.
-   `make logs`: View the live logs of the running bot.
-   `make build`: Force a rebuild of the Docker image.
-   `make restart`: Restart the container.
-   `make help`: Show a list of all available commands.

## 🤖 Bot Commands

-   `/start`: Display a welcome message.
-   `/setprovider <provider>`: Switch between AI providers (e.g., `gemini`, `groq`).
-   `/reprocess`: Reprocess the last file you sent.
-   `/modelinfo`: Get details about the current AI model.
-   `/lang <language>`: Set the default language for AI responses.
-   `/stats`: View usage statistics.

Simply send an image, PDF, or text message to the bot to start processing.
