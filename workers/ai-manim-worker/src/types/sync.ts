import type { ProcessingStatus } from './index';

export interface JobStateSync {
  job_id: string;
  status: ProcessingStatus;
  status_message?: string;
  timestamp: string;
  component: 'worker' | 'renderer' | 'r2' | 'telegram';
  data?: JobStateData;
  error?: string;
}

export interface JobStateData {
  progress_percent?: number;
  video_size_bytes?: number;
  render_duration_seconds?: number;
  provider_used?: string;
  attempt_number?: number;
}

export interface StateTransition {
  from_status: ProcessingStatus;
  to_status: ProcessingStatus;
  allowed: boolean;
  reason?: string;
}

export const STATE_TRANSITIONS: Record<ProcessingStatus, StateTransition[]> = {
  queued: [
    { from_status: 'queued', to_status: 'ai_generating', allowed: true },
    { from_status: 'queued', to_status: 'failed', allowed: true, reason: 'invalid_problem' },
  ],
  ai_generating: [
    { from_status: 'ai_generating', to_status: 'code_validating', allowed: true },
    { from_status: 'ai_generating', to_status: 'failed', allowed: true, reason: 'ai_generation_failed' },
  ],
  code_validating: [
    { from_status: 'code_validating', to_status: 'rendering', allowed: true },
    { from_status: 'code_validating', to_status: 'failed', allowed: true, reason: 'code_validation_failed' },
  ],
  rendering: [
    { from_status: 'rendering', to_status: 'uploading', allowed: true },
    { from_status: 'rendering', to_status: 'failed', allowed: true, reason: 'render_failed' },
    { from_status: 'rendering', to_status: 'failed', allowed: true, reason: 'render_timeout' },
  ],
  uploading: [
    { from_status: 'uploading', to_status: 'ready', allowed: true },
    { from_status: 'uploading', to_status: 'failed', allowed: true, reason: 'upload_failed' },
  ],
  ready: [
    { from_status: 'ready', to_status: 'delivered', allowed: true },
    { from_status: 'ready', to_status: 'expired', allowed: true, reason: 'not_accessed_in_time' },
  ],
  delivered: [
    { from_status: 'delivered', to_status: 'expired', allowed: true },
  ],
  failed: [
    { from_status: 'failed', to_status: 'queued', allowed: true, reason: 'retry' },
  ],
  expired: [
    { from_status: 'expired', to_status: 'queued', allowed: true, reason: 'new_submission' },
  ],
};

export function isValidTransition(from: ProcessingStatus, to: ProcessingStatus): boolean {
  const transitions = STATE_TRANSITIONS[from];
  if (!transitions) return false;
  return transitions.some(t => t.to_status === to && t.allowed);
}

export function canRetry(status: ProcessingStatus): boolean {
  return status === 'failed' || status === 'expired';
}

export interface SyncEvent {
  job_id: string;
  event_type: 'state_change' | 'progress_update' | 'error_occurred' | 'completed';
  previous_status?: ProcessingStatus;
  current_status: ProcessingStatus;
  timestamp: string;
  metadata?: Record<string, unknown>;
}
