# TODO - Obsidian Automation Bot

This document outlines future improvements and tasks for the Obsidian Automation Bot, as identified during technical documentation review.

## Future Improvements

1.  **Implement Ollama Fallback**: The groundwork is laid for an Ollama fallback, but the final implementation in `processor.go` is needed. This would provide an offline/local AI option if the Gemini service is completely unavailable.
    *   **Status**: Pending
    *   **Priority**: High
    *   **Assigned To**: [Team/Developer Name]
    *   **Notes**: Requires integration with `processor.go` to handle Ollama as an alternative AI service.

2.  **Per-User Language Settings**: Store the language preference on a per-user or per-chat basis instead of a single global variable.
    *   **Status**: Pending
    *   **Priority**: Medium
    *   **Assigned To**: [Team/Developer Name]
    *   **Notes**: Involves modifying user data storage and language handling logic.

3.  **More Advanced Commands**: Implement a `/search` command to search within the Obsidian vault, or a `/config` command to change settings via the bot.
    *   **Status**: Pending
    *   **Priority**: Medium
    *   **Assigned To**: [Team/Developer Name]
    *   **Notes**: Requires extending the bot's command handling and potentially interacting with the Obsidian vault structure.

4.  **Web Dashboard**: Create a simple web dashboard (perhaps using the `health.go` server as a base) to view stats and bot status graphically.
    *   **Status**: Pending
    *   **Priority**: Low
    *   **Assigned To**: [Team/Developer Name]
    *   **Notes**: Could leverage existing `health.go` for basic stats, but will need new UI components.

5.  **Support for More File Types**: Add support for processing other document types like `.docx` or `.txt`.
    *   **Status**: Pending
    *   **Priority**: Low
    *   **Assigned To**: [Team/Developer Name]
    *   **Notes**: May require external libraries or tools for parsing new file formats.
