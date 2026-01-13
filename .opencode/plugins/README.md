# Opencode Plugin System

This directory contains plugins and extensions that enhance Opencode capabilities.

## Plugin Architecture

### Plugin Types
- **Tool Plugins** - Add new tools and functionality
- **Agent Plugins** - Create specialized AI agents
- **Integration Plugins** - Connect with external services
- **Theme Plugins** - Customize appearance and UI
- **Command Plugins** - Add custom commands

### Plugin Structure
Each plugin follows this structure:
```
plugin-name/
├── package.json          # Plugin metadata
├── index.js             # Main plugin entry point
├── config.yaml          # Plugin configuration
├── README.md            # Plugin documentation
├── examples/            # Usage examples
└── tests/               # Plugin tests
```

## Available Plugins

### @opencode-ai/plugin (Official)
**Version**: 1.1.18
**Description**: Core plugin providing standard Opencode functionality
**Features**:
- Basic tool implementations
- Standard agent configurations
- Session management
- Security features

### Plugin Configuration

#### `package.json`
```json
{
  "name": "my-opencode-plugin",
  "version": "1.0.0",
  "description": "Custom plugin for Opencode",
  "main": "index.js",
  "opencode": {
    "version": "^1.0.0",
    "permissions": ["file_access", "network"],
    "hooks": ["before_command", "after_command"]
  },
  "dependencies": {
    "@opencode-ai/sdk": "^1.0.0"
  }
}
```

#### `config.yaml`
```yaml
name: "my-plugin"
version: "1.0.0"
enabled: true

permissions:
  - "file_read"
  - "file_write"
  - "network_access"

settings:
  auto_load: true
  priority: 100

hooks:
  before_execution: "validateCommand"
  after_execution: "logResult"
```

## Plugin Development

### Creating a Tool Plugin
```javascript
// index.js
const { Tool } = require('@opencode-ai/sdk');

class MyTool extends Tool {
  constructor() {
    super({
      name: 'my-tool',
      description: 'Custom tool functionality'
    });
  }

  async execute(params) {
    // Tool implementation
    return { success: true, data: 'result' };
  }

  validate(params) {
    // Parameter validation
    return params.input ? true : false;
  }
}

module.exports = MyTool;
```

### Creating an Agent Plugin
```javascript
// index.js
const { Agent } = require('@opencode-ai/sdk');

class MyAgent extends Agent {
  constructor() {
    super({
      name: 'my-agent',
      type: 'specialized',
      capabilities: ['custom_task_1', 'custom_task_2']
    });
  }

  async process(task) {
    // Agent logic
    return this.executeTask(task);
  }

  async executeTask(task) {
    // Task execution implementation
  }
}

module.exports = MyAgent;
```

## Plugin Integration

### Plugin Loading
Plugins are automatically loaded on startup if:
- Listed in main configuration
- Enabled in plugin settings
- Pass validation checks

### Plugin Hooks
Plugins can hook into various events:
- `before_command` - Before tool execution
- `after_command` - After tool completion
- `on_error` - When errors occur
- `on_session_start` - When session begins
- `on_session_end` - When session ends

### Plugin Communication
Plugins can communicate via:
- Shared context objects
- Event system
- Direct function calls
- Message passing

## Plugin Security

### Sandboxing
- Limited file system access
- Network permission controls
- Resource usage limits
- Execution time restrictions

### Validation
- Code signature verification
- Dependency scanning
- Security vulnerability checks
- Permission validation

## Plugin Marketplace

### Official Repository
Access official plugins from:
```
https://plugins.opencode.ai
```

### Community Plugins
Community-contributed plugins available at:
```
https://github.com/opencode-community/plugins
```

### Plugin Installation
```bash
# Install official plugin
npm install @opencode-ai/plugin-name

# Install community plugin
npm install opencode-plugin-name

# Install from URL
opencode plugin install https://github.com/user/plugin.git
```

## Plugin Management

### List Plugins
```bash
opencode plugin list
```

### Enable/Disable
```bash
opencode plugin enable plugin-name
opencode plugin disable plugin-name
```

### Update Plugin
```bash
opencode plugin update plugin-name
```

### Remove Plugin
```bash
opencode plugin remove plugin-name
```

## Plugin Development Guidelines

### Best Practices
1. **Security First** - Always validate inputs and sanitize outputs
2. **Performance** - Optimize for speed and memory usage
3. **Documentation** - Provide comprehensive docs and examples
4. **Testing** - Include unit and integration tests
5. **Compatibility** - Support multiple Opencode versions

### Code Style
- Use ES6+ features
- Follow async/await patterns
- Implement proper error handling
- Add comprehensive logging

### Testing Requirements
- Unit tests with >80% coverage
- Integration tests for key workflows
- Performance benchmarks
- Security validation tests

## Plugin Support

### Documentation
- API reference docs
- Getting started guides
- Best practices
- Troubleshooting guide

### Community
- Discord server
- GitHub discussions
- Stack Overflow tag
- Monthly meetups

### Contributing
- Pull request guidelines
- Code review process
- Release process
- Support policy