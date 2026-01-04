# GEMINI.md - Technical Documentation

This document provides a comprehensive technical overview of the AI-Powered Obsidian Automation Bot. It is intended for developers and contributors to the project.

## 1. Project Overview

The Obsidian Automation Bot is a Go-based Telegram bot designed to serve as a powerful, AI-enhanced assistant for note-taking and
knowledge management with Obsidian. The bot can process images and PDFs, extract text, and use Google's Gemini AI to summarize
content, answer questions, and generate structured data. It offers an interactive, conversational experience directly within
Telegram, including real-time streaming of AI responses. It also features a web dashboard for monitoring and management.

The project is designed to be robust and resilient, featuring automatic API key rotation for the Gemini service and a flexible
configuration system. It is fully containerized with Docker for easy deployment and management.

## 2. Core Features

### AI & Content Processing

-   **AI-Powered Analysis**: Leverages Google's Gemini models (configurable via `config.yml`) and **Groq** for content summarization, question answering, and categorization.
-   **Intelligent Model Switching**: Automatically selects the best AI provider (e.g., Gemini, Groq) for a given task based on configurable criteria such as cost, latency, and accuracy. This ensures optimal performance and cost-effectiveness.
-   **Multi-Provider AI Support**: Seamlessly switch between Google Gemini and Groq AI providers, offering flexibility and optimized performance. The active provider can be changed via bot command or the web dashboard.
-   **Streaming Responses**: AI-generated responses are streamed in real-time to the user for an interactive, "live-typing" experience.
-   **Multi-Language Support**: AI responses can be configured to default to any language on-the-fly via
    the `/lang` command. This flexibility is enabled by the underlying AI models' multilingual capabilities.
-   **Question Answering**: The AI is prompted to answer any questions it finds within the text of a provided document.
-   **OCR & Text Extraction**: Uses Tesseract for OCR on images and `pdftotext` for extracting content from PDF files.
-   **Improved PDF Conversion**: Utilizes a headless Chrome instance to convert Markdown notes to PDF, ensuring high-fidelity rendering of complex notes, including those with LaTeX.
-   **Reliable AI Interaction**: Implements a two-call AI strategy. The first call obtains a human-readable summary, streamed
    in real-time. The second, separate call, is used to extract structured JSON data (e.g., category, topics). This separation
    significantly reduces the risk of JSON parsing errors, as the AI can focus on generating valid JSON without interference
    from streaming or conversational elements.

### Bot & User Interaction

-   **Web Dashboard**: A web-based interface for monitoring bot status, managing AI providers, and viewing system information (accessible on `DASHBOARD_PORT`, default 8080).
-   **Chatbot Mode**: Functions as a general-purpose chatbot, treating any non-command text message as a prompt for the AI.
-   **Interactive Commands**: A rich set of slash commands for managing the bot and accessing features, including:
    *   `/setprovider`: To switch AI models.
    *   `/last`: To show the last note created.
    *   `/reprocess`: To reprocess the last sent file.
    *   `/pid`: To get the process ID of the bot.
    *   `/modelinfo`: To display information about the configured AI models.
    *   `/lang`: To set the AI response language.
    *   `/stats`: To view bot usage statistics.
    *   `/service_status`: To check the status of core bot services.
    *   `/pause_bot` and `/resume_bot`: To control bot processing.
-   **Command Autocompletion**: Registers its command list with Telegram, providing users with an easy-to-use command menu.
-   **"Typing" Indicator**: Provides real-time user feedback by displaying the "typing..." status while processing requests.

### Robustness & Configuration

-   **Multi-API Key Support**: Manages a list of Gemini API keys from a single environment variable (`GEMINI_API_KEYS`).
-   **Automatic Key Rotation**: Automatically detects `429` quota errors from the Gemini API and switches to the next available
    key, ensuring high availability.
