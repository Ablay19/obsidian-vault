import type { RenderRequest, RenderResponse } from '../types';
import { createLogger } from '../utils/logger';

const logger = createLogger({ level: 'info', component: 'renderer-client' });

const RETRY_ATTEMPTS = 3;
const RETRY_DELAY_MS = 1000;

export class RendererError extends Error {
  public readonly jobId: string;
  public readonly statusCode?: number;

  constructor(message: string, jobId: string, statusCode?: number) {
    super(message);
    this.name = 'RendererError';
    this.jobId = jobId;
    this.statusCode = statusCode;
  }
}

export class RendererClient {
  private rendererUrl: string;
  private timeout: number;

  constructor(config: { rendererUrl: string; timeout?: number }) {
    this.rendererUrl = config.rendererUrl;
    this.timeout = config.timeout || 300000;
  }

  async submitRender(request: RenderRequest): Promise<RenderResponse> {
    logger.info('Submitting render request', { job_id: request.job_id });

    let lastError: Error | null = null;

    for (let attempt = 1; attempt <= RETRY_ATTEMPTS; attempt++) {
      try {
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), this.timeout);

        const response = await fetch(`${this.rendererUrl}/render`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(request),
          signal: controller.signal,
        });

        clearTimeout(timeoutId);

        if (!response.ok) {
          const error = await response.text();
          throw new RendererError(`Renderer returned ${response.status}: ${error}`, request.job_id, response.status);
        }

        const result = (await response.json()) as RenderResponse;

        logger.info('Render request successful', {
          job_id: request.job_id,
          status: result.status,
          attempt,
        });

        return result;
      } catch (error) {
        lastError = error as Error;
        logger.warn('Render request failed, retrying', {
          job_id: request.job_id,
          attempt,
          error: (error as Error).message,
        });

        if (attempt < RETRY_ATTEMPTS) {
          await new Promise(resolve => setTimeout(resolve, RETRY_DELAY_MS));
        }
      }
    }

    logger.error(`All render request attempts failed for job ${request.job_id}: ${lastError?.message}`);

    throw lastError || new Error('Render request failed');
  }

  async getRenderStatus(jobId: string): Promise<RenderResponse> {
    try {
      const response = await fetch(`${this.rendererUrl}/status/${jobId}`);

      if (!response.ok) {
        throw new RendererError(`Status check failed: ${response.status}`, jobId, response.status);
      }

      return (await response.json()) as RenderResponse;
    } catch (error) {
      logger.error(`Failed to check render status for job ${jobId}: ${(error as Error).message}`);
      throw error;
    }
  }
}
