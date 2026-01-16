import type { KVNamespace } from '@cloudflare/workers-types';

export type LogLevel = "debug" | "info" | "warn" | "error";

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
  session_id: string;
  telegram_chat_id: string;
  created_at: string;
  last_activity: string;
  language_preference?: string;
  video_history: VideoMetadata[];
  total_submissions: number;
  successful_generations: number;
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
  job_id: string;
  session_id: string;
  problem_text: string;
  problem_language?: string;
  status: ProcessingStatus;
  status_message?: string;
  created_at: string;
  started_at?: string;
  completed_at?: string;
  ai_provider_used?: string;
  manim_code?: string;
  ai_error?: string;
  video_url?: string;
  video_key?: string;
  video_expires_at?: string;
  render_duration_seconds?: number;
  video_size_bytes?: number;
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

