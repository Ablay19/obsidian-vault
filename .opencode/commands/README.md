# Opencode Commands Registry

This directory contains all available Opencode commands and their configurations.

## Command Categories

### 🛠️ **Core Commands**
- `bash` - Execute shell commands with persistent session
- `read` - Read files from filesystem
- `write` - Write files to filesystem  
- `edit` - Make precise string replacements in files
- `glob` - Fast file pattern matching
- `grep` - Search file contents with regex

### 🤖 **AI & Intelligence Commands**
- `task` - Launch specialized AI agents
- `question` - Ask user questions for clarification
- `websearch` - Real-time web search with Exa AI
- `codesearch` - Search programming documentation and APIs

### 📋 **Project Management Commands**
- `todowrite` - Create and manage task lists
- `todoread` - Read current task lists
- `skill` - Load skill instructions for specific tasks

### 🔧 **System Commands**
- `fetch` - Retrieve content from URLs
- `log` - Access system logs and diagnostics

## Command Configuration

Each command has its own configuration file in this directory:
- `bash.yaml` - Bash tool settings
- `read.yaml` - File reading configuration
- `write.yaml` - File writing configuration
- `task.yaml` - AI agent task settings
- etc.

## Usage Examples

```bash
# Core file operations
/read path="src/main.js"
/write content="console.log('Hello')" filePath="test.js"
/edit oldString="old" newString="new" filePath="file.js"

# AI assistance
/task description="Review this code" subagent_type="general"
/websearch query="React hooks best practices"
/codesearch query="useState hook examples" tokensNum="3000"

# Project management
/todowrite todos='[{"content":"Fix bug","status":"pending","priority":"high","id":"1"}]'
/todoread

# System operations
/bash command="npm install" description="Install dependencies"
```