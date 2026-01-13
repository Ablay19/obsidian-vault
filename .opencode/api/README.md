# Opencode API

This directory contains API server configuration, endpoints, and integration capabilities.

## API Architecture

### Server Configuration
- **RESTful API** - Standard HTTP endpoints for all operations
- **WebSocket API** - Real-time communication and streaming
- **GraphQL API** - Flexible data querying (optional)
- **gRPC API** - High-performance binary protocol (optional)

### API Components
```
api/
├── server.js           # Main API server
├── routes/            # API route definitions
├── middleware/        # Request/response middleware
├── schemas/           # Request/response schemas
├── docs/              # API documentation
├── tests/             # API tests
└── examples/          # Usage examples
```

## API Configuration

### Server Settings (`config/api.yaml`)
```yaml
enabled: true
host: "localhost"
port: 8080
protocol: "http"        # http, https, both

cors:
  enabled: true
  origins: ["*"]       # Allowed origins
  methods: ["GET", "POST", "PUT", "DELETE"]
  headers: ["Content-Type", "Authorization", "X-Session-ID"]
  credentials: false

rate_limiting:
  enabled: true
  requests_per_minute: 100
  burst_limit: 20
  strategy: "sliding_window"

authentication:
  enabled: false        # Enable for production
  method: "jwt"         # jwt, api_key, oauth
  secret: "${API_SECRET}"
  expire_hours: 24

ssl:
  enabled: false
  cert_file: "./certs/server.crt"
  key_file: "./certs/server.key"
  
logging:
  enabled: true
  level: "info"
  include_headers: false
  include_body: false
```

## API Endpoints

### Core Operations

#### `/api/v1/tools` - Tool Management
```http
GET    /api/v1/tools              # List available tools
GET    /api/v1/tools/:tool         # Get tool details
POST   /api/v1/tools/:tool         # Execute tool
PUT    /api/v1/tools/:tool/config # Update tool config
DELETE /api/v1/tools/:tool         # Remove tool
```

#### `/api/v1/sessions` - Session Management
```http
GET    /api/v1/sessions           # List sessions
POST   /api/v1/sessions           # Create session
GET    /api/v1/sessions/:id       # Get session details
PUT    /api/v1/sessions/:id       # Update session
DELETE /api/v1/sessions/:id       # Delete session
```

#### `/api/v1/agents` - Agent Operations
```http
GET    /api/v1/agents             # List available agents
POST   /api/v1/agents/:agent      # Execute agent task
GET    /api/v1/agents/:agent/task # Get task status
DELETE /api/v1/agents/:agent/task # Cancel task
```

### File Operations

#### `/api/v1/files` - File Management
```http
GET    /api/v1/files              # List files
GET    /api/v1/files/:path        # Read file
POST   /api/v1/files              # Write/create file
PUT    /api/v1/files/:path        # Update file
DELETE /api/v1/files/:path        # Delete file
```

#### `/api/v1/search` - Search Operations
```http
GET    /api/v1/search/files       # Search files by pattern
POST   /api/v1/search/content     # Search file content
GET    /api/v1/search/history     # Search history
```

### System Operations

#### `/api/v1/system` - System Information
```http
GET    /api/v1/system/status      # System status
GET    /api/v1/system/health      # Health check
GET    /api/v1/system/metrics     # Performance metrics
GET    /api/v1/system/config      # Configuration
```

#### `/api/v1/cache` - Cache Management
```http
GET    /api/v1/cache/stats        # Cache statistics
POST   /api/v1/cache/clear        # Clear cache
PUT    /api/v1/cache/config       # Update cache config
```

## API Schemas

### Request/Response Models

#### Tool Execution Request
```json
{
  "tool": "bash",
  "parameters": {
    "command": "npm install",
    "description": "Install dependencies",
    "timeout": 30000
  },
  "options": {
    "async": false,
    "stream": false
  }
}
```

#### Tool Execution Response
```json
{
  "success": true,
  "data": {
    "stdout": "npm packages installed",
    "stderr": "",
    "exit_code": 0,
    "duration": 15234
  },
  "metadata": {
    "tool": "bash",
    "timestamp": "2026-01-13T10:30:00Z",
    "session_id": "ses_abc123"
  }
}
```

#### Agent Task Request
```json
{
  "agent": "general",
  "task": "Analyze codebase structure",
  "parameters": {
    "description": "comprehensive analysis",
    "thoroughness": "medium"
  },
  "options": {
    "timeout": 300000,
    "stream_progress": true
  }
}
```

