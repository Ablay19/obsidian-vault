# Plan: Implement the Authoritative Dashboard for Bot Control

This plan outlines the phases and tasks required to build the authoritative dashboard, which will serve as the central control plane for the bot.

---

## Phase 1: Foundation and State Management

This phase focuses on building the core data structures and management layer for the bot's runtime configuration.

- [x] Task: Build `RuntimeConfigManager` struct to manage state in-memory with database persistence. [c57257c]
- [x] Task: Define `ProviderState` and `APIKeyState` structs to model the state of AI providers and their keys. [a7de958]
- [x] Task: Refactor the AI Service to read its configuration exclusively from the `RuntimeConfigManager`. [1a71235]
- [x] Task: Conductor - User Manual Verification 'Foundation and State Management' (Protocol in workflow.md) [checkpoint: a4119db]

---

## Phase 2: Dashboard Backend and Core Logic

This phase implements the backend HTTP handlers that will allow the dashboard to mutate the bot's state.

- [x] Task: Implement dashboard HTTP handlers for managing AI providers and API keys (enable, disable, rotate). [550b00d]
- [x] Task: Implement dashboard HTTP handlers for controlling the runtime environment (switching between dev/prod). [32b0864]
- [x] Task: Enforce runtime configuration checks within the bot's logic before all AI provider calls. [a7429a7]
- [x] Task: Implement the persistence logic to save critical state changes to the database. [56cdbed]
- [x] Task: Conductor - User Manual Verification 'Dashboard Backend and Core Logic' (Protocol in workflow.md) [checkpoint: c63756c]

---

## Phase 3: Frontend Layout and Rendering

This phase focuses on building the user-facing HTML interface for the dashboard using Templ and HTMX.

- [x] Task: Design and implement the main panels, menus, and navigation sections for the dashboard UI. [56baa3f]
- [x] Task: Implement the base Templ layouts, configure static asset serving, and ensure HTMX fragment handling is correct. [f125867]
- [x] Task: Build the "Bot Status & Runtime Monitoring" panel to display active processes and AI runtime info. [427a3af]
- [x] Task: Build the "File Processing Panel" to show file status and summary previews. [1e8c21a]
- [~] Task: connected Users management panel and db congfigs
- [ ] Task: Build the "Interactive Q&A Console" for real-time interaction with the AI.
- [ ] Task: Build the "Settings & Config Management" UI to allow users to update keys and parameters.
- [ ] Task: Conductor - User Manual Verification 'Frontend Layout and Rendering' (Protocol in workflow.md)

---

## Phase 4: Final Integration and Testing

This phase ensures all components work together seamlessly and prepares the dashboard for use.

- [ ] Task: Prepare integration hooks for any future MCP server or automation scripts.
- [ ] Task: Write comprehensive integration tests for the full dashboard and bot control plane functionality.
- [ ] Task: Conductor - User Manual Verification 'Final Integration and Testing' (Protocol in workflow.md)
