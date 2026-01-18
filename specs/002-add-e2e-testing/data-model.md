# Data Model: E2E Testing with Doppler Environment Variables

**Date**: January 18, 2026
**Feature**: 002-add-e2e-testing

## Overview

The data model supports Doppler integration for secure environment variable management in E2E testing. All entities focus on configuration management, credential security, and test environment isolation.

## Core Entities

### DopplerConfig

**Purpose**: Configuration settings for Doppler CLI integration and project setup.

**Fields**:
- `Project`: string (required) - Doppler project name
- `Config`: string (required) - Doppler config name (dev, staging, prod)
- `Token`: string (sensitive) - Doppler service token (encrypted)
- `Endpoint`: string - Custom Doppler API endpoint
- `Timeout`: duration - API timeout (default 30s)

**Validation**:
- Project: alphanumeric + hyphens, max 50 chars
- Config: valid environment name
- Token: encrypted storage, never logged

**Relationships**:
- 1:many with TestEnvironment
- references EnvironmentVariable (via Doppler)

### EnvironmentVariable

**Purpose**: Individual environment variable with metadata and security classification.

**Fields**:
- `Name`: string (required) - Variable name (e.g., API_KEY)
- `Value`: string (sensitive) - Variable value (encrypted at rest)
- `Description`: string - Human-readable description
- `Required`: bool - Whether variable must be present for tests
- `Sensitive`: bool - Whether variable contains sensitive data
- `Environment`: string - Target environment (dev, staging, prod)
- `Source`: enum (doppler, env, default) - Variable source

**Validation**:
- Name: uppercase with underscores, max 100 chars
- Value: encrypted for sensitive vars
- Description: max 200 chars

**Relationships**:
- belongs to TestEnvironment
- grouped by CredentialSet for related variables

### TestEnvironment

**Purpose**: Named configuration set for different testing scenarios.

**Fields**:
- `Name`: string (required) - Environment name (e.g., integration, e2e, staging)
- `Description`: string - Purpose and scope
- `Variables`: []EnvironmentVariable - Associated variables
- `Active`: bool - Whether environment is currently active
- `CreatedAt`: timestamp - Environment creation time
- `UpdatedAt`: timestamp - Last modification time

**Validation**:
- Name: lowercase alphanumeric + hyphens, max 50 chars
- Description: max 500 chars

**Relationships**:
- references DopplerConfig
- contains multiple EnvironmentVariable
- 1:many with CredentialSet

### CredentialSet

**Purpose**: Grouped API credentials for transport integrations.

**Fields**:
- `Name`: string (required) - Set name (e.g., whatsapp-creds, telegram-api)
- `Type`: enum (api_key, oauth, basic_auth) - Authentication type
- `Variables`: []string - Required variable names
- `Transport`: string - Associated transport (whatsapp, telegram, etc.)
- `Validated`: bool - Whether credentials have been validated

**Validation**:
- Name: descriptive, max 100 chars
- Variables: valid EnvironmentVariable names

**Relationships**:
- belongs to TestEnvironment
- references EnvironmentVariable

### FallbackConfig

**Purpose**: Alternative variable sources when Doppler is unavailable.

**Fields**:
- `Priority`: int (required) - Fallback priority (1=primary, higher=fallback)
- `Type`: enum (env_file, defaults, vault) - Fallback source type
- `Path`: string - File path for .env files
- `Enabled`: bool - Whether fallback is active
- `Variables`: map[string]string - Default values for non-sensitive vars

**Validation**:
- Priority: positive integer
- Path: valid file path if type is env_file

**Relationships**:
- referenced by TestEnvironment for fallback scenarios

## Data Flow

1. **Test Execution** → DopplerConfig loads project settings
2. **Environment Selection** → TestEnvironment provides variable set
3. **Variable Resolution** → EnvironmentVariable sourced from Doppler/fallback
4. **Credential Injection** → CredentialSet provides transport credentials
5. **Fallback Handling** → FallbackConfig provides alternatives

## Validation Rules

### Global Rules
- Sensitive data never persisted in plain text
- Environment variables validated before test execution
- Doppler tokens rotated regularly (not stored long-term)
- Fallback configs provide minimal viable defaults

### Business Rules
- Each TestEnvironment must have at least one DopplerConfig
- CredentialSet variables must exist in associated EnvironmentVariable
- FallbackConfig priority determines fallback order (1=first, higher=later)
- EnvironmentVariable names follow UPPER_SNAKE_CASE convention

## Security Considerations

- Doppler tokens encrypted at rest and in transit
- EnvironmentVariable values hashed in logs (sensitive=true)
- CredentialSet validation prevents incomplete configurations
- FallbackConfig avoids exposing defaults in error messages

## Performance Considerations

- DopplerConfig cached for session duration
- EnvironmentVariable lazy-loaded on first access
- TestEnvironment configurations pre-validated
- FallbackConfig provides fast local alternatives