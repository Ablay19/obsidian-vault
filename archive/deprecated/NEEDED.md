# Needed Improvements and Future Enhancements

This document outlines key areas for improvement, refactoring, and future development within the AI-Powered Obsidian Automation Bot project, compiled from code reviews, project context, and ongoing development tasks.

## 1. Codebase Refactoring & Quality

### 1.1 Bot Command Logic Refactor
-   **Description:** The `handleCommand` function in `internal/bot/main.go` exhibits tight coupling with state management.
-   **Action:** Extract command handling logic into a separate state machine or a more modular handler pattern to improve testability and extensibility.

### 1.2 Comprehensive Unit Testing
-   **Description:** Critical core logic in `internal/bot/processor.go` (file processing) and `internal/ai/ai_service.go` (AI orchestration) lacks sufficient unit tests.
-   **Action:** Prioritize writing robust unit tests for these components, utilizing mocks for external dependencies like AI providers, database interactions, and file system operations.

### 1.3 External Binary Dependency Check
-   **Description:** The bot implicitly relies on external binaries (`tesseract` for OCR, `pdftotext` for PDF extraction). Failures to find these binaries lead to runtime errors.
-   **Action:** Implement a startup check for these required external binaries. If missing, log clear, actionable error messages and potentially provide guidance for installation or offer a graceful fallback mechanism.

## 2. Feature Enhancements & Roadmap Items

### 2.1 WhatsApp Integration
-   **Description:** Implement a production-ready connector for WhatsApp, potentially utilizing Meta Cloud API or a local gateway, to expand multi-platform support.

### 2.2 Bidirectional Synchronization
-   **Description:** Enhance the account linking protocol to enable command triggers from the web dashboard to Telegram or WhatsApp. This would allow for more integrated control and interaction.

### 2.3 Advanced Obsidian Vault Search
-   **Description:** Integrate vector-based search capabilities for the Obsidian vault within the web dashboard, providing more powerful and contextual search results.

### 2.4 Local AI Fallback Integration
-   **Description:** Integrate with local AI models (e.g., via Ollama) to provide offline processing capabilities, reducing reliance on external AI providers and enhancing privacy/performance.

## 3. CLI & TUI Improvements

### 3.1 Populate TUI Views with Real Data
-   **Description:** The newly developed TUI views (`Bot Status`, `AI Providers`, `User Management`) currently display dummy data.
-   **Action:** Connect these views to the backend services to fetch and display real-time information regarding bot status, configured AI providers and their states, and SSH user details. This will involve:
    -   Implementing data fetching logic within each TUI view's `Update` method or using `tea.Cmd` for asynchronous data loading.
    -   Ensuring proper error handling and loading states for data retrieval.
