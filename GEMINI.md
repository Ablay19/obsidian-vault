# GEMINI.md - Technical Documentation

This document provides a comprehensive technical overview of the AI-Powered Obsidian Automation Bot. It is intended for developers and contributors to the project.

## 1. Project Overview

The Obsidian Automation Bot is a Go-based Telegram bot designed to serve as a powerful, AI-enhanced assistant for note-taking and knowledge management with Obsidian. The bot can process images and PDFs, extract text, and use a variety of AI providers (including Google's Gemini and Groq) to summarize content, answer questions, and generate structured data. It offers an interactive, conversational experience directly within Telegram and WhatsApp, including real-time streaming of AI responses. It also features a web dashboard for monitoring and management.

The project is designed to be robust and resilient, featuring automatic API key rotation, intelligent model switching based on performance and cost, and a flexible configuration system. It is fully containerized with Docker for easy deployment and management.

## 2. Core Features

### AI & Content Processing
- **AI-Powered Analysis**: Leverages a variety of AI models (configurable via `config.yml`) for content summarization, question answering, and categorization.
- **Intelligent Model Switching**: Automatically selects the best AI provider (e.g., Gemini, Groq) for a given task based on configurable criteria such as cost, latency, and accuracy. This ensures optimal performance and cost-effectiveness.
- **Multi-Provider AI Support**: Seamlessly switch between AI providers, offering flexibility and optimized performance.
- **Streaming Responses**: AI-generated responses are streamed in real-time for an interactive, "live-typing" experience.
- **Multi-Language Support**: AI responses can be configured to default to any language on-the-fly.
- **OCR & Text Extraction**: Uses Tesseract for OCR on images and `pdftotext` for extracting content from PDF files.

### Bot & User Interaction
- **Multi-Platform Support**: Interact with the bot via Telegram and WhatsApp.
- **Web Dashboard**: A web-based interface for monitoring bot status, managing AI providers, and viewing system information.
- **Interactive Commands**: A rich set of slash commands for managing the bot and accessing features.
- **Command Autocompletion**: Registers its command list with Telegram, providing users with an easy-to-use command menu.

### Robustness & Configuration
- **Automatic Key Rotation**: Automatically detects API quota errors and switches to the next available key.
- **Database-Driven State**: Bot instance and chat history are managed in a Turso database.
- **Type-Safe Database Layer**: Uses `sqlc` to generate Go code from SQL queries, ensuring type safety.

## 3. Architecture

The bot's architecture is centered around a main Go application running inside a Docker container, backed by a Turso database.

```mermaid
graph TD
    subgraph "User"
        A[User on Telegram]
        B[User on WhatsApp]
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

    A -- "Text, Images, PDFs, Commands" --> C
    B -- "Webhook" --> C
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

## 4. Configuration

The application is configured via a `.env` file for secrets and a `config.yml` file for application-level settings. The `config.yml` defines AI providers, models, provider profiles, switching rules, and content classification patterns. High-level configuration management is handled by the `internal/config` package.

## 5. Codebase Deep Dive

- **`cmd/bot/main.go`**: Main entrypoint, initializes services.
- **`internal/bot`**: Core bot logic, command handlers, and platform integrations.
- **`internal/ai`**: AI service, provider implementations, and the intelligent selector.
- **`internal/config`**: Configuration loading and management.
- **`internal/database`**: Database connection, migrations, and queries.
- **`internal/dashboard`**: Web UI and API.
- **`internal/pipeline`**: The message processing pipeline.

## 7. Future Improvements
- **Performance Optimization**: Profile and optimize critical code paths.
- **Advanced Contact Management**: Implement robust contact synchronization and management.
- **Broader File Type Support**: Add support for more document types.
- **Enhanced UI/UX**: Improve the web dashboard and in-bot interactions.
- **Local AI Fallback**: Integrate with local AI models for offline processing.
