import { REQUEST_TIMEOUT_MS, MAX_VIDEO_DURATION_SECONDS, RENDER_TIMEOUT_MS } from './constants';

export class TimeoutError extends Error {
  public readonly timeout: number;
  public readonly operation: string;

  constructor(operation: string, timeout: number) {
    super(`Operation '${operation}' timed out after ${timeout}ms`);
    this.name = 'TimeoutError';
    this.operation = operation;
    this.timeout = timeout;
  }
}

export async function withTimeout<T>(
  operation: string,
  promise: Promise<T>,
  timeoutMs?: number
): Promise<T> {
  const timeout = timeoutMs || REQUEST_TIMEOUT_MS;

  return Promise.race([
    promise,
    new Promise<never>((_, reject) =>
      setTimeout(() => reject(new TimeoutError(operation, timeout)), timeout)
    ),
  ]);
}

export function getTimeoutForOperation(operation: string): number {
  const timeouts: Record<string, number> = {
    'render': RENDER_TIMEOUT_MS,
    'video_generation': MAX_VIDEO_DURATION_SECONDS * 1000,
    'ai_generation': 60000,
    'upload': 60000,
    'default': REQUEST_TIMEOUT_MS,
  };

  return timeouts[operation] || timeouts.default;
}
