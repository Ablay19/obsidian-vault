export class APIError extends Error {
  public readonly code: string;
  public readonly statusCode: number;
  public readonly details?: Record<string, unknown>;

  constructor(code: string, message: string, statusCode: number = 500, details?: Record<string, unknown>) {
    super(message);
    this.name = 'APIError';
    this.code = code;
    this.statusCode = statusCode;
    this.details = details;
  }
}

export function createErrorResponse(error: unknown): { statusCode: number; body: Record<string, unknown> } {
  if (error instanceof APIError) {
    return {
      statusCode: error.statusCode,
      body: {
        success: false,
        error: error.message,
        code: error.code,
        details: error.details,
      },
    };
  }

  if (error instanceof Error) {
    return {
      statusCode: 500,
      body: {
        success: false,
        error: error.message,
      },
    };
  }

  return {
    statusCode: 500,
    body: {
      success: false,
      error: 'Unknown error occurred',
    },
  };
}

export function getUserFacingErrorMessage(error: unknown): string {
  if (error instanceof APIError) {
    return error.message;
  }

  if (error instanceof Error) {
    return error.message;
  }

  return 'An unexpected error occurred. Please try again.';
}

export class ValidationError extends APIError {
  constructor(field: string, message: string) {
    super('VALIDATION_ERROR', `${field}: ${message}`, 400, { field });
  }
}

export class NotFoundError extends APIError {
  constructor(resource: string, id: string) {
    super('NOT_FOUND', `${resource} with ID '${id}' not found`, 404, { resource, id });
  }
}

export class RateLimitError extends APIError {
  constructor(retryAfter?: number) {
    super('RATE_LIMIT', 'Too many requests. Please try again later.', 429, { retryAfter });
  }
}

export class ServiceUnavailableError extends APIError {
  constructor(service: string) {
    super('SERVICE_UNAVAILABLE', `${service} is currently unavailable. Please try again later.`, 503, { service });
  }
}
