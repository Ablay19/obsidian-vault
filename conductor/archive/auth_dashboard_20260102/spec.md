# Specification: Authoritative Dashboard Implementation

## 1. Overview

This document outlines the specification for a fully functional control dashboard that acts as the authoritative runtime control plane for the AI bot. The dashboard is not just a UI; it is the central nervous system for runtime configuration, provider management, and bot control.

## 2. Core Responsibilities

*   **Runtime Configuration Ownership:** The dashboard owns the `RuntimeConfig` and is the single source of truth for all runtime settings.
*   **Provider and API Key Management:** It controls the usage of AI providers and their API keys, including enabling, disabling, blocking, and rotation.
*   **Environment Routing:** The dashboard manages environment routing (e.g., dev vs. prod) and prevents environment leakage.
*   **Bot Process Control:** It has the authority to stop, resume, or force fallback AI processing for the bot.
*   **Configuration Override:** The dashboard's configuration overrides any static configuration from `.env` files at runtime. The bot must never read `.env` files dynamically after its initial bootstrap.

## 3. Architecture

The data and control flow follows this strict path:

**Browser → HTMX → Dashboard Handlers → Runtime State Manager → AI Service → Providers**

*   **Bot's Data Source:** The bot reads its configuration *only* from the `RuntimeConfig` provided by the State Manager. It never reads from `.env`.
*   **Server-Side State:** All state mutations are triggered by HTMX requests, but all decisions and state changes happen exclusively on the server-side.
*   **Rendering:** Templ is responsible for server-side rendering of HTML. Alpine.js is optional and may only be used for non-essential local UI enhancements.

## 4. API Key Management

*   **Dynamic Management:** Keys can be added and removed dynamically via the dashboard.
*   **Individual Control:** Each key can be individually enabled or disabled.
*   **Health Tracking:** The system must track key health, including quota status, rate-limit errors, and invalid key states.
*   **Automatic Rotation:** The backend will implement automatic key rotation logic when a key becomes unhealthy.
*   **Persistence:** All dashboard actions related to keys must update the `RuntimeConfig` and be persisted to the database.
*   **Interface Enforcement:** The bot must access keys only through a `KeyManager` interface, which prevents it from overriding dashboard-set policies.

## 5. Runtime Environment Control

*   **Atomic Switching:** The dashboard must allow for atomic switching between `dev` and `prod` modes.
*   **Backend Configuration:** The backend host and base URL for services must be configurable through the dashboard.
*   **Strict Isolation:** The system must enforce strict environment isolation to prevent `dev` configurations or traffic from leaking into `prod`.
*   **Bot Awareness:** The bot reads the current `EnvironmentState` from the `RuntimeConfig`.

## 6. Bot Control Plane

The `RuntimeConfig` is the authoritative state and includes:
*   `AIEnabled` (global switch)
*   `Providers` (map of provider states)
*   `APIKeys` (map of API key states)
*   `EnvironmentState`

Before every API call, the bot must perform the following checks against the `RuntimeConfig`:
1.  Is the global `AIEnabled` flag true?
2.  Is the specific provider enabled?
3.  Is the selected API key valid and enabled?
4.  Does the request match the current `EnvironmentState`?

The dashboard must be able to hard-disable providers or keys, forcing the bot to a fallback or to stop processing.

## 7. Styling & Rendering

*   **Static Assets:** All static assets (CSS, JS) must be served correctly via an appropriate file server.
*   **CSS Preservation:** HTMX swaps must only replace HTML fragments and must not interfere with the main page layout or CSS links in the `<head>`.
*   **HTML Fragments:** All HTMX-driven handler responses must return HTML fragments, never a full HTML page.
*   **Base Template:** A single base template must define all global CSS and layout structure.
*   **JS-less Usability:** The dashboard must be fundamentally usable without JavaScript.

## 8. Data & State Model

*   **RuntimeConfig (Singleton):** The single, authoritative configuration object.
*   **ProviderState:** A struct holding the state for each AI provider.
*   **APIKeyState:** A struct holding the state for each API key.
*   **EnvironmentState:** A struct for `dev`/`prod` state.
*   **Persistence:** State must be stored in memory for fast access, with critical state changes persisted to the database.
*   **Concurrency:** State must be managed with concurrency in mind, using tools like `sync.RWMutex` and atomic swaps.
*   **Read-Only Interfaces:** The bot must interact with state via read-only interfaces.

## 9. Common Pitfalls to Avoid

*   Treating the dashboard as a read-only display.
*   Allowing the bot to read `.env` dynamically after startup.
*   Mixing UI logic into backend handlers.
*   Returning JSON instead of HTML fragments for HTMX requests.
*   Allowing providers or keys to self-enable.
*   Trusting any state sent from the frontend.
