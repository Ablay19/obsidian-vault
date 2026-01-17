export interface RenderRequest {
  job_id: string;
  problem_text: string;
  manim_code: string;
  quality?: 'low' | 'medium' | 'high' | 'ultra';
  format?: 'mp4' | 'webm';
  callback_url?: string;
}

export interface RenderResponse {
  job_id: string;
  status: 'success' | 'error' | 'timeout';
  video_path?: string;
  video_size_bytes?: number;
  render_duration_seconds?: number;
  error?: string;
  error_code?: string;
}

export interface CallbackPayload {
  job_id: string;
  status: RenderResponse['status'];
  video_path?: string;
  video_size_bytes?: number;
  render_duration_seconds?: number;
  error?: string;
  timestamp: string;
}

export interface RendererConfig {
  base_url: string;
  timeout_seconds: number;
  max_retries: number;
  retry_delay_ms: number;
}

export interface HealthCheckResponse {
  status: 'healthy' | 'unhealthy';
  manim_version?: string;
  ffmpeg_available: boolean;
  active_jobs: number;
  timestamp: string;
}

export type WorkerToRendererMessage =
  | { type: 'render_request'; payload: RenderRequest }
  | { type: 'cancel_request'; job_id: string }
  | { type: 'health_check' };

export type RendererToWorkerMessage =
  | { type: 'render_complete'; payload: CallbackPayload }
  | { type: 'render_progress'; job_id: string; progress: number }
  | { type: 'render_error'; job_id: string; error: string }
  | { type: 'health_response'; payload: HealthCheckResponse };
