# TODO - Obsidian Automation Bot

This document outlines future improvements and tasks for the Obsidian Automation Bot.

## Blockers
*No active blockers.*

## Future Improvements

1.  **Per-User Language Settings**: Store the language preference on a per-user or per-chat basis instead of a single global variable.
    *   **Status**: Pending
    *   **Priority**: Medium
    *   **Notes**: This would require a schema change to the `chats` table to add a `language` column and modifications to the `/lang` command handler.

2.  **Advanced Search Command**: Implement a `/search` command to search for notes within the Obsidian vault directly from Telegram.
    *   **Status**: Pending
    *   **Priority**: Medium
    *   **Notes**: This would involve adding a new command handler and logic to search Markdown files in the `vault` directory.

3.  **Expanded File Type Support**: Add support for processing more document types, such as `.docx` or plain `.txt` files.
    *   **Status**: Pending
    *   **Priority**: Low
    *   **Notes**: This would require integrating additional libraries or tools for text extraction from these formats (e.g., `pandoc` for `.docx`).

4.  **Local AI Fallback (Ollama)**: Complete the integration with a local AI provider like Ollama as a fallback.
    *   **Status**: Partially implemented
    *   **Priority**: Low
    *   **Notes**: This would provide an offline-first option.

## Completed

-   **Web Dashboard**: A simple web dashboard to view stats and bot status graphically. The dashboard is now implemented using Templ, HTMX, and Alpine.js, providing a real-time view of the bot's status.
-   **Refactor Dockerfile**: Simplified the `Dockerfile` to a single-stage build for a smaller image and faster builds. Removed the Python dependency.
-   **Refactor Makefile**: Cleaned up the `Makefile`, improved help text, and simplified targets.
-   **Database-driven PID Lock**: Replaced the file-based PID lock with a database-driven approach in `internal/database/instance.go` to better support containerized deployments.
