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
  - "Tesseract OCR"
  - "Pandoc"
  - "Tectonic"
entrypoint: "cmd/bot/main.go"
documentation:
  - "GEMINI.md"
  - "doc/README.md"
  - "doc/TODO.md"
---

# Project Context for AI

## 1. Project Purpose

This project is a Go-based Telegram bot that automates note-taking in Obsidian. It processes images and PDFs, extracts text, and uses AI to generate summaries, topics, and questions. It is designed to be a "digital assistant" for knowledge management.

## 2. Core Functionality

-   **Input**: Receives images, PDFs, and text messages from a Telegram user.
-   **Processing**:
    -   Extracts text from images using Tesseract OCR.
    -   Extracts text from PDFs using `pdftotext`.
    -   Uses AI (Gemini or Groq) to analyze the text, generate a summary, key topics, and review questions.
    -   Converts the generated note from Markdown to PDF.
-   **Output**:
    -   Sends a real-time stream of the AI-generated summary back to the user.
    -   Creates a new Markdown note in the user's Obsidian vault.
    -   Sends the final note to the user as a PDF.

## 3. Key Files and Directories

-   `cmd/bot/main.go`: The main application entrypoint.
-   `internal/bot/`: Contains the core bot logic, including message handlers, the main update loop, and file processing orchestration.
-   `internal/ai/`: Manages the AI providers (Gemini, Groq) and defines the `AIProvider` interface.
-   `internal/converter/`: Handles the conversion of Markdown to PDF.
-   `internal/database/`: Manages the database connection and schema.
-   `internal/health/`: Provides HTTP endpoints for health checks, pausing, and resuming the bot.
-   `internal/pid/`: Implements the PID lock mechanism to ensure single-instance operation.
-   `Dockerfile`: Defines the Docker image for the application.
-   `Makefile`: Contains commands for building, running, and managing the bot.
-   `.env`: Configuration file for API keys and other secrets (not checked into git).

## 4. Key Commands

-   `make up`: Builds the Docker image and starts the bot.
-   `make down`: Stops and removes the Docker container.
-   `make logs`: Tails the live logs from the running container.
-   `make build`: Forces a rebuild of the Docker image.
-   `make help`: Shows a list of all available `make` commands.

## 5. Detailed Data Flow

1.  **Message Reception**: The bot receives a message from a Telegram user. The message can be a command, a text message, an image, or a PDF.
2.  **Message Handling**: The `handleCommand`, `handlePhoto`, and `handleDocument` functions in `internal/bot/main.go` are called based on the message type.
3.  **File Download**: If the message contains a file, the file is downloaded and saved to the `./vault/attachments` directory.
4.  **Duplicate Check**: The SHA256 hash of the file is calculated, and the `IsDuplicate` function in `internal/bot/dedup.go` is called to check if the file has been processed before.
5.  **Text Extraction**:
    -   If the file is an image, Tesseract OCR is used to extract text.
    -   If the file is a PDF, `pdftotext` is used to extract text.
6.  **AI Processing**: The extracted text is sent to the AI service (`internal/ai/ai_service.go`). The active AI provider (Gemini or Groq) is used to generate a summary, key topics, and review questions.
7.  **Note Creation**: A new Markdown note is created in the `./vault/Inbox` directory. The note contains the AI-generated content, the extracted text, and metadata about the file.
8.  **PDF Conversion**: The Markdown note is converted to a PDF using Pandoc and Tectonic.
9.  **Response to User**: The AI-generated summary is streamed back to the user in real-time. The final note is sent to the user as a PDF.

## 6. Key Structs and Interfaces

-   `bot.Bot`: An interface that abstracts the Telegram Bot API, allowing for mock implementations in tests.
-   `bot.TelegramBot`: The production implementation of the `Bot` interface.
-   `bot.UserState`: Stores the state of a user, such as their language preference and the last processed file.
-   `ai.AIProvider`: An interface for AI services, allowing for multiple providers to be used interchangeably.
-   `ai.AIService`: Manages the available AI providers and selects the active one.
-   `ai.GeminiProvider`: Implements the `AIProvider` interface for the Google Gemini API.
-   `ai.GroqProvider`: Implements the `AIProvider` interface for the Groq API.
-   `database.DB`: The global database connection.

## 7. Error Handling Philosophy

-   Errors are generally handled by logging them to the console.
-   For user-facing errors (e.g., invalid command, file type not supported), a message is sent to the user.
-   For critical errors (e.g., failed to initialize the bot), the application will exit with a non-zero status code.
-   The AI service has built-in retry logic and key rotation for the Gemini API to handle `429` quota errors.

## 8. Configuration in Depth

The application is configured via a `.env` file in the project root.

-   `TELEGRAM_BOT_TOKEN` (Required): Your token from Telegram's BotFather.
-   `GEMINI_API_KEYS` (Required): A comma-separated list of your Gemini API keys. The bot will use these keys and rotate them automatically upon hitting quota limits.
-   `GROQ_API_KEY` (Required for Groq provider): Your Groq API key.
-   `OLLAMA_HOST` (Optional): The host for a local Ollama instance. The fallback to Ollama is not yet implemented.

## 9. Deployment

The bot is deployed as a Docker container. The `Dockerfile` defines the build process, which includes installing dependencies, compiling the Go application, and setting up the runtime environment. The `docker-compose.yml` file can be used to manage the container's lifecycle. The `make up` command builds and starts the container.

## 10. Codebase Structure (JSON)

```json
{
  "cmd/bot/main.go": {
    "functions": [
      {
        "name": "main",
        "description": "The main entrypoint of the application. It initializes the PID lock, handles signals for graceful shutdown, and starts the bot."
      }
    ]
  },
  "internal/bot/main.go": {
    "structs": [
      {
        "name": "UserState",
        "fields": ["Language", "LastProcessedFile", "LastCreatedNote"]
      },
      {
        "name": "TelegramBot",
        "embeds": ["*tgbotapi.BotAPI"]
      }
    ],
    "interfaces": [
      {
        "name": "Bot",
        "methods": ["Send", "Request", "GetFile"]
      }
    ],
    "functions": [
      {
        "name": "Run",
        "description": "Initializes the bot, sets up command handlers, and enters the main update loop."
      },
      {
        "name": "handleCommand",
        "description": "Handles incoming commands from Telegram users."
      },
      {
        "name": "handlePhoto",
        "description": "Handles incoming photos from Telegram users."
      },
      {
        "name": "handleDocument",
        "description": "Handles incoming documents from Telegram users."
      },
      {
        "name": "createObsidianNote",
        "description": "Orchestrates the creation of a new Obsidian note."
      }
    ]
  },
  "internal/ai/ai_service.go": {
    "structs": [
      {
        "name": "AIService",
        "fields": ["providers", "ActiveProvider", "mu"]
      }
    ],
    "functions": [
      {
        "name": "NewAIService",
        "description": "Initializes the AI service with provided providers."
      },
      {
        "name": "SetProvider",
        "description": "Changes the active AI provider."
      },
      {
        "name": "GetActiveProviderName",
        "description": "Returns the name of the currently active provider."
      },
      {
        "name": "GetActiveProvider",
        "description": "Returns the active AI provider."
      },
      {
        "name": "GetAvailableProviders",
        "description": "Returns a list of available provider names."
      },
      {
        "name": "GenerateContent",
        "description": "Delegates the call to the active provider."
      },
      {
        "name": "GenerateJSONData",
        "description": "Delegates the call to the active provider."
      }
    ]
  },
  "internal/ai/provider.go": {
    "interfaces": [
      {
        "name": "AIProvider",
        "methods": ["GenerateContent", "GenerateJSONData", "ProviderName"]
      }
    ]
  },
  "internal/health/health.go": {
    "functions": [
      {
        "name": "StartHealthServer",
        "description": "Starts the HTTP server for health and control endpoints."
      },
      {
        "name": "UpdateActivity",
        "description": "Updates the last activity timestamp."
      }
    ]
  },
  "internal/pid/pid.go": {
    "functions": [
      {
        "name": "CreatePIDFile",
        "description": "Creates a PID file to ensure only one instance of the bot is running."
      },
      {
        "name": "RemovePIDFile",
        "description": "Removes the PID file."
      }
    ]
  }
}
```
