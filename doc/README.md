# AI-Powered Obsidian Automation Bot

A powerful, AI-enhanced Telegram bot to automate your note-taking workflow with Obsidian. Send images, PDFs, or just chat with the bot, and it will intelligently process the content, create organized notes, and even stream responses back to you in real-time. It also features a web dashboard for monitoring and management.

## ✨ Features

-   **Web Dashboard**: A web-based interface for monitoring bot status, managing AI providers, and viewing system information.
-   **AI-Powered Content Analysis**: Uses Google's Gemini and **Groq** models (configurable via `config.yml`) to summarize text, answer questions, and categorize content.
-   **Multi-Provider AI Support**: Seamlessly switch between Google Gemini and Groq AI providers, offering flexibility and optimized performance. The active provider can be changed via bot command or the web dashboard.
-   **Streaming Responses**: Get real-time answers from the AI, just like a modern chatbot.
-   **Chatbot Mode**: Chat directly with the bot for quick questions and answers, processing any non-command text as an AI prompt.
-   **Multi-Language Support**: AI responses can be configured to default to any language on-the-fly with the `/lang` command.
-   **File Processing**: Extracts text from images (via Tesseract OCR) and PDFs.
-   **Improved PDF Conversion**: The bot uses `pandoc` and `tectonic` to convert Markdown notes to PDF, ensuring high-fidelity rendering of complex notes, including those with LaTeX.
-   **Intelligent Categorization**: Automatically categorizes content based on patterns defined in `config.yml`.
-   **Duplicate Detection**: Prevents processing the same file twice.
-   **Interactive Commands**: Manage the bot and your notes with a rich set of slash commands, including `/pid`, `/reprocess`, `/modelinfo`, `/lang`, `/setprovider`, `/stats`, and `/last`.
-   **Dockerized**: Easy to set up and run in a containerized environment.
-   **Database-Driven**: Uses a Turso database for persistent state and chat history, with `sqlc` for type-safe database interactions.

## 🚀 Getting Started

### Prerequisites

-   Docker
-   `make`
-   Go (for development)
-   A Turso database instance URL and auth token.

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

# Turso Database URL and Auth Token (Required for database features)
TURSO_DATABASE_URL=your-turso-database-url
TURSO_AUTH_TOKEN=your-turso-auth-token

# Port for the web dashboard (Optional, defaults to 8080)
DASHBOARD_PORT=8080
```

Also, create a `config.yml` file in the project root to define AI models and classification patterns:

```yaml
providers:
  gemini:
    model: gemini-1.5-pro-latest # Or other Gemini models
  groq:
    model: llama-3.1-8b-instant # Or other Groq models

classification:
  patterns:
    physics: ["force", "energy", "mass"]
    math: ["equation", "function"]
  # ... more patterns

language_detection:
  french_words: ["le", "la", "de"]
  # ... more language specific words
```