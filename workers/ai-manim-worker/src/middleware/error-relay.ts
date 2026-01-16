import type { Context } from 'hono';

const errors = new Map<string, { error: Error; timestamp: number }>();
const ERROR_TTL_MS = 60000;

export function createErrorRelayMiddleware() {
  return async (c: Context, next: () => Promise<void>) => {
    try {
      await next();
    } catch (error) {
      const errorKey = `${c.req.path}-${Date.now()}`;
      errors.set(errorKey, {
        error: error as Error,
        timestamp: Date.now(),
      });

      c.set('error', error);

      c.header('X-Error-ID', errorKey);

      return c.json(
        {
          success: false,
          error: (error as Error).message,
          error_id: errorKey,
        },
        500
      );
    }
  };
}

export function getErrorById(errorId: string): Error | null {
  const errorData = errors.get(errorId);

  if (!errorData) {
    return null;
  }

  if (Date.now() - errorData.timestamp > ERROR_TTL_MS) {
    errors.delete(errorId);
    return null;
  }

  return errorData.error;
}
