# GEMINI.md - AI-Powered Obsidian Automation Bot

This document provides a technical overview of the AI-Powered Obsidian Automation Bot. It is intended for developers and contributors to the project.

## 1. Project Overview

The Obsidian Automation Bot is a Go-based platform designed to be an intelligent assistant for knowledge management and automation. It integrates with various AI providers (Google Gemini, Groq, etc.) to process and analyze content from different sources, including text, images, and PDFs. The bot is accessible through multiple channels like Telegram, WhatsApp, and a web dashboard.

The project is containerized using Docker for easy deployment and management. It includes services for the main bot application, a Redis cache, an optional SSH server for secure access, and an optional HashiCorp Vault for secrets management.

## 2. Key Technologies

*   **Backend**: Go
*   **Web Framework**: Gin
*   **Database**: SQLite with SQLc for type-safe queries. Turso is also used.
*   **AI**: Google Gemini, Groq, LangchainGo
*   **Communication**: Telegram Bot API, WhatsApp, WebSockets
*   **CLI**: Bubbletea, Cobra
*   **Containerization**: Docker
*   **CI/CD**: GitHub Actions
*   **Frontend**: HTML/CSS/JavaScript with Alpine.js

## 3. Building and Running

The project uses a `Makefile` for common development tasks.

### Docker Development

The recommended way to run the project is by using Docker Compose.

*   **Start all services**:
    ```bash
    docker-compose up -d
    ```
*   **Stop all services**:
    ```bash
    docker-compose down
    ```
*   **View logs**:
    ```bash
    docker-compose logs -f
    ```

### Local Development (without Docker)

You can also run the application locally using the `Makefile`.

*   **Run the application**:
    ```bash
    make run
    ```
*   **Run in development mode (with hot-reloading)**:
    ```bash
    make dev
    ```
*   **Build the binary**:
    ```bash
    make build
    ```
*   **Run tests**:
    ```bash
    make test
    ```
*   **Run tests with coverage**:
    ```bash
    make test-coverage
    ```

## 4. Configuration

The application is configured through a combination of a `.env` file and a `config.yaml` file.

*   **`.env`**: This file is used for secrets and environment-specific variables. A `.env.example` file is provided as a template. Key variables include `TELEGRAM_BOT_TOKEN`, `TURSO_DATABASE_URL`, `TURSO_AUTH_TOKEN`, and API keys for AI providers.
*   **`config.yaml`**: This file contains high-level application settings, including:
    *   AI provider configurations and model selection.
    *   Rules for switching between AI providers.
    *   WhatsApp integration settings.
    *   Web dashboard and authentication settings.

## 5. Development Conventions

*   **Linting**: The project uses `gofmt` and `go vet` for code formatting and analysis. Run the linter with:
    ```bash
    make lint
    ```
*   **SQL Generation**: The project uses `sqlc` to generate type-safe Go code from SQL queries. To regenerate the code after changing SQL files in the `database/queries` directory, run:
    ```bash
    make sqlc-generate
    ```
*   **CI/CD**: GitHub Actions are used for continuous integration, linting, and deployment. The workflows are defined in the `.github/workflows` directory.
