# AI Manim Integration Plan for CLI Bots

## Overview

Integrate AI-powered mathematical animation video generation capabilities into the existing Mauritania CLI Telegram and WhatsApp bots. Leverage the already-deployed Cloudflare Workers (@workers/) to provide Manim video generation without requiring local Manim installation.

## Current Architecture

### Existing CLI Bots
- **Telegram Bot**: `cmd/mauritania-cli/internal/transports/telegram/telegram.go`
- **WhatsApp Bot**: `cmd/mauritania-cli/internal/transports/whatsapp/whatsapp.go`
- **Transport Layer**: Modular interface supporting text messaging
- **Session Management**: Basic command execution tracking

### Cloudflare Workers (@workers/)
- **ai-manim-worker**: Multi-provider AI service for Manim code generation
- **manim-renderer**: Python-based video rendering service
- **ai-proxy**: AI provider management and cost optimization
- **Deployment**: Already live on Cloudflare Workers platform

## Integration Strategy

### 1. Extend Transport Interfaces for Binary Content

**Current Limitation**: Transports only handle text messages
**Solution**: Add binary file transfer capabilities

```go
type Transport interface {
    // Existing methods...
    SendBinary(ctx context.Context, data []byte, metadata map[string]interface{}) error
    ReceiveBinary(ctx context.Context, sessionID string) ([]byte, error)
    SendFile(ctx context.Context, filePath string, metadata map[string]interface{}) error
}
```

### 2. AI Service Integration

**Approach**: Create AI service client that communicates with Cloudflare Workers

```go
type AIManimService interface {
    GenerateVideo(problem string, options VideoOptions) (*VideoJob, error)
    GetJobStatus(jobID string) (*JobStatus, error)
    DownloadVideo(jobID string) ([]byte, error)
}
```

### 3. Command Extensions

**New CLI Commands**:
- `/math <problem>` - Generate math explanation video
- `/animate <concept>` - Create animation for mathematical concept
- `/status <job_id>` - Check video generation status
- `/download <job_id>` - Download completed video

### 4. Session Management Enhancement

**Current**: Basic command tracking
**Enhanced**: Long-running AI job management

```go
type AIJob struct {
    JobID       string
    UserID      string
    Platform    string
    Problem     string
    Status      JobStatus
    CreatedAt   time.Time
    CompletedAt *time.Time
    VideoURL    string
}
```

## Implementation Phases

### Phase 1: Transport Extensions (Binary Support)

1. **Extend Transport Interfaces**
   - Add binary message methods to transport abstractions
   - Implement chunked file transfer for large videos
   - Add progress reporting for downloads

2. **Update Transport Implementations**
   - Telegram: Use document/file sharing APIs
   - WhatsApp: Use media upload APIs
   - Handle transport-specific limitations

### Phase 2: AI Service Integration

1. **Create AI Client Library**
   - HTTP client for Cloudflare Workers communication
   - Authentication and rate limiting
   - Error handling and retries

2. **Implement Service Interfaces**
   - Video generation request/response handling
   - Job status polling
   - Video download and caching

### Phase 3: CLI Command Integration

1. **Extend Bot Handlers**
   - Add math video generation commands
   - Implement conversation flow for AI interactions
   - Add progress notifications

2. **Update Session Management**
   - Track AI jobs across conversations
   - Handle interrupted sessions
   - Provide status updates

### Phase 4: Error Handling & Testing

1. **Comprehensive Error Handling**
   - Network failures and retries
   - AI service unavailability
   - Transport limitations

2. **Testing & Validation**
   - Unit tests for all components
   - Integration tests with Cloudflare Workers
   - End-to-end bot testing

## API Integration Details

### Cloudflare Workers Endpoints

**ai-manim-worker**:
- `POST /api/generate` - Generate Manim code from problem description
- `GET /api/job/{id}/status` - Check rendering status
- `GET /api/job/{id}/download` - Download completed video

**Authentication**: Use API tokens or Cloudflare authentication

### Request/Response Format

```json
// Generate Request
{
  "problem": "Explain Pythagorean theorem",
  "quality": "medium",
  "format": "mp4"
}

// Status Response
{
  "jobId": "abc123",
  "status": "rendering",
  "progress": 65,
  "estimatedTimeRemaining": 120
}

// Download Response
{
  "url": "https://temp-video-url.mp4",
  "expiresAt": "2024-01-20T10:00:00Z"
}
```

## Technical Considerations

### File Size Management
- Manim videos: 5-50MB typically
- Implement resumable downloads
- Temporary storage with cleanup
- Compression options for bandwidth

### Rate Limiting
- Respect Cloudflare Workers limits
- Implement client-side rate limiting
- Queue requests during high load

### Cost Optimization
- Cache frequently requested videos
- Optimize AI prompts for better results
- Monitor usage patterns

### Security
- Validate file uploads
- Secure API communications
- User authorization for video access

## Migration Strategy

1. **Backward Compatibility**: Existing text commands continue working
2. **Gradual Rollout**: Add AI commands incrementally
3. **Feature Flags**: Enable AI features per user/configuration
4. **Fallback Handling**: Graceful degradation when AI services unavailable

## Success Metrics

- 90% of math-related queries use AI video generation
- Average video generation time < 5 minutes
- User satisfaction rating > 4.0/5.0
- Successful delivery rate > 95%

## Dependencies

- Existing CLI transport infrastructure
- Cloudflare Workers deployment (already available)
- Video storage and CDN capabilities
- User authentication system

## Risk Mitigation

- **Service Availability**: Implement fallback to text explanations
- **Cost Control**: Monitor API usage and implement limits
- **User Experience**: Provide clear progress indicators
- **Technical Debt**: Maintain clean separation between CLI and AI services

This plan provides a clear path to integrate powerful AI capabilities into your existing CLI bots while leveraging the already-deployed Cloudflare infrastructure.