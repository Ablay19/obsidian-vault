import { APIResponse, ErrorResponse } from '@obsidian-vault/shared-types';

export interface RequestConfig {
  method: string;
  endpoint: string;
  headers?: Record<string, string>;
  body?: unknown;
  timeout?: number;
}

export interface ResponseData {
  statusCode: number;
  body: unknown;
  headers: Headers;
  duration: number;
}

export class HttpClient {
  private baseURL: string;
  private logger: ReturnType<typeof import('@obsidian-vault/shared-types').createLogger>;
  private timeout: number;
  private maxRetries: number = 0; // Fail-fast: no retries

  constructor(baseURL: string, logger: ReturnType<typeof import('@obsidian-vault/shared-types').createLogger>, timeout = 30000) {
    this.baseURL = baseURL;
    this.logger = logger;
    this.timeout = timeout;
  }

  setTimeout(timeout: number): void {
    this.timeout = timeout;
  }

  async do(config: RequestConfig): Promise<APIResponse> {
    const startTime = Date.now();
    const url = `${this.baseURL}${config.endpoint}`;

    this.logger.debug('Making HTTP request', {
      method: config.method,
      url,
      timeout: config.timeout || this.timeout,
    });

    const options: RequestInit = {
      method: config.method,
      headers: {
        'Content-Type': 'application/json',
        ...config.headers,
      },
    };

    if (config.body) {
      options.body = JSON.stringify(config.body);
    }

    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), config.timeout || this.timeout);
    options.signal = controller.signal;

    try {
      const response = await fetch(url, options);
      clearTimeout(timeoutId);

      const duration = Date.now() - startTime;

      this.logger.info('HTTP request completed', {
        method: config.method,
        url,
        status: response.status,
        duration_ms: duration,
      });

      // Fail-fast: return error immediately on non-2xx status
      if (response.status < 200 || response.status >= 300) {
        const errorBody = await response.json() as ErrorResponse;
        this.logger.error('HTTP request failed', 
          new Error(errorBody.message || 'Unknown error'),
          { status: response.status, url, details: errorBody.details }
        );
        throw new Error(`HTTP ${response.status}: ${errorBody.message || response.statusText}`);
      }

      const apiResponse = (await response.json()) as APIResponse;

      this.logger.debug('Response parsed successfully', { status: apiResponse.status });
      return apiResponse;
    } catch (error) {
      clearTimeout(timeoutId);

      if (error instanceof Error) {
        if (error.name === 'AbortError') {
          this.logger.error('HTTP request timed out', error, { url, timeout: this.timeout });
        } else {
          this.logger.error('HTTP request failed', error, { url });
        }
      } else {
        this.logger.error('Unknown HTTP request error', undefined, { url });
      }

      throw error;
    }
  }

  async get(endpoint: string, headers?: Record<string, string>): Promise<APIResponse> {
    return this.do({
      method: 'GET',
      endpoint,
      headers,
      timeout: this.timeout,
    });
  }

  async post(endpoint: string, body: unknown, headers?: Record<string, string>): Promise<APIResponse> {
    return this.do({
      method: 'POST',
      endpoint,
      headers,
      body,
      timeout: this.timeout,
    });
  }

  async put(endpoint: string, body: unknown, headers?: Record<string, string>): Promise<APIResponse> {
    return this.do({
      method: 'PUT',
      endpoint,
      headers,
      body,
      timeout: this.timeout,
    });
  }

  async delete(endpoint: string, headers?: Record<string, string>): Promise<APIResponse> {
    return this.do({
      method: 'DELETE',
      endpoint,
      headers,
      timeout: this.timeout,
    });
  }

  async healthCheck(endpoint: string): Promise<void> {
    this.logger.info('Performing health check', { endpoint });

    const startTime = Date.now();

    try {
      const response = await fetch(`${this.baseURL}${endpoint}`, {
        method: 'GET',
        signal: AbortSignal.timeout(this.timeout),
      });

      const duration = Date.now() - startTime;

      if (response.status !== 200) {
        this.logger.error('Health check failed', 
          new Error('Unexpected status code'),
          { status: response.status }
        );
        throw new Error(`Health check failed: unexpected status code ${response.status}`);
      }

      this.logger.info('Health check passed', { duration_ms: duration });
    } catch (error) {
      if (error instanceof Error) {
        this.logger.error('Health check failed', error);
      } else {
        this.logger.error('Unknown health check error', undefined);
      }
      throw error;
    }
  }
}

export function createHttpClient(
  baseURL: string,
  logger: ReturnType<typeof import('@obsidian-vault/shared-types').createLogger>,
  timeout = 30000
): HttpClient {
  return new HttpClient(baseURL, logger, timeout);
}