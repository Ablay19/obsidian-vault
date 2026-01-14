# Logging Guide

This document describes the logging implementation using slog with jq-style colored output.

## Overview

All services use Go's `log/slog` package for structured logging with:
- **Go services**: `ColoredJSONHandler` that outputs JSON with ANSI color codes
- **JavaScript workers**: `jqColorize()` function for the same jq-style colored JSON

## Color Scheme

The logging uses the same color scheme as jq:
- **Strings**: Orange (`\x1b[38;5;214m`)
- **Numbers**: Green (`\x1b[38;5;154m`)
- **Booleans/Null**: Yellow (`\x1b[38;5;220m`)
- **Brackets/Colons/Commas**: Blue (`\x1b[38;5;39m`)

## Usage

### Go Services

```go
import "github.com/abdoullahelvogani/obsidian-vault/packages/shared-types/go"

// Create colored logger
logger := types.NewColoredLogger("my-service")

// Log with structured data
logger.Info("User logged in", "user_id", "123", "action", "login")

// Log errors
logger.Error("Request failed", "error", err.Error(), "path", "/api/users")
```

### JavaScript Workers

```typescript
import { createLogger } from '@obsidian-vault/shared-types/typescript/logger';

const logger = createLogger('my-worker');

logger.info('Request received', { method: 'GET', path: '/api/users' });
logger.error('Processing failed', error, { user_id: '123' });
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `LOG_LEVEL` | Logging level (debug, info, warn, error) | info |

## Output Example

```
{"time":"2026-01-14T20:33:19.895Z","level":"INFO","msg":"User logged in","component":"auth-service","user_id":"123"}
```

With colors (when terminal supports ANSI):
- `"User logged in"` appears in orange
- `"auth-service"` appears in orange
- `"123"` appears in green
- `INFO` appears in yellow

## Shared Types Package

The logging utilities are available in:
- Go: `packages/shared-types/go/types.go`
- TypeScript: `packages/shared-types/typescript/logger.ts`
