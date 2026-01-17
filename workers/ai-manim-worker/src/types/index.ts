import type { KVNamespace } from '@cloudflare/workers-types';

export type LogLevel = "debug" | "info" | "warn" | "error";

export type Platform = "telegram" | "whatsapp" | "web";

export interface Env {
  LOG_LEVEL: string;
  TELEGRAM_BOT_TOKEN: string;
  TELEGRAM_SECRET: string;
  SESSIONS: KVNamespace;
  CLOUDFLARE_API_TOKEN: string;
  CLOUDFLARE_ACCOUNT_ID: string;
  GROQ_API_KEY: string;
  HUGGINGFACE_API_KEY: string;
  HF_TOKEN: string;
  OPENAI_API_KEY: string;
  GEMINI_API_KEY: string;
  DEEPSEEK_API_KEY: string;
  AI_PROVIDER: string;
  MANIM_RENDERER_URL: string;
  USE_MOCK_RENDERER: string;
  WHATSAPP_WEBHOOK_SECRET?: string;
  AI_PROXY_URL?: string;
  AI_PROXY_TOKEN?: string;
}

export interface TelegramUpdate {
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

export interface UserSession {
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
  platform_primary?: Platform;   // Primary platform for this session

  // Video metadata only (no videos stored)
  video_history: VideoMetadata[];

  // Statistics
  total_submissions: number;
  successful_generations: number;

  // Accessibility preferences
  accessibility_enabled?: boolean;
  screen_reader_preferred?: boolean;
}

export interface VideoMetadata {
  job_id: string;
  problem_preview: string;
  status: ProcessingStatus;
  created_at: string;
  delivered_at?: string;
  render_duration_seconds?: number;
}

export type ProcessingStatus =
  | "queued"
  | "ai_generating"
  | "code_validating"
  | "rendering"
  | "uploading"
  | "ready"
  | "delivered"
  | "failed"
  | "expired";

export interface ProcessingJob {
  // Primary key: job_id (UUID v4)
  job_id: string;

  // Foreign keys
  session_id: string;  // Reference to UserSession

  // Submission metadata
  submission_type: "problem" | "direct_code";  // New field for direct code submissions
  platform: Platform;   // Source platform

  // Input
  problem_text?: string;       // User's mathematical problem (for AI generation)
  manim_code?: string;         // Direct code submission (for direct rendering)
  problem_language?: string;  // Detected language

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

export interface JobResponse {
  job_id: string;
  status: ProcessingStatus;
  message: string;
  created_at: string;
}

export interface HealthResponse {
  status: "healthy" | "degraded" | "unhealthy";
  version: string;
  timestamp: string;
  providers: {
    cloudflare: "available" | "unavailable";
    groq: "available" | "unavailable";
    huggingface: "available" | "unavailable";
  };
}

export interface VideoStorageConfig {
  type: 'ephemeral' | 'external' | 'r2';
  baseUrl: string;
  expirySeconds: number;
}

export interface VideoUploadResult {
  success: boolean;
  url?: string;
  key?: string;
  error?: string;
}

export interface VideoStorageService {
  upload(videoPath: string, jobId: string): Promise<VideoUploadResult>;
  getUrl(key: string): string;
  delete(key: string): Promise<boolean>;
}

export interface RenderedVideo {
  jobId: string;
  rendererUrl: string;
  localPath?: string;
  uploadedUrl?: string;
  expiresAt?: Date;
}

export interface RenderRequest {
  job_id: string;
  problem_text: string;
  manim_code: string;
  callback_url?: string;
}

export interface RenderResponse {
  job_id: string;
  status: "success" | "error" | "timeout";
  video_path?: string;
  video_size_bytes?: number;
  render_duration_seconds?: number;
  error?: string;
}

export interface JobStateSync {
  job_id: string;
  status: ProcessingStatus;
  timestamp: string;
  component: string;
  data?: Record<string, unknown>;
}

export interface WhatsAppMessage {
  id: string;
  from: string;  // Phone number
  type: "text" | "image" | "document";
  timestamp: string;
  text?: {
    body: string;
  };
  image?: {
    id: string;
    mime_type: string;
    sha256: string;
  };
  document?: {
    id: string;
    mime_type: string;
    sha256: string;
    filename?: string;
  };
}

export interface DirectCodeSubmission {
  code: string;
  options?: {
    quality: "low" | "medium" | "high";
    format: "mp4" | "webm";
    max_duration_seconds?: number;
  };
}