#### Agent Task Response
```json
{
  "task_id": "task_123",
  "status": "running",
  "progress": {
    "percentage": 45,
    "current_step": "analyzing components",
    "estimated_completion": "2026-01-13T10:35:00Z"
  },
  "results": null
}
```

## API Usage Examples

### Tool Execution
```javascript
// Execute bash command
const response = await fetch('/api/v1/tools/bash', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    parameters: {
      command: 'ls -la',
      description: 'List directory contents'
    }
  })
});

const result = await response.json();
console.log(result.data.stdout);
```

### File Operations
```javascript
// Read file
const readResponse = await fetch('/api/v1/files/src/main.js');
const fileData = await readResponse.json();

// Write file
const writeResponse = await fetch('/api/v1/files', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    path: 'src/new-file.js',
    content: 'console.log("Hello World");'
  })
});
```

### Agent Tasks
```javascript
// Start agent task
const taskResponse = await fetch('/api/v1/agents/general', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    task: 'Review and optimize this code',
    parameters: {
      files: ['src/main.js', 'src/utils.js']
    }
  })
});

const { task_id } = await taskResponse.json();

// Check task status
const statusResponse = await fetch(`/api/v1/agents/general/task?id=${task_id}`);
const status = await statusResponse.json();
```

## WebSocket API

### Connection
```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
  console.log('Connected to Opencode WebSocket API');
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  handleMessage(message);
};
```

### WebSocket Events

#### Tool Execution Streaming
```json
{
  "type": "tool_stream",
  "data": {
    "tool": "bash",
    "output": "npm packages installing...",
    "stream_type": "stdout"
  }
}
```

#### Agent Progress Updates
```json
{
  "type": "agent_progress",
  "data": {
    "task_id": "task_123",
    "progress": 75,
    "current_step": "generating report",
    "message": "Analysis complete, generating documentation"
  }
}
```

#### System Events
```json
{
  "type": "system_event",
  "data": {
    "event": "cache_cleared",
    "timestamp": "2026-01-13T10:30:00Z",
    "details": {
      "cache_type": "files",
      "entries_cleared": 45
    }
  }
}
```

## API Authentication

### JWT Authentication
```javascript
// Login to get token
const loginResponse = await fetch('/api/v1/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    username: 'user',
    password: 'password'
  })
});

const { token } = await loginResponse.json();

// Use token for authenticated requests
const response = await fetch('/api/v1/tools', {
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  }
});
```

### API Key Authentication
```javascript
const response = await fetch('/api/v1/tools', {
  headers: {
    'X-API-Key': 'your-api-key-here',
    'Content-Type': 'application/json'
  }
});
```

## API Documentation

### OpenAPI Specification
```yaml
openapi: 3.0.0
info:
  title: "Opencode API"
  version: "1.0.0"
  description: "API for Opencode AI assistant operations"

servers:
  - url: "http://localhost:8080/api/v1"
    description: "Development server"

paths:
  /tools:
    get:
      summary: "List available tools"
      responses:
        '200':
          description: "List of tools"
          content:
            application/json:
              schema:
                type: "array"
                items:
                  $ref: "#/components/schemas/Tool"
  
  /tools/{tool}:
    post:
      summary: "Execute tool"
      parameters:
        - name: "tool"
          in: "path"
          required: true
          schema:
            type: "string"
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/ToolRequest"
      responses:
        '200':
          description: "Tool execution result"
```

### Interactive Documentation
- Swagger UI at `/docs`
- ReDoc at `/redoc`
- API playground at `/playground`

## API Testing

### Automated Tests
```javascript
// Example API test
describe('Tools API', () => {
  test('should list available tools', async () => {
    const response = await fetch('/api/v1/tools');
    const tools = await response.json();
    
    expect(response.status).toBe(200);
    expect(tools).toContain('bash');
    expect(tools).toContain('read');
  });
  
  test('should execute bash command', async () => {
    const response = await fetch('/api/v1/tools/bash', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        parameters: {
          command: 'echo "test"',
          description: 'Test command'
        }
      })
    });
    
    const result = await response.json();
    expect(response.status).toBe(200);
    expect(result.success).toBe(true);
    expect(result.data.stdout).toContain('test');
  });
});
```

### Load Testing
```bash
# Load test with artillery
artillery run load-test-config.yaml

# Performance test with k6
k6 run performance-test.js
```

## API Monitoring

### Metrics
- Request count and response times
- Error rates by endpoint
- Concurrent connections
- Memory and CPU usage
- Cache hit rates

### Health Checks
```bash
# Basic health check
curl http://localhost:8080/api/v1/system/health

# Detailed status
curl http://localhost:8080/api/v1/system/status

# Performance metrics
curl http://localhost:8080/api/v1/system/metrics
```