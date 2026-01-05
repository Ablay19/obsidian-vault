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
  - "OpenRouter"
  - "Tesseract OCR"
  - "Poppler utils"
  - "sqlc"
  - "Templ"
  - "HTMX"
  - "Alpine.js"
  - "Bubbletea"
entrypoint: "cmd/bot/main.go"
documentation:
  - "GEMINI.md"
  - "doc/README.md"
  - "doc/TODO.md"
---

# Project Context for AI

## 1. Project Purpose

This project is a Go-based automation system that bridges the gap between messaging apps (Telegram, WhatsApp) and Obsidian. It acts as an intelligent ingestion engine that processes documents, extracts knowledge using OCR and multi-provider AI, and maintains a synchronized, version-controlled vault.

## 2. Core Functionality

-   **Multi-Source Ingestion**: Receives content from Telegram, Google Drive (stub), and WhatsApp (planned).
-   **Asynchronous Pipeline**: Uses a worker pool to process files in the background, improving scalability and responsiveness.
-   **AI Knowledge Extraction**:
    *   Extracts text using Tesseract (images) and Poppler (PDFs).
    *   Uses AI (Gemini, Groq, HF, OpenRouter) to generate summaries and answer questions within documents.
    *   **Failover**: Automatically rotates through providers (Round-Robin) on transient failures (429, 5xx).
-   **Security**: Implements AES-256-GCM encryption for API keys at rest and redacts sensitive data in the UI.
-   **Output & Synchronization**:
    *   Creates Markdown notes and high-fidelity PDFs (default).
    *   **Git Sync**: Automatically commits and pushes new notes to a Git repository.
    *   **Real-time Dashboard**: HTMX-powered SPA for management, featuring real-time chat history and provider control.
-   **Interactive CLI**: A `bubbletea`-based terminal user interface for managing the bot, including views for bot status, AI providers, and user management.

## 3. Key Files and Directories

-   `internal/pipeline/`: The core ingestion engine and worker pool logic.
-   `internal/git/`: Robust Git automation manager.
-   `internal/security/`: Encryption, validation, and rate limiting utilities.
-   `internal/bot/`: Telegram specific logic and pipeline adapters.
-   `internal/ai/`: Multi-provider AI service with failover logic.
-   `internal/dashboard/`: UI components using Templ, HTMX, and Alpine.js.
-   `cmd/cli/tui/`: The `bubbletea`-based TUI for the CLI.

## 4. Roadmap & Next Steps

-   **WhatsApp Integration**: Implement a production-ready connector for WhatsApp using Meta Cloud API or a local gateway.
-   **Bidirectional Sync**: Enhance the account linking protocol to allow dashboard-to-telegram command triggers.
-   **Advanced Search**: Add vector-based search for the Obsidian vault within the dashboard.
-   **Local AI Fallback**: Integrate Ollama for offline processing.

## 4. Key Commands

-   `make up`: Builds the Docker image and starts the bot and its dashboard.
-   `make down`: Stops and removes the Docker container.
-   `make logs`: Tails the live logs from the running container.
-   `make build`: Forces a rebuild of the Docker image.
-   `make help`: Shows a list of all available `make` commands.
-   `go run ./cmd/cli tui`: Starts the interactive CLI.
-   `go run ./cmd/cli add-user <username>`: Adds a new SSH user.

## 5. Detailed Data Flow

1.  **Service Startup**: The main application in `cmd/bot/main.go` starts all the services:
    *   Web dashboard and API server.
    *   SSH server for secure remote access.
    *   Telegram bot listener.
2.  **Message Reception**: The bot's main loop receives an update from the Telegram API.
3.  **Message Handling**: The appropriate handler (`handleCommand`, `handlePhoto`, `handleDocument`) is called based on the message type.
4.  **File Download**: If the message contains a file, it's downloaded to `./vault/attachments`.
5.  **Duplicate Check**: The file's SHA256 hash is checked against the `processed_files` table in the database to prevent reprocessing.
6.  **Text Extraction**: Text is extracted using Tesseract (for images) or `pdftotext` (for PDFs).
7.  **AI Processing**: The extracted text is sent to the active AI provider (`internal/ai/ai_service.go`) to generate a summary and structured data (category, topics).
8.  **Note Creation**: A new Markdown note is created in the `./vault/Inbox` directory (or a categorized subdirectory).
9.  **Response to User**: The AI-generated summary is streamed back to the user in real-time.
10. **Dashboard**: The web dashboard at `http://localhost:8080` continuously polls API endpoints in `internal/dashboard/dashboard.go` to display real-time status.

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
-   **Instance Management**: A heartbeat-based mechanism is used to detect and clean up stale bot instances, preventing deadlocks during restarts in containerized environments.

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