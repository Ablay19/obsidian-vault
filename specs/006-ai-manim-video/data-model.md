# Data Model: AI Manim Video Generator

**Feature**: 006-ai-manim-video  
**Date**: January 15, 2026  
**Based On**: [spec.md](spec.md) + [research.md](research.md)

---

## Entities

### UserSession

**Purpose**: Track anonymous user sessions for video generation requests

**Storage**: Cloudflare KV (7-day TTL, auto-extend)

```typescript
interface UserSession {
  // Primary key: session_id (UUID v4)
  session_id: string;
  
  // Telegram chat ID for delivery
  telegram_chat_id: string;
  
  // Timestamps
  created_at: string;      // ISO 8601
  last_activity: string;   // ISO 8601 (auto-extend on activity)
  
  // Session metadata
  language_preference?: string;  // Default: "en"
  
  // Video metadata only (no videos stored)
  video_history: VideoMetadata[];
  
  // Statistics
  total_submissions: number;
  successful_generations: number;
}
```

**Relationships**:
- One-to-many with `ProcessingJob`
- No direct video storage (videos deleted immediately)

---

### ProcessingJob

**Purpose**: Track the state and progress of video generation requests

**Storage**: Cloudflare KV

```typescript
interface ProcessingJob {
  // Primary key: job_id (UUID v4)
  job_id: string;
  
  // Foreign keys
  session_id: string;  // Reference to UserSession
  
  // Input
  problem_text: string;       // User's mathematical problem
  problem_language?: string;  // Detected language
  
  // Processing state
  status: ProcessingStatus;
  status_message?: string;    // Human-readable status
  
  // Timestamps
  created_at: string;         // ISO 8601
  started_at?: string;        // When AI processing began
  completed_at?: string;      // When video was ready
  
  // AI Generation
  ai_provider_used?: string;      // "cloudflare", "groq", "huggingface"
  manim_code?: string;            // Generated code (for debugging)
  ai_error?: string;              // Error if AI failed
  
  // Video Output (metadata only - video deleted after delivery)
  video_url?: string;             // Presigned URL for access
  video_key?: string;             // R2 object key (for deletion)
  video_expires_at?: string;      // URL expiration time
  
  // Quality metrics
  render_duration_seconds?: number;
  video_size_bytes?: number;
}
```

**Status Enum**:

```typescript
enum ProcessingStatus {
  QUEUED = "queued",           // Waiting to be processed
  AI_GENERATING = "ai_generating",  // AI creating Manim code
  CODE_VALIDATING = "code_validating",  // Syntax check
  RENDERING = "rendering",     // Manim rendering video
  UPLOADING = "uploading",     // Uploading to R2
  READY = "ready",             // Video available for download
  DELIVERED = "delivered",     // User accessed video
  FAILED = "failed",           // Processing failed
  EXPIRED = "expired"          // Not accessed within window
}
```

**State Transition Diagram**:

```
QUEUED → AI_GENERATING → CODE_VALIDATING → RENDERING → UPLOADING → READY
                                                                      ↓
  FAILED ←──── AI_GENERATING   ←──── CODE_VALIDATING   ←──── RENDERING ←───
                      (retry)            (fix code)          (timeout)

READY → DELIVERED → (video auto-deleted)
READY → EXPIRED → (cleanup after 24h)
```

---

### VideoMetadata

**Purpose**: Lightweight reference to delivered videos (for history/analytics)

**Storage**: UserSession.video_history array

```typescript
interface VideoMetadata {
  job_id: string;
  problem_preview: string;    // First 50 chars of problem
  status: ProcessingStatus;
  created_at: string;
  delivered_at?: string;      // When user accessed video
  render_duration_seconds?: number;
}
```

---

## Validation Rules

### UserSession

| Field | Rule |
|-------|------|
| session_id | UUID v4, required |
| telegram_chat_id | Numeric string, required |
| created_at | ISO 8601, required |
| last_activity | ISO 8601, required |
| video_history | Array, max 100 entries |
| total_submissions | Integer, >= 0 |

### ProcessingJob

| Field | Rule |
|-------|------|
| job_id | UUID v4, required |
| session_id | Reference to existing UserSession |
| problem_text | String, 10-5000 chars, required |
| status | Enum value, required |
| created_at | ISO 8601, required |

---

## API Data Types

### Telegram Webhook Payload

```typescript
interface TelegramUpdate {
  update_id: number;
  message?: {
    message_id: number;
    from: {
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
```

### Video Generation Request (Internal)

```typescript
interface VideoGenerationRequest {
  job_id: string;
  problem_text: string;
  session_id: string;
  telegram_chat_id: string;
  priority: "normal";
}
```

### Video Generation Response (to User)

```typescript
interface VideoGenerationResponse {
  status: "queued" | "processing" | "ready" | "failed";
  job_id: string;
  message: string;
  video_url?: string;  // Only when status=ready
  expires_at?: string; // URL expiration
}
```

---

## Database Schema (KV)

### Keys

| Entity | Key Pattern | Example |
|--------|-------------|---------|
| UserSession | `session:{uuid}` | `session:550e8400-e29b-41d4-a716-446655440000` |
| ProcessingJob | `job:{uuid}` | `job:6ba7b810-9dad-11d1-80b4-00c04fd430c8` |
| SessionIndex | `sessions:by_chat:{chat_id}` | `sessions:by_chat:123456789` |
| JobIndex | `jobs:by_session:{session_id}` | `jobs:by_session:550e8400...` |

### TTL Strategy

| Entity | TTL | Reason |
|--------|-----|--------|
| UserSession | 7 days | Session expiration |
| ProcessingJob (QUEUED) | 1 hour | Processing timeout |
| ProcessingJob (READY) | 24 hours | Access window |
| ProcessingJob (DELIVERED) | 7 days | Audit trail |
| ProcessingJob (FAILED) | 24 hours | Retry window |

---

## Relationships Diagram

```
UserSession (1) ───────┬────── (N) ProcessingJob
   │                           │
   └── stores ─────────────────┘
        video_history[]
        (metadata only)
```

```
ProcessingJob (1) ────── produces ────── VideoFile
        │                              │
        └── metadata ──────────────────┘
             video_url
             video_key
             video_size_bytes
```

---

## Compliance Notes

- **Privacy**: No video content stored, only metadata
- **Retention**: All entities have TTL/expiration
- **Anonymity**: No PII, session-based tracking only
- **Audit**: 7-day metadata retention for debugging
