# API Reference

This document provides comprehensive API documentation for all Obsidian Vault services.

## Base URLs

| Environment | URL |
|-------------|-----|
| Development | `http://localhost:8080` |
| Staging | `https://staging-api.obsidianvault.com` |
| Production | `https://api.obsidianvault.com` |

## Content Type

All requests and responses use JSON:
```
Content-Type: application/json
```

## Authentication

### Bearer Token
```http
Authorization: Bearer <your_token>
```

### API Key
```http
X-API-Key: <your_api_key>
```

---

## API Gateway Endpoints

### Health Check

**GET** `/health`

Returns service health status.

**Response (200 OK)**
```json
{
  "status": "healthy",
  "timestamp": "2026-01-17T06:00:00Z",
  "version": "1.0.0",
  "services": {
    "database": "healthy",
    "cache": "healthy",
    "storage": "healthy"
  }
}
```

---

### API Status

**GET** `/api/v1/status`

Returns API status and configuration.

**Response (200 OK)**
```json
{
  "status": "operational",
  "version": "1.0.0",
  "environment": "production",
  "features": {
    "whatsapp": true,
    "ai": true,
    "visualizer": true,
    "manim": true
  },
  "rate_limits": {
    "per_minute": 60,
    "per_hour": 1000
  }
}
```

---

## WhatsApp Service API

### Webhook

**POST** `/api/v1/whatsapp/webhook`

Receives WhatsApp message webhooks from Meta.

**Headers**
```
X-Hub-Signature-256: sha256=<signature>
X-Telegram-Bot-Api-Secret-Token: <your_secret>
```

**Request Body**
```json
{
  "object": "whatsapp_business_account",
  "entry": [
    {
      "id": "WHATSAPP_BUSINESS_ACCOUNT_ID",
      "changes": [
        {
          "field": "messages",
          "value": {
            "messaging_product": "whatsapp",
            "metadata": {
              "display_phone_number": "1234567890",
              "phone_number_id": "1234567890"
            },
            "messages": [
              {
                "id": "wamid.ABC123",
                "from": "15551234567",
                "timestamp": "1672531200",
                "type": "text",
                "text": {
                  "body": "Hello, I need help"
                }
              }
            ]
          }
        }
      ]
    }
  ]
}
```

**Response (200 OK)**
```json
{
  "success": true,
  "message_id": "wamid.ABC123"
}
```

**Response (401 Unauthorized)**
```json
{
  "error": "Invalid webhook signature",
  "code": "INVALID_SIGNATURE"
}
```

---

### Send Message

**POST** `/api/v1/whatsapp/messages`

Sends a WhatsApp message to a user.

**Headers**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body**
```json
{
  "phone": "15551234567",
  "type": "text",
  "message": {
    "body": "Hello! How can I help you today?"
  }
}
```

**Response (200 OK)**
```json
{
  "success": true,
  "message_id": "wmid.DEF456",
  "timestamp": "2026-01-17T06:00:00Z"
}
```

---

### Send Template Message

**POST** `/api/v1/whatsapp/messages/template`

Sends a WhatsApp template message.

**Request Body**
```json
{
  "phone": "15551234567",
  "template_name": "hello_world",
  "language": "en",
  "components": [
    {
      "type": "body",
      "parameters": [
        {
          "type": "text",
          "text": "John"
        }
      ]
    }
  ]
}
```

**Response (200 OK)**
```json
{
  "success": true,
  "message_id": "wmid.GHI789",
  "template_name": "hello_world",
  "status": "sent"
}
```

---

### Media Upload

**POST** `/api/v1/whatsapp/media`

Uploads media for WhatsApp messaging.

**Headers**
```
Authorization: Bearer <token>
Content-Type: multipart/form-data
```

**Request**: File upload with `media` field

**Response (200 OK)**
```json
{
  "success": true,
  "media_id": "media_id_123",
  "mime_type": "image/png",
  "file_size": 102400
}
```

---

## AI Service API

### Chat Completion

**POST** `/api/v1/ai/chat`

Creates a chat completion.

**Headers**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body**
```json
{
  "model": "gpt-4o",
  "messages": [
    {
      "role": "system",
      "content": "You are a helpful assistant."
    },
    {
      "role": "user",
      "content": "Explain quantum computing in simple terms"
    }
  ],
  "temperature": 0.7,
  "max_tokens": 1000,
  "stream": false
}
```

**Response (200 OK)**
```json
{
  "id": "chatcmpl-abc123",
  "object": "chat.completion",
  "created": 1677858242,
  "model": "gpt-4o",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Quantum computing is..."
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 25,
    "completion_tokens": 150,
    "total_tokens": 175
  }
}
```

---

### Stream Chat Completion

**POST** `/api/v1/ai/chat`

Set `stream: true` for streaming response.

**Response (Chunked)**
```
data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1677858242,"model":"gpt-4o","choices":[{"index":0,"delta":{"content":"Quantum"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1677858242,"model":"gpt-4o","choices":[{"index":0,"delta":{"content":" computing"},"finish_reason":null}]}

data: [DONE]
```

---

### Generate Manim Code

**POST** `/api/v1/ai/manim`

Generates Manim animation code for mathematical visualizations.

**Request Body**
```json
{
  "problem": "Show the Pythagorean theorem with a visual proof",
  "style": "educational",
  "duration": 30,
  "quality": "high"
}
```

**Response (200 OK)**
```json
{
  "success": true,
  "job_id": "job_abc123",
  "status": "queued",
  "estimated_time": 120,
  "code": "from manim import *\n\nclass PythagoreanTheorem(Scene):\n    def construct(self):\n        ..."
}
```

---

### Check Generation Status

**GET** `/api/v1/ai/manim/{job_id}`

