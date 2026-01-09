# FATBOT Project Development Plan

This document outlines the development plan for **Fatbot**, a Telegram-based task management bot, based on the provided specifications.

## 1. Project Overview

Fatbot will be a feature-rich, Go-based Telegram bot designed for efficient task management. It will provide users with a secure, interactive way to manage their tasks, set reminders, and receive notifications directly through Telegram. The project will be built with a modern, scalable, and maintainable architecture, containerized with Docker, and designed for robust, production-ready deployment.

## 2. Core Features

-   **User Authentication**: Secure user registration and login via Telegram, with session management using JSON Web Tokens (JWT).
-   **Task Management**: Full CRUD (Create, Read, Update, Delete) operations for tasks, including attributes like due dates, priorities (Low, Medium, High), and status (Pending, In-Progress, Completed).
-   **Automated Reminders**: A cron-based system to schedule and send reminders for tasks, with support for user-specific timezones.
-   **Real-time Notifications**: Proactive push notifications sent to users via Telegram for reminders and task status changes (e.g., "Due Soon," "Overdue").

## 3. Technical Architecture

The system will be designed as a stateless API with a modular structure.

-   **RESTful API**: The core logic will be exposed via a versioned RESTful API (`/api/v1`). This API will handle all business logic for users, tasks, and reminders.
-   **Telegram Bot**: The bot will be the primary user interface, interacting with the core API. It will handle user commands and present data from the API.
-   **PostgreSQL Database**: A PostgreSQL database will provide persistent storage for all application data, including users, tasks, and reminders.
-   **Reminder Service**: A background service, running as a goroutine, will use a cron scheduler to periodically check for reminders that need to be sent.
-   **Notification Service**: A dedicated service will be responsible for formatting and sending notifications to users via the Telegram Bot API.

```mermaid
graph TD
    subgraph "User"
        A[User on Telegram]
    end

    subgraph "Fatbot System (Docker Compose)"
        B[Fatbot Go Application]
        C[PostgreSQL Database]
        D[Redis (Optional Caching)]
    end

    subgraph "Fatbot Go Application"
        E[RESTful API Server]
        F[Telegram Bot Handler]
        G[Reminder Service (Cron)]
        H[Notification Service]
    end

    A -- "/start, /new_task, etc." --> F
    F -- "Calls" --> E
    E -- "Manages Data" --> C
    E -- "Caches Data" --> D
    
    G -- "Triggers" --> H
    H -- "Sends Message" --> F
    F -- "Pushes Notification" --> A

    E -- "Schedules Job" --> G
```

## 4. Technology Stack

-   **Language**: Go
-   **Database**: PostgreSQL (using `pgx` driver)
-   **API Framework**: Standard library `net/http` with a lightweight router (e.g., `gorilla/mux`).
-   **Containerization**: Docker & Docker Compose
-   **Configuration**: Viper
-   **Logging**: Zap
-   **Migrations**: `golang-migrate/migrate`
-   **Authentication**: JWT (`golang-jwt/jwt`)
-   **Scheduling**: `robfig/cron`
-   **Validation**: `go-playground/validator`
-   **Testing**: `testify` (assert, mock)
-   **Dependency Injection**: `uber-go/fx`
-   **CI/CD**: GitHub Actions

## 5. Development Plan

The project will be implemented in phases to ensure a structured and iterative development process.

### Phase 1: Project Scaffolding & Core API

-   **Objective**: Set up the foundational structure of the project.
-   **Tasks**:
    1.  Initialize Go module and directory structure (e.g., `/cmd`, `/internal`, `/pkg`).
    2.  Set up Docker Compose for the Go application and PostgreSQL database.
    3.  Implement database migrations for `users`, `tasks`, and `reminders` schemas using `golang-migrate`.
    4.  Create basic RESTful API endpoints for user and task CRUD operations.
    5.  Implement configuration management with Viper and structured logging with Zap.
    6.  Write initial unit tests for core data models and API handlers.

### Phase 2: Telegram Bot & Authentication

-   **Objective**: Integrate the Telegram bot and implement the user authentication flow.
-   **Tasks**:
    1.  Integrate the Telegram Bot API SDK.
    2.  Implement user registration via a `/start` command.
    3.  Implement JWT generation upon user registration/login.
    4.  Create API middleware to protect endpoints using JWT authentication.
    5.  Develop bot commands to interact with the task management API (e.g., `/new_task`, `/my_tasks`).

### Phase 3: Task Management & Reminders

-   **Objective**: Build out the complete task and reminder logic.
-   **Tasks**:
    1.  Implement full CRUD functionality for tasks via bot commands.
    2.  Integrate the `robfig/cron` library for reminder scheduling.
    3.  Create a background service that runs periodically to check for due reminders.
    4.  Implement logic to handle user-specific timezones for reminders.

### Phase 4: Notifications & Finalization

-   **Objective**: Implement the notification system and prepare for deployment.
-   **Tasks**:
    1.  Create a notification service to send reminders and alerts to users via Telegram.
    2.  Implement robust error handling and logging across the application.
    3.  Add input validation to all API endpoints.
    4.  Increase test coverage to meet the 80% target.

### Phase 5: CI/CD & Documentation

-   **Objective**: Automate the development lifecycle and create comprehensive documentation.
-   **Tasks**:
    1.  Set up a GitHub Actions pipeline for linting (`golangci-lint`), building, and testing on every push.
    2.  Implement a CD pipeline to (optionally) deploy the application on merge to the `main` branch.
    3.  Generate OpenAPI/Swagger documentation for the RESTful API.
    4.  Write a comprehensive `README.md` with setup and usage instructions.
    5.  Create a secure secrets management strategy for local (`.env`) and production environments.