-   **Manual Key Switching**: Allows manual rotation of API keys via the `/switchkey` command (Note: This command might be deprecated in favor of `/setprovider`).
-   **Resilient Initialization**: The AI service can be disabled without crashing the bot if no API keys are provided.
-   **Database-Driven State**: Bot instance and chat history are managed in a Turso database, replacing previous file-based PID locks.
-   **Type-Safe Database Layer**: Uses `sqlc` to generate Go code from SQL queries, ensuring type safety and reducing runtime errors.

## 3. Architecture

The bot's architecture is centered around a main Go application running inside a Docker container, backed by a Turso database.

```mermaid
graph TD
    subgraph "User"
        A[User on Telegram]
        UDB[User on Web Browser (Dashboard)]
    end

    subgraph "Go Application (Docker)"
        C[cmd/bot/main.go: Entrypoint]
        D[internal/bot: Core Bot Logic]
        E[internal/ai: AI Service]
        F[internal/converter: File Conversion]
        G[internal/database: Database (Turso client)]
        H[internal/health: Health & Control]
        Dash[internal/dashboard: Web UI & API]
        Stat[internal/status: Shared Bot Status]
    end

    subgraph "AI Providers"
        I[Google Gemini API]
        J[Groq API]
    end

    subgraph "Database"
        K[Turso Database]
    end

    A -- "Text, Images, PDFs, Commands" --> B(Telegram Bot API)
    B -- "Streamed AI Responses, Messages" --> A
    B -- "Webhook/Long-polling Updates" --> C
    C -- "Starts Bot & Dashboard" --> D
    C -- "Starts Bot & Dashboard" --> Dash
    D -- "Uses" --> E
    D -- "Uses" --> F
    D -- "Uses" --> G
    D -- "Updates" --> Stat
    Dash -- "Serves HTML/CSS/JS" --> UDB
    UDB -- "API Requests" --> Dash
    Dash -- "Uses" --> E
    Dash -- "Reads" --> Stat
    E -- "API Requests" --> I
    E -- "API Requests" --> J
    G -- "SQL Queries" --> K
    D -- "SQL Queries" --> K
    Stat -- "Bot Status/Activity" --> H(Health Endpoint)
    H -- "Reads" --> Stat
```

_Note: This Mermaid diagram replaces the previous text-based version for better clarity and maintainability._

## 4. Configuration

The application is configured via a `.env` file in the project root and a `config.yml` file.

**`.env` file variables:**
-   `TELEGRAM_BOT_TOKEN` (Required): Your token from Telegram's BotFather.
-   `GEMINI_API_KEYS` (Required): A **comma-separated list** of your Gemini API keys. Do not include spaces between the keys.
    The bot will use these keys and rotate them automatically upon hitting quota limits.
-   `GROQ_API_KEY` (Required for Groq provider): Your Groq API key.
-   `DASHBOARD_PORT` (Optional): The port for the web dashboard. Defaults to `8080`.

**Example `.env` file:**
```dotenv
TELEGRAM_BOT_TOKEN=12345:your-long-telegram-token
GEMINI_API_KEYS=key-one,key-two,key-three
GROQ_API_KEY=your-groq-api-key
DASHBOARD_PORT=8080
```

**`config.yml` file:**
This file (located in the project root) defines the AI models to use for each provider, as well as classification patterns and language detection settings.

**Example `config.yml` structure:**
```yaml
providers:
  gemini:
    model: gemini-1.5-pro-latest
  groq:
    model: llama-3.1-8b-instant

provider_profiles:
  gemini:
    provider_name: "gemini"
    model_name: "gemini-1.5-pro-latest"
    input_cost_per_token: 0.000007
    output_cost_per_token: 0.000021
    max_input_tokens: 8192
    max_output_tokens: 2048
    latency_ms_threshold: 3000
    accuracy_pct_threshold: 0.95
  groq:
    provider_name: "groq"
    model_name: "llama-3.1-8b-instant"
    input_cost_per_token: 0.00000005
    output_cost_per_token: 0.0000001
    max_input_tokens: 4096
    max_output_tokens: 1024
    latency_ms_threshold: 500
    accuracy_pct_threshold: 0.90

switching_rules:
  default_provider: "gemini"
  latency_target: 1000
  accuracy_threshold: 0.92
  retry_count: 3
  retry_delay_ms: 1000
  on_error: "switch_provider"

classification:
  patterns:
    physics:
      - force
      - energy
      - mass
      - velocity
      - acceleration
    math:
      - equation
      - function
      - derivative
      - integral
      - matrix
    chemistry:
      - molecule
      - atom
      - reaction
      - chemical
    admin:
      - invoice
      - contract
      - form
      - certificate
    general: # Default fallback category
      - general
      - miscellaneous

language_detection:
  french_words: # Used for basic French language detection
    - le
    - la
    - de
    - et
    - un
```

