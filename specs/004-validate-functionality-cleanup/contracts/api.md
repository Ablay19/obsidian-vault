# API Contracts for Validation Service

## Overview

This section defines the RESTful API contracts for the functionality validation, documentation management, and directory cleanup service.

## Validation Endpoints

### GET /api/v1/features
Retrieves all features and their validation status.

**Response:**
```json
{
  "features": [
    {
      "id": "mcp-server",
      "name": "MCP Server",
      "description": "Model Context Protocol server implementation",
      "module": "internal/mcp",
      "status": "validated",
      "test_coverage": 0.85,
      "last_validated": "2026-01-18T20:00:00Z",
      "test_cases": [
        {
          "id": "test-001",
          "name": "Server startup",
          "feature_id": "mcp-server",
          "type": "integration",
          "status": "passing",
          "last_run": "2026-01-18T20:00:00Z",
          "duration": "1.2s"
        }
      ]
    }
  ]
}
```

### POST /api/v1/features/{id}/validate
Triggers validation for a specific feature.

**Path Parameters:**
- `id` (string): Feature identifier

**Response:**
```json
{
  "validation_id": "val-12345",
  "status": "started",
  "estimated_duration": "5m"
}
```

### GET /api/v1/features/{id}/coverage
Retrieves test coverage details for a specific feature.

**Path Parameters:**
- `id` (string): Feature identifier

**Response:**
```json
{
  "feature_id": "mcp-server",
  "overall_coverage": 0.85,
  "coverage_by_file": {
    "server.go": 0.90,
    "tools.go": 0.75,
    "transport.go": 0.80
  },
  "uncovered_lines": [15, 42, 78],
  "last_analyzed": "2026-01-18T20:00:00Z"
}
```

## Documentation Endpoints

### GET /api/v1/documents
Retrieves all documentation and their completeness status.

**Response:**
```json
{
  "documents": [
    {
      "id": "doc-readme-001",
      "type": "README",
      "location": "README.md",
      "last_updated": "2026-01-18T20:00:00Z",
      "completeness_score": 0.90,
      "coverage_feature": "general",
      "maintained_by": "team"
    }
  ]
}
```

### POST /api/v1/documents/generate
Triggers documentation generation for missing documentation.

**Request Body:**
```json
{
  "target_type": "API",
  "feature_id": "mcp-server",
  "template": "standard"
}
```

**Response:**
```json
{
  "generation_id": "gen-67890",
  "status": "started",
  "estimated_duration": "2m"
}
```

## Directory Management Endpoints

### GET /api/v1/directories
Retrieves directory structure analysis.

**Response:**
```json
{
  "directories": [
    {
      "id": "dir-cmd-001",
      "path": "cmd/mauritania-cli",
      "purpose": "CLI application entry point",
      "file_count": 25,
      "last_cleaned": "2026-01-18T20:00:00Z",
      "module": "cli"
    }
  ]
}
```

### POST /api/v1/directories/cleanup
Triggers directory cleanup operation.

**Request Body:**
```json
{
  "targets": ["cmd/mauritania-cli", "internal"],
  "dry_run": false
}
```

**Response:**
```json
{
  "cleanup_id": "clean-11111",
  "status": "started",
  "files_to_remove": 12,
  "directories_to_remove": 2,
  "estimated_disk_saved": "5.2MB"
}
```

## Status Endpoints

### GET /api/v1/status
Retrieves overall system status and progress.

**Response:**
```json
{
  "overall_status": "healthy",
  "validation_progress": {
    "total_features": 8,
    "validated_features": 6,
    "failed_features": 0,
    "pending_features": 2
  },
  "documentation_progress": {
    "total_documents": 15,
    "complete_documents": 12,
    "in_progress_documents": 3,
    "average_completeness": 0.85
  },
  "cleanup_progress": {
    "total_directories": 12,
    "cleaned_directories": 8,
    "files_removed": 23,
    "disk_space_saved": "15.7MB"
  }
}
```

## Error Handling

All endpoints follow standard HTTP status codes:

- `200 OK`: Request successful
- `400 Bad Request`: Invalid request parameters
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

**Error Response Format:**
```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Feature validation failed",
    "details": "Test coverage below 70% threshold",
    "timestamp": "2026-01-18T20:00:00Z"
  }
}
```

## Rate Limiting

API endpoints are rate-limited to prevent abuse:
- 100 requests per minute per IP
- 1000 requests per hour globally

## Authentication

All API endpoints require authentication via API key:
```
Authorization: Bearer <api_key>
```

## Data Format

All requests and responses use JSON format with UTF-8 encoding.
Date/time fields use ISO 8601 format (YYYY-MM-DDTHH:mm:ssZ).

## OpenAPI Specification

This API can be generated as an OpenAPI 3.0 specification for integration with API documentation tools and client SDK generation.