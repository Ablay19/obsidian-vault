# Data Model: AI Manim Video Generator with WhatsApp & Direct Code Enhancement

**Feature**: 006-ai-manim-video
**Date**: January 17, 2026
**Based On**: [spec.md](spec.md) + [research.md](research.md) + Session clarifications

---

## Entities

### UserSession (Enhanced)

**Purpose**: Track anonymous user sessions across multiple platforms (Telegram, WhatsApp, Web)

**Storage**: Cloudflare KV (7-day TTL, auto-extend)

```typescript
interface UserSession {
  // Primary key: session_id (UUID v4)
  session_id: string;

  // Platform-specific identifiers
  telegram_chat_id?: string;     // Telegram chat ID
  whatsapp_phone_number?: string; // WhatsApp phone number
  web_session_token?: string;    // Web dashboard session token

  // Timestamps
  created_at: string;      // ISO 8601
  last_activity: string;   // ISO 8601 (auto-extend on activity)

  // Session metadata
  language_preference?: string;  // Default: "en"
  platform_primary?: "telegram" | "whatsapp" | "web";

  // Video metadata only (no videos stored)
  video_history: VideoMetadata[];

  // Statistics
  total_submissions: number;
  successful_generations: number;

  // Accessibility preferences
  accessibility_enabled?: boolean;
  screen_reader_preferred?: boolean;
}
```

**Relationships**:
- One-to-many with `ProcessingJob`
- No direct video storage (videos deleted immediately)

---

### ProcessingJob (Enhanced)

**Purpose**: Track the state and progress of video generation requests from any platform

**Storage**: Cloudflare KV

```typescript
interface ProcessingJob {
  // Primary key: job_id (UUID v4)
  job_id: string;

  // Foreign keys
  session_id: string;  // Reference to UserSession

  // Submission metadata
  submission_type: "problem" | "direct_code";  // New field for direct code submissions
  platform: "telegram" | "whatsapp" | "web";   // Source platform

  // Input
  problem_text?: string;       // User's mathematical problem (for AI generation)
  manim_code?: string;         // Direct code submission (for direct rendering)

  // Processing state
  status: ProcessingStatus;
  status_message?: string;    // Human-readable status

  // Timestamps
  created_at: string;         // ISO 8601
  started_at?: string;        // When processing began
  completed_at?: string;      // When video was ready

  // AI Generation (only for problem submissions)
  ai_provider_used?: string;      // "cloudflare", "groq", "huggingface"
  manim_code_generated?: string;  // AI-generated code
  ai_error?: string;              // Error if AI failed

  // Video Output (metadata only - video deleted after delivery)
  video_url?: string;             // Presigned URL for access
  video_key?: string;             // R2 object key (for deletion)
  video_expires_at?: string;      // URL expiration time

  // Quality metrics
  render_duration_seconds?: number;
  video_size_bytes?: number;

  // Error handling
  retry_count?: number;           // Number of retry attempts
  last_error?: string;            // Last error message
  graceful_degradation_used?: boolean; // If graceful degradation was applied
}
```

**Status Enum**:

```typescript
enum ProcessingStatus {
  QUEUED = "queued",           // Waiting to be processed
  AI_GENERATING = "ai_generating",  // AI creating Manim code (problem submissions)
  CODE_VALIDATING = "code_validating",  // Syntax check (direct code)
  RENDERING = "rendering",     // Manim rendering video
  UPLOADING = "uploading",     // Uploading to R2
  READY = "ready",             // Video available for download
  DELIVERED = "delivered",     // User accessed video
  FAILED = "failed",           // Processing failed
  EXPIRED = "expired"          // Not accessed within window
}
```

---

### VideoMetadata (Enhanced)

**Purpose**: Lightweight reference to delivered videos with accessibility information

**Storage**: UserSession.video_history array

```typescript
interface VideoMetadata {
  job_id: string;
  problem_preview: string;    // First 50 chars of problem or code
  submission_type: "problem" | "direct_code";
  status: ProcessingStatus;
  created_at: string;
  delivered_at?: string;      // When user accessed video
  platform: string;           // Delivery platform
  render_duration_seconds?: number;

  // Accessibility metadata
  has_audio_description?: boolean;
  keyboard_navigable?: boolean;
  screen_reader_friendly?: boolean;
}
```

---

### PlatformConfiguration

**Purpose**: Store platform-specific configuration and credentials

**Storage**: Cloudflare KV (environment-specific)

```typescript
interface PlatformConfiguration {
  platform: "telegram" | "whatsapp";

  // Telegram configuration
  telegram_bot_token?: string;
  telegram_webhook_secret?: string;

  // WhatsApp configuration
  whatsapp_api_key?: string;
  whatsapp_api_secret?: string;
  whatsapp_phone_number_id?: string;
  whatsapp_webhook_verify_token?: string;

  // Rate limiting
  rate_limit_per_hour: number;
  rate_limit_burst: number;

  // Feature flags
  direct_code_enabled: boolean;
  accessibility_enabled: boolean;
}
```

---

## Validation Rules

### UserSession