## 5. Development Guide

The project uses a `Makefile` to simplify common development tasks.

-   **`make up`**: Builds the Docker image and starts the container. This is the main command to get the bot running.
-   **`make down`**: Stops and removes the Docker container.
-   **`make logs`**: Tails the live logs from the running container.
-   **`make build`**: Forces a rebuild of the Docker image.
-   **`make help`**: Shows a list of all available `make` commands.

### Documentation Standards

To maintain high quality and consistency, `GEMINI.md` and other Markdown documentation files in this project are subject
to linting using `markdownlint`. A configuration file (`.markdownlint.json`) in the project root defines the specific
rules and styles enforced. It is recommended to run `markdownlint` locally before committing changes, and it is integrated
into the project's CI/CD pipeline to ensure adherence to standards.

## 6. Codebase Deep Dive

### `cmd/bot/main.go`

-   **`main()`**: The main entrypoint of the application. It initializes the database, applies schema and migrations, starts the AI service, sets up the dashboard, and handles graceful shutdown signals. The PID lock mechanism is currently disabled but was previously located here.

### `internal/bot`

-   This package contains the core bot logic.
-   **`main.go`**: Contains the main `Run` function that initializes the Telegram bot, sets up command handlers, and enters the main update loop.
-   **`handler.go`**: (This file might not exist or its role might be merged into `main.go` in recent refactors. **Verify if this file exists and its current role.**) Originally intended to contain `handleCommand`, `handlePhoto`, and `handleDocument` functions.
-   **`processor.go`**: Contains the `processFileWithAI` function, which orchestrates text extraction, AI processing, and note creation. Also includes content classification and language detection logic, now configurable via `config.yml`.
-   **`dedup.go`**: Contains the `IsDuplicate` function for checking for duplicate files.
-   **`stats.go`**: (This file might be deprecated or its logic moved to `internal/status`. **Verify if this file exists and its current role.**) Originally contained logic for tracking usage statistics.

### `internal/ai`

-   **`ai_service.go`**: Contains the `AIService` struct, which manages multiple AI providers (Gemini, Groq) and handles switching between them.
-   **`selector.go`**: Implements the `select_provider` function, which contains the core logic for dynamically selecting an AI provider.
-   **`provider.go`**: Defines the `AIProvider` interface and `ModelInfo` struct.
-   **`gemini_provider.go`**: Implements the `AIProvider` interface for the Google Gemini API, using models specified in `config.yml`.
-   **`groq_provider.go`**: Implements the `AIProvider` interface for the Groq API, using models specified in `config.yml`.
-   **`mock_provider.go`**: This file has been removed.

### `internal/config`

-   **`config.go`**: Contains the `Config` struct and `LoadConfig` function for loading application settings from `config.yml` and environment variables.

### `internal/converter`

-   **`converter.go`**: Contains the `ConvertMarkdownToPDF` function for converting Markdown to PDF using pandoc and tectonic.

### `internal/dashboard`

-   **`dashboard.go`**: Contains the `Dashboard` struct and `RegisterRoutes` function for serving the web UI and handling API requests related to bot status and AI provider management.
-   **`static/`**: Contains the `index.html`, `style.css`, and `script.js` for the web dashboard UI.

### `internal/database`

-   **`db.go`**: Contains the database connection logic (for Turso) and the `ApplySchemaAndMigrations` function for in-app database initialization and migration.
-   **`schema.sql`**: Defines the database schema, including `processed_files` and `instances` tables.
-   **`migrations/`**: Contains SQL migration files, such as `001_create_chat_history.sql`, which are applied on startup.
-   **`instance.go`**: Contains logic for managing bot instance PIDs in the database (currently disabled in `main.go`).
-   **`sqlc/`**: (This directory exists, but is not explicitly mentioned in the old documentation. **Add a description for its role if needed.**) Contains generated Go code for type-safe database interactions.

### `internal/health`

-   **`health.go`**: Contains the `StartHealthServer` function (now integrated into the main router) which provides health check functionality.

### `internal/pid`

-   **`pid.go`**: This package and its file-based PID lock mechanism have been deprecated/removed in favor of a database-driven approach (which is currently disabled for multi-instance compatibility).

### `internal/status`

-   **`get.go`**: Contains logic for tracking and reporting the bot's runtime status (e.g., paused state, uptime, last activity) and basic statistics, used by both the bot and the dashboard.

## 7. Future Improvements

This section outlines potential enhancements and new features for the bot. For a detailed list of specific tasks and their
current status, please refer to the `TODO.md` file.

Key areas for future development include:
-   **Local AI Fallback**: Completing the integration with Ollama for local AI processing.
-   **User Customization**: Implementing per-user settings for features like language preferences.
-   **More Advanced Commands**: Adding functionality like `/search` or `/config` commands.
-   **Observability**: Adding a web dashboard for monitoring bot statistics and health.
-   **Broader File Type Support**: Expanding the range of document types the bot can process (e.g., .docx, .txt).

## 8. Bot Instance Management

### Database-Driven Instance Lock (Currently Disabled for Multi-Instance Docker)
The bot previously implemented a database-driven instance lock to prevent multiple instances from processing Telegram updates simultaneously, which causes "Conflict: terminated by other getUpdates request" errors from the Telegram API.

**How it worked (when enabled):**
1.  On startup, the bot would attempt to record its process ID (PID) in a dedicated `instances` table in the database.
2.  If an entry already existed and the associated process was still running, the new instance would exit, preventing conflicts.
3.  On graceful shutdown, the PID entry was removed.

**Current Status and Multi-Instance Deployment:**
*   For compatibility with Docker environments where each container runs as PID 1 and to allow for potentially scaled deployments, the explicit `CheckExistingInstance` and `AddInstance` logic has been **disabled** in `cmd/bot/main.go`.
*   This means the bot no longer self-regulates against multiple instances via the database.
*   **Important:** Running multiple bot instances that *all use Telegram's long-polling `getUpdates` method concurrently will still result in Telegram API conflicts ("Conflict: terminated by other getUpdates request").**
*   **To run multiple bot instances effectively**, you must implement an external strategy to manage Telegram updates, such as:
    *   **Webhooks:** Configure Telegram to send updates to a single webhook URL, which then distributes them to your bot instances behind a load balancer. This is the recommended approach for scaling.
    *   **Single Active Instance:** Ensure only one Docker container/bot instance is actively running and polling the Telegram API at any given time.

**Files:**
-   `internal/database/instance.go`: Contains the database interaction logic for managing instance PIDs.
-   `cmd/bot/main.go`: Where the instance management logic was previously called (now disabled).

### Observability

The bot provides a web dashboard and API endpoints for monitoring and control:

-   **Web Dashboard**: Accessible on `http://localhost:8080` (or configured `DASHBOARD_PORT`). Provides a UI for:
    *   Viewing bot status (online/offline, uptime, last activity).
    *   Viewing system information (OS, Architecture, Go Version, PID).
    *   Managing AI providers (view available, set active).
    *   Viewing status of core bot services.
-   `/api/services/status`: Returns JSON data with the current status of bot core and other services.
-   `/api/ai/providers`: Returns JSON data with available and active AI providers.
-   `/api/ai/provider/set` (POST): Allows setting the active AI provider.
-   `/health`: (Note: This specific endpoint might be superseded by the `/api/services/status` providing more granular detail).

The `internal/dashboard` package handles the web UI and API endpoints, while the `internal/status` package tracks the bot's runtime state.