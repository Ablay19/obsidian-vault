# Opencode Tools Configuration

This directory contains configurations for all available tools in the Opencode ecosystem.

## Available Tools

### 📁 **File System Tools**
- **Read** - File reading with encoding and size limits
- **Write** - File writing with backup and validation  
- **Edit** - Precise string replacement with conflict detection
- **Glob** - Fast file pattern matching with recursive search
- **Grep** - Content search with regex support

### 🔧 **Development Tools**
- **Bash** - Shell command execution with persistent sessions
- **Websearch** - Real-time web search with Exa AI
- **Codesearch** - Programming documentation and API search

### 🤖 **AI & Agent Tools**
- **Task** - Specialized AI agent orchestration
- **Question** - User clarification and feedback collection
- **Skill** - Load specialized skill instructions

### 📋 **Productivity Tools**
- **TodoWrite** - Create and manage task lists
- **TodoRead** - Read and analyze task lists
- **Fetch** - URL content retrieval

## Tool Configuration Structure

Each tool has:
- `name` - Tool identifier
- `description` - Human-readable description
- `parameters` - Input parameter definitions
- `settings` - Default configurations
- `security` - Safety restrictions
- `usage_examples` - Common usage patterns

## Tool Integration

Tools are automatically loaded and can:
- Share context and state
- Execute in parallel when safe
- Chain operations together
- Validate inputs and outputs
- Handle errors gracefully

## Adding New Tools

1. Create configuration file in this directory
2. Implement tool logic in `.opencode/tools/`
3. Register in main configuration
4. Add documentation and examples

## Security Model

All tools include:
- Input validation and sanitization
- Permission and access controls
- Resource limits and timeouts
- Audit logging
- Error handling

Example tool configuration can be found in `example-tool.yaml`.