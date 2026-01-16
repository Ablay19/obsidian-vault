export const REQUEST_TIMEOUT_MS = 300000;
export const MAX_VIDEO_SIZE_BYTES = 50 * 1024 * 1024;
export const MAX_VIDEO_DURATION_SECONDS = 300;
export const MAX_PROBLEM_TEXT_LENGTH = 5000;
export const MIN_PROBLEM_TEXT_LENGTH = 10;
export const SESSION_TTL_SECONDS = 7 * 24 * 60 * 60;
export const JOB_TTL_SECONDS = {
  queued: 3600,
  ai_generating: 600,
  code_validating: 300,
  rendering: 600,
  uploading: 300,
  ready: 86400,
  delivered: 604800,
  failed: 86400,
  expired: 86400,
} as const;
export const VIDEO_PRESIGNED_URL_EXPIRY_SECONDS = 300;
export const RETRY_ATTEMPTS = 3;
export const RETRY_DELAY_MS = 1000;
export const AI_RETRY_ATTEMPTS = 3;
export const RATE_LIMIT_PER_MINUTE = 10;
export const RATE_LIMIT_PER_HOUR = 100;
export const DEFAULT_LANGUAGE = 'en';
export const MANIM_VERSION = '0.18.0';
export const RENDER_TIMEOUT_MS = 300000;
