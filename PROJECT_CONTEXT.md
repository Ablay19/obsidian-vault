---
project_name: "AI-Powered Obsidian Automation Bot"
language: "Go"
containerization: "Docker"
primary_technologies:
  - "Go"
  - "Docker"
  - "Telegram Bot API"
  - "Google Gemini"
  - "Groq"
  - "Hugging Face"
  - "Tesseract OCR"
  - "Poppler utils"
  - "sqlc"
  - "Templ"
  - "HTMX"
entrypoint: "cmd/bot/main.go"
documentation:
  - "GEMINI.md"
  - "doc/README.md"
  - "doc/TODO.md"
---

# Project Context for AI

## 1. Project Purpose

This project is a Go-based Telegram bot that automates note-taking in Obsidian. It processes images and PDFs, extracts text using OCR, and leverages AI models (Gemini, Groq) to generate summaries and categorize content. It is designed to be a "digital assistant" for knowledge management, featuring a real-time web dashboard for monitoring.

## 2. Core Functionality

-   **Input**: Receives images, PDFs, and text messages from a Telegram user.
-   **Processing**:
    -   Extracts text from images using Tesseract OCR.
    -   Extracts text from PDFs using `pdftotext` (from Poppler utils).
        - Uses AI (Gemini, Groq or Hugging Face) to analyze the text, generate a summary, and categorize the content based on keywords.
-   **Output**:
    -   Streams the AI-generated summary back to the user in real-time.
    -   Creates a new Markdown note in the user's Obsidian vault.
    -   Provides a web dashboard to monitor bot status and manage AI providers.

## 3. Key Files and Directories

-   `cmd/bot/main.go`: The main application entrypoint.
-   `internal/bot/`: Contains the core bot logic, including message handlers and file processing orchestration.
-   `internal/ai/`: Manages the AI providers (Gemini, Groq) and the `AIProvider` interface.
-   `internal/database/`: Manages the Turso database connection, schema, and migrations. Uses `sqlc` to generate type-safe Go code from SQL queries.
-   `internal/dashboard/`: Contains the web dashboard implementation using `Templ` and `HTMX`.
    -   `dashboard.go`: The backend handler for the dashboard.
    -   `dashboard.templ`: The `Templ` file defining the UI components.
-   `Dockerfile`: A single-stage Dockerfile for building a minimal runtime image.
-   `Makefile`: Contains commands for building, running, and managing the bot container.
-   `config.yml`: Configuration for AI models and classification patterns.
-   `.env`: Configuration file for API keys and other secrets (not checked into git).

## 4. Key Commands

-   `make up`: Builds the Docker image and starts the bot and its dashboard.
-   `make down`: Stops and removes the Docker container.
-   `make logs`: Tails the live logs from the running container.
-   `make build`: Forces a rebuild of the Docker image.
-   `make help`: Shows a list of all available `make` commands.

## 5. Detailed Data Flow

1.  **Message Reception**: The bot's main loop in `internal/bot/main.go` receives an update from the Telegram API.
2.  **Message Handling**: The appropriate handler (`handleCommand`, `handlePhoto`, `handleDocument`) is called based on the message type.
3.  **File Download**: If the message contains a file, it's downloaded to `./vault/attachments`.
4.  **Duplicate Check**: The file's SHA256 hash is checked against the `processed_files` table in the database to prevent reprocessing.
5.  **Text Extraction**: Text is extracted using Tesseract (for images) or `pdftotext` (for PDFs).
6.  **AI Processing**: The extracted text is sent to the active AI provider (`internal/ai/ai_service.go`) to generate a summary and structured data (category, topics).
7.  **Note Creation**: A new Markdown note is created in the `./vault/Inbox` directory (or a categorized subdirectory).
8.  **Response to User**: The AI-generated summary is streamed back to the user in real-time.
9.  **Dashboard**: The web dashboard at `http://localhost:8080` continuously polls API endpoints in `internal/dashboard/dashboard.go` to display real-time status.

## 6. Key Structs and Interfaces

-   `bot.TelegramBot`: The production implementation of the bot, handling interactions with the Telegram API.
-   `ai.AIProvider`: An interface for AI services, allowing for multiple providers (Gemini, Groq) to be used interchangeably.
-   `ai.AIService`: Manages the available AI providers and allows switching the active one.
-   `ai.GeminiProvider`: Implements the `AIProvider` interface for the Google Gemini API.
-   `ai.GroqProvider`: Implements the `AIProvider` interface for the Groq API.
-   `ai.HuggingFaceProvider`: Implements the `AIProvider` interface for the Hugging Face API.
-   `dashboard.Dashboard`: Holds dependencies for the web dashboard and registers its HTTP handlers.

## 7. Error Handling Philosophy

-   Errors are logged to the console for debugging.
-   User-facing errors (e.g., invalid command) result in a message sent to the user.
-   Critical errors on startup (e.g., failed to connect to the database) will cause the application to exit.
-   The Gemini provider includes automatic API key rotation to handle `429` quota errors.

## 8. Configuration in Depth

The application is configured via a `.env` file and `config.yml`.

**`.env` file:**
-   `TELEGRAM_BOT_TOKEN` (Required): Your Telegram bot token.
-   `TURSO_DATABASE_URL` (Required): URL for your Turso database.
-   `TURSO_AUTH_TOKEN` (Required): Auth token for your Turso database.
-   `GEMINI_API_KEYS` (Optional): Comma-separated list of Gemini API keys.
-   `GROQ_API_KEY` (Optional): Your Groq API key.
-   `HUGGINGFACE_API_KEY` (Optional): Your Hugging Face API key.
-   `DASHBOARD_PORT` (Optional): Port for the web dashboard (defaults to 8080).

**`config.yml` file:**
- Defines the specific AI models to use for each provider and the keywords for content classification.

## 9. Deployment

The bot is deployed as a Docker container. The simplified `Dockerfile` creates a small, single-stage image. The `make up` command is the standard way to build and run the container, reading configuration from the `.env` file.