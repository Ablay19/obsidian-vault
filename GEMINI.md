# GEMINI.md

## Project Overview

This project is a Go-based Telegram bot designed to automate the process of adding notes to an Obsidian vault. The bot listens for images and PDF files sent to it, processes them to extract text and other relevant information, and then creates new, organized notes in the vault. The project is containerized using Docker for easy deployment and management.

The bot uses Tesseract for OCR and pdftotext for PDF text extraction. It also integrates with Google's Gemini for more advanced content processing, including classification, summarization, and generating key topics and review questions.

## Building and Running

The project includes a set of shell scripts for common tasks:

*   **Quick Start:** `./quick-start.sh`
    *   This script is the main entry point for starting the bot. It likely builds the Go application and starts the Docker container.
*   **Update:** `./update.sh`
    *   This script should be run after making any changes to the Go source code. It rebuilds the application and restarts the bot.
*   **Status:** `./status.sh`
    *   Checks the status of the bot.
*   **Logs:** `./logs.sh`
    *   Displays the logs from the bot.
*   **Restart:** `./restart.sh`
    *   Restarts the bot.
*   **Stop:** `./stop-bot.sh`
    *   Stops the bot.
*   **Backup:** `./backup-vault.sh`
    *   Backs up the Obsidian vault to a Git repository.

### Prerequisites

*   Go
*   Docker
*   Tesseract OCR
*   pdftotext

### Configuration

Before running the bot, you need to create a `.env` file with your Telegram bot token:

```
TELEGRAM_BOT_TOKEN=your-token-goes-here
```

## Development Conventions

*   **Code Style:** The code follows standard Go formatting and conventions.
*   **Dependencies:** Go modules are used for dependency management, with dependencies defined in the `go.mod` file.
*   **Testing:** There are no explicit tests in the provided file list.
*   **Contribution:** The project is structured into several files, each with a specific responsibility:
    *   `main.go`: Handles the main bot logic and message processing.
    *   `processor.go`:  Handles file processing, including text extraction and classification.
    *   `organizer.go`: Organizes notes into subdirectories.
    *   `stats.go`: Tracks statistics about the bot's usage.
    *   `health.go`: Provides a health check endpoint.
    *   `dedup.go`: Handles duplicate file detection.
    *   `ai_gemini.go`:  Contains the logic for interacting with the Gemini API.
*   **Error Handling:** Errors are generally handled by logging them to the console.
