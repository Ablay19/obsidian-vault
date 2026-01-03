# Specification: AI Provider Expansion and Dynamic Selection

## 1. Overview

This track aims to enhance the bot's AI capabilities by integrating two new AI providers: Hugging Face and OpenRouter. The implementation will add these as new, selectable options alongside the existing providers (Gemini, Groq). A key part of this effort is to refactor the provider selection mechanism to be more robust and user-friendly by introducing a dynamic, status-aware sub-menu within the Telegram bot command.

## 2. Functional Requirements

-   **Hugging Face Provider:**
    -   An `AIProvider` for Hugging Face must be implemented.
    -   It should be able to handle text generation and inference tasks.
    -   The implementation must be robust and well-supported.
-   **OpenRouter Provider:**
    -   An `AIProvider` for OpenRouter must be implemented.
-   **Dynamic Provider Selection Command:**
    -   The existing `/setprovider` Telegram command will be enhanced.
    -   When executed, the command shall present the user with a dynamic sub-menu of available AI providers.
    -   This menu will only display providers that are currently operational and healthy, based on a real-time status check.

## 3. Non-Functional Requirements

-   **Robustness:** The new provider integrations, especially for Hugging Face inference, must be stable and handle potential API errors gracefully.
-   **Configuration:** API keys and model names for the new providers must be configurable via the application's configuration files or environment variables.

## 4. Acceptance Criteria

-   A user can successfully select the Hugging Face provider via the improved Telegram command and receive a valid AI response.
-   A user can successfully select the OpenRouter provider via the improved Telegram command and receive a valid AI response.
-   The sub-menu in the `/setprovider` command accurately reflects the current health status of all providers (e.g., a downed provider does not appear as an option).

## 5. Out of Scope

-   Adding provider selection controls to the web dashboard. The focus is solely on the Telegram command interface.
-   Automatic switching of providers based on query content.