# API Contracts Package

This package contains the OpenAPI 3.0 specifications and contracts for communication between workers and Go applications in the architectural separation project.

## 📋 Contents

- `openapi.yaml` - Complete OpenAPI 3.0 specification for all API endpoints
- Generated client libraries (future)

## 🚀 API Overview

The API provides RESTful endpoints for managing:

- **Worker Modules**: JavaScript/TypeScript Cloudflare Workers
- **Go Applications**: Backend Go services
- **Shared Packages**: Reusable packages across components
- **Deployment Pipelines**: CI/CD pipeline management

## 🔧 Using the API Specification

### Viewing the API Documentation

```bash
# Install swagger-ui or use online viewers
npm install -g swagger-ui
swagger-ui openapi.yaml
```

### Generating Client Libraries

```bash
# Generate Go client
openapi-generator-cli generate -i openapi.yaml -g go -o generated/go

# Generate TypeScript client
openapi-generator-cli generate -i openapi.yaml -g typescript-fetch -o generated/typescript
```

### Validation

```bash
# Validate OpenAPI spec
npm install -g @apidevtools/swagger-parser
swagger-parser validate openapi.yaml
```

## 📊 API Endpoints

### Worker Module Management
- `GET /workers` - List all worker modules
- `POST /workers` - Create new worker module
- `GET /workers/{workerId}` - Get worker by ID
- `PUT /workers/{workerId}` - Update worker
- `DELETE /workers/{workerId}` - Delete worker

### Go Application Management
- `GET /go-applications` - List all Go applications
- `POST /go-applications` - Create new Go application
- `GET /go-applications/{appId}` - Get Go application by ID

### Shared Package Management
- `GET /shared-packages` - List all shared packages

### Deployment Pipeline Management
- `GET /deployment-pipelines` - List deployment pipelines
- `POST /deployment-pipelines/{pipelineId}/deploy` - Trigger deployment

### Health & Monitoring
- `GET /health` - Health check endpoint
- `GET /metrics` - Prometheus metrics

## 🔒 Authentication & Security

All API endpoints implement:
- Input validation and sanitization
- Rate limiting
- Structured error responses
- Audit logging

## 📝 Error Handling

The API uses consistent error response format:

```json
{
  "error": "error_code",
  "message": "Human readable error message",
  "details": {}
}
```

## 🔄 Versioning

API follows semantic versioning with backward compatibility guarantees for minor versions.

## 🧪 Testing

API contracts are validated through:
- Unit tests for individual endpoints
- Integration tests for cross-component communication
- Contract tests ensuring API compliance
- End-to-end tests for complete workflows

## 📚 Development

### Updating the API Specification

1. Edit `openapi.yaml`
2. Validate the specification
3. Update client libraries if needed
4. Update tests
5. Update documentation

### Adding New Endpoints

1. Add endpoint definition to `openapi.yaml`
2. Implement endpoint in appropriate service
3. Add validation and error handling
4. Update tests
5. Update client libraries

## 🤝 Contributing

- Follow OpenAPI 3.0 specification standards
- Maintain backward compatibility
- Include comprehensive examples
- Add appropriate security considerations
- Update documentation for all changes