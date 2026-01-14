# Shared Types Package

This package provides type definitions shared between Go and TypeScript applications.

## Structure

```
packages/shared-types/
├── go/
│   ├── types.go           # Go type definitions
│   ├── go.mod             # Go module
│   └── tests/
│       └── types_test.go  # Type tests
├── typescript/
│   ├── index.ts           # TypeScript interfaces
│   ├── logger.ts          # Logging utilities
│   └── package.json       # NPM package config
└── README.md              # This file
```

## Usage

### Go

```go
import "github.com/abdoullahelvogani/obsidian-vault/packages/shared-types/go"

func Example() {
    worker := types.WorkerModule{
        ID:      "worker-001",
        Name:    "ai-worker",
        Version: "1.0.0",
        Status:  "active",
    }

    logger := types.NewColoredLogger("my-component")
    logger.Info("Worker created", "id", worker.ID)
}
```

### TypeScript

```typescript
import { createLogger, WorkerModule } from '@obsidian-vault/shared-types/typescript';

const logger = createLogger('my-worker');

interface WorkerConfig {
  id: string;
  name: string;
}
```

## Available Types

- `WorkerModule` - Cloudflare Worker deployment unit
- `GoApplication` - Go backend service
- `SharedPackage` - Reusable package definition
- `APIGateway` - API gateway configuration
- `DeploymentPipeline` - CI/CD pipeline definition
- `APIResponse` / `ErrorResponse` - Standard API response types
- `Pagination` - Pagination metadata
- `LogConfig` - Logging configuration

## Versioning

This package follows semantic versioning. Changes are tracked via git tags.

## Contributing

When adding new types:
1. Add to both Go and TypeScript implementations
2. Ensure JSON serialization compatibility
3. Add tests for validation logic