**Response (200 OK)**
```json
{
  "job_id": "job_abc123",
  "status": "ready",
  "video_url": "https://r2.obsidianvault.com/videos/job_abc123.mp4",
  "video_expires_at": "2026-01-17T07:00:00Z",
  "render_duration_seconds": 45
}
```

---

## Visualizer API

### Analyze Image

**POST** `/api/v1/visualizer/analyze`

Analyzes an image and generates a description.

**Request Body (multipart/form-data)**
```
image: <file>
context: "What's in this image?"
options: {
  "detail_level": "high",
  "include_text": true
}
```

**Response (200 OK)**
```json
{
  "success": true,
  "analysis": {
    "description": "A modern office space with natural lighting",
    "objects": ["desk", "chair", "monitor", "plant"],
    "text_detected": false,
    "confidence": 0.95
  }
}
```

---

### Generate Visualization

**POST** `/api/v1/visualizer/generate`

Generates a data visualization.

**Request Body**
```json
{
  "data": [10, 20, 30, 40, 50],
  "type": "bar_chart",
  "title": "Monthly Sales",
  "labels": ["Jan", "Feb", "Mar", "Apr", "May"],
  "style": {
    "colors": ["#3498db", "#e74c3c"],
    "background": "white"
  }
}
```

**Response (200 OK)**
```json
{
  "success": true,
  "visualization_id": "viz_abc123",
  "image_url": "https://r2.obsidianvault.com/visualizations/viz_abc123.png"
}
```

---

## RAG API

### Query Knowledge Base

**POST** `/api/v1/rag/query`

Queries the knowledge base using RAG.

**Request Body**
```json
{
  "query": "How do I configure OAuth authentication?",
  "filters": {
    "category": "documentation",
    "tags": ["auth", "oauth"]
  },
  "options": {
    "top_k": 5,
    "include_sources": true
  }
}
```

**Response (200 OK)**
```json
{
  "success": true,
  "answer": "To configure OAuth authentication...",
  "sources": [
    {
      "id": "doc_123",
      "title": "OAuth Configuration Guide",
      "url": "https://docs.obsidianvault.com/auth/oauth",
      "relevance_score": 0.95
    }
  ],
  "metadata": {
    "query_tokens": 10,
    "answer_tokens": 150,
    "retrieval_time_ms": 45
  }
}
```

---

### Add Document to Knowledge Base

**POST** `/api/v1/rag/documents`

Adds a document to the knowledge base.

**Request Body (multipart/form-data)**
```
file: <document.pdf or .md file>
metadata: {
  "title": "API Documentation",
  "category": "documentation",
  "tags": ["api", "reference"]
}
```

**Response (200 OK)**
```json
{
  "success": true,
  "document_id": "doc_abc123",
  "status": "indexed",
  "chunks_created": 45
}
```

---

## Session API

### Create Session

**POST** `/api/v1/sessions`

Creates a new user session.

**Request Body**
```json
{
  "user_id": "user_123",
  "metadata": {
    "source": "telegram",
    "language": "en"
  }
}
```

**Response (201 Created)**
```json
{
  "success": true,
  "session_id": "ses_abc123",
  "expires_at": "2026-01-24T06:00:00Z",
  "created_at": "2026-01-17T06:00:00Z"
}
```

---

### Get Session

**GET** `/api/v1/sessions/{session_id}`

**Response (200 OK)**
```json
{
  "session_id": "ses_abc123",
  "user_id": "user_123",
  "created_at": "2026-01-17T06:00:00Z",
  "last_activity": "2026-01-17T06:30:00Z",
  "metadata": {
    "source": "telegram"
  },
  "video_history": [
    {
      "job_id": "job_abc123",
      "problem_preview": "Explain Pythagorean theorem",
      "status": "delivered",
      "created_at": "2026-01-17T06:00:00Z"
    }
  ]
}
```

---

### List User Sessions

**GET** `/api/v1/users/{user_id}/sessions`

**Response (200 OK)**
```json
{
  "success": true,
  "sessions": [
    {
      "session_id": "ses_abc123",
      "created_at": "2026-01-17T06:00:00Z",
      "last_activity": "2026-01-17T06:30:00Z",
      "total_submissions": 5
    }
  ],
  "total": 1,
  "page": 1,
  "per_page": 20
}
```

---

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `INVALID_REQUEST` | 400 | Malformed request body |
| `UNAUTHORIZED` | 401 | Invalid or missing authentication |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `RATE_LIMITED` | 429 | Too many requests |
| `INTERNAL_ERROR` | 500 | Server-side error |
| `SERVICE_UNAVAILABLE` | 503 | Service temporarily unavailable |

### Error Response Format

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "Invalid parameter 'temperature': must be between 0 and 2",
    "details": {
      "parameter": "temperature",
      "constraint": "0 <= temperature <= 2"
    }
  },
  "request_id": "req_abc123",
  "timestamp": "2026-01-17T06:00:00Z"
}
```

---

## Rate Limits

| Endpoint | Requests/Minute | Requests/Hour |
|----------|-----------------|---------------|
| `/api/v1/ai/*` | 60 | 1000 |
| `/api/v1/whatsapp/*` | 120 | 5000 |
| `/api/v1/rag/*` | 30 | 500 |
| `/api/v1/visualizer/*` | 20 | 200 |

Rate limit headers included in response:
```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1672531260
```

---

## Versioning

API versioning is handled via URL path prefix: `/api/v1/`

Breaking changes will be introduced in new versions (e.g., `/api/v2/`), maintaining backward compatibility with previous versions for at least 6 months.

---

## Deprecation Policy

- Deprecation notices are given 90 days before removal
- Deprecated endpoints return `Deprecation` header
- Sunset headers indicate final removal date
