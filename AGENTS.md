# obsidian-vault Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-01-18

## Active Technologies
- Go 1.21+ + Doppler CLI, testify (testing), godotenv (.env handling) (002-add-e2e-testing)
- .env files, Doppler configs, local test databases (002-add-e2e-testing)
- Go 1.25.4 (from go.mod) + testify (testing), zap (logging), cobra (CLI), modelcontextprotocol/go-sdk (MCP integration) (004-validate-functionality-cleanup)
- File system for docs, SQLite for testing, environment variables for config (004-validate-functionality-cleanup)

- Go 1.21+ + github.com/modelcontextprotocol/go-sdk (001-add-mcp-integration)

## Project Structure

```text
src/
tests/
```

## Commands

# Add commands for Go 1.21+

## Code Style

Go 1.21+: Follow standard conventions

## Recent Changes
- 004-validate-functionality-cleanup: Added Go 1.25.4 (from go.mod) + testify (testing), zap (logging), cobra (CLI), modelcontextprotocol/go-sdk (MCP integration)
- 002-add-e2e-testing: Added Go 1.21+ + Doppler CLI, testify (testing), godotenv (.env handling)

- 001-add-mcp-integration: Added Go 1.21+ + github.com/modelcontextprotocol/go-sdk

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