| Field | Rule |
|-------|------|
| session_id | UUID v4, required |
| telegram_chat_id | Numeric string, optional |
| whatsapp_phone_number | E.164 format, optional |
| web_session_token | UUID v4, optional |
| created_at | ISO 8601, required |
| last_activity | ISO 8601, required |
| video_history | Array, max 100 entries |
| total_submissions | Integer, >= 0 |
| platform_primary | Must match one of the provided platform IDs |

### ProcessingJob

| Field | Rule |
|-------|------|
| job_id | UUID v4, required |
| session_id | Reference to existing UserSession |
| submission_type | Required enum value |
| platform | Required platform identifier |
| problem_text | String, 10-5000 chars (if submission_type=problem) |
| manim_code | Valid Python code, max 50KB (if submission_type=direct_code) |
| status | Enum value, required |
| created_at | ISO 8601, required |
| retry_count | Integer, 0-3 (max retries) |

---

## API Data Types

### Platform-Agnostic Request Types

```typescript
interface VideoSubmissionRequest {
  platform: "telegram" | "whatsapp" | "web";
  submission_type: "problem" | "direct_code";
  content: string;  // problem text or Manim code
  user_id: string;  // platform-specific user identifier
}

interface VideoStatusRequest {
  job_id: string;
  platform: string;
}

interface VideoAccessRequest {
  job_id: string;
  platform: string;
  accessibility_options?: {
    audio_description: boolean;
    high_contrast: boolean;
    large_text: boolean;
  };
}
```

### Enhanced Platform-Specific Types

```typescript
interface TelegramUpdate {
  update_id: number;
  message?: {
    message_id: number;
    from?: {
      id: number;
      is_bot: boolean;
      language_code?: string;
    };
    chat: {
      id: number;
      type: string;
    };
    text?: string;
    date: number;
  };
}

interface WhatsAppWebhook {
  object: string;
  entry: Array<{
    id: string;
    changes: Array<{
      value: {
        messages?: Array<{
          id: string;
          from: string;
          type: string;
          timestamp: string;
          text?: { body: string };
        }>;
        contacts?: Array<{
          profile: { name: string };
          wa_id: string;
        }>;
      };
      field: string;
    }>;
  }>;
}

interface DirectCodeSubmission {
  code: string;
  options?: {
    quality: "low" | "medium" | "high";
    format: "mp4" | "webm";
    max_duration_seconds?: number;
  };
}
```

---

## Database Schema (KV)

### Keys

| Entity | Key Pattern | Example |
|--------|-------------|---------|
| UserSession | `session:{uuid}` | `session:550e8400-e29b-41d4-a716-446655440000` |
| ProcessingJob | `job:{uuid}` | `job:6ba7b810-9dad-11d1-80b4-00c04fd430c8` |
| PlatformConfig | `config:platform:{platform}` | `config:platform:whatsapp` |
| SessionIndex | `sessions:by_telegram:{chat_id}` | `sessions:by_telegram:123456789` |
| SessionIndex | `sessions:by_whatsapp:{phone}` | `sessions:by_whatsapp:+1234567890` |
| SessionIndex | `sessions:by_web:{token}` | `sessions:by_web:abc123def456` |
| JobIndex | `jobs:by_session:{session_id}` | `jobs:by_session:550e8400...` |

### TTL Strategy

| Entity | TTL | Reason |
|--------|-----|--------|
| UserSession | 7 days | Session expiration |
| ProcessingJob (QUEUED) | 1 hour | Processing timeout |
| ProcessingJob (READY) | 24 hours | Access window |
| ProcessingJob (DELIVERED) | 7 days | Audit trail |
| ProcessingJob (FAILED) | 24 hours | Retry window |
| PlatformConfig | No TTL | Persistent config |

---

## API Contracts Directory Structure

```
contracts/
├── openapi.yaml              # Main API specification
├── telegram-webhook.yaml     # Telegram-specific contracts
├── whatsapp-webhook.yaml     # WhatsApp-specific contracts
├── direct-code-api.yaml      # Direct code submission contracts
└── video-delivery.yaml       # Video access and delivery contracts
```

---

## Compliance Notes

- **Multi-Platform**: Single session can span multiple platforms
- **Privacy-First**: No persistent data, immediate video deletion
- **Accessibility**: WCAG 2.1 AAA compliance with metadata tracking
- **Graceful Degradation**: Clear error handling with retry options
- **Security**: Basic input validation, platform-specific authentication
- **Scalability**: Flexible scaling with no strict performance limits

---

## Relationships Diagram

```
UserSession (1) ───────┬────── (N) ProcessingJob
    │                   │
    ├── telegram_chat_id    ├── submission_type
    ├── whatsapp_phone      ├── platform
    └── web_session_token   ├── problem_text OR manim_code
                            └── video_url (temporary)

PlatformConfiguration (1) ──── Manages ─── (N) UserSession
        │
        ├── telegram_config
        ├── whatsapp_config
        └── rate_limits
```</content>
<parameter name="filePath">specs/006-ai-manim-video/data-model.md