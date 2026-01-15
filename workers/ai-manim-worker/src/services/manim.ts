import { createLogger } from '../utils/logger';

const logger = createLogger({ level: 'info', component: 'manim-renderer' });

export interface ManimRenderRequest {
  jobId: string;
  code: string;
  problem: string;
  outputFormat?: 'mp4' | 'webm';
  quality?: 'low' | 'medium' | 'high';
}

export interface ManimRenderResponse {
  jobId: string;
  status: 'queued' | 'rendering' | 'complete' | 'failed';
  videoUrl?: string;
  error?: string;
  duration?: number;
}

export interface RendererConfig {
  endpoint: string;
  timeout: number;
  maxRetries: number;
}

export class ManimRendererService {
  private config: RendererConfig;
  private activeJobs: Map<string, ManimRenderRequest>;

  constructor(config: Partial<RendererConfig> = {}) {
    this.config = {
      endpoint: config.endpoint || process.env.MANIM_RENDERER_URL || 'http://localhost:8080',
      timeout: config.timeout || 300000,
      maxRetries: config.maxRetries || 3,
    };
    this.activeJobs = new Map();
  }

  async submitRender(request: ManimRenderRequest): Promise<ManimRenderResponse> {
    logger.info('Submitting render job', {
      jobId: request.jobId,
      codeLength: request.code.length,
      format: request.outputFormat,
    });

    this.activeJobs.set(request.jobId, request);

    try {
      const response = await fetch(`${this.config.endpoint}/render`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          job_id: request.jobId,
          code: request.code,
          problem: request.problem,
          output_format: request.outputFormat || 'mp4',
          quality: request.quality || 'medium',
        }),
      });

      if (!response.ok) {
        const error = await response.text();
        throw new Error(`Renderer error: ${error}`);
      }

      const result = await response.json();

      logger.info('Render job submitted', {
        jobId: request.jobId,
        status: result.status,
      });

      return {
        jobId: request.jobId,
        status: 'queued',
      };
    } catch (error) {
      logger.error('Failed to submit render job', error as Error, { jobId: request.jobId });
      return {
        jobId: request.jobId,
        status: 'failed',
        error: (error as Error).message,
      };
    }
  }

  async getStatus(jobId: string): Promise<ManimRenderResponse> {
    const request = this.activeJobs.get(jobId);
    if (!request) {
      return {
        jobId,
        status: 'failed',
        error: 'Job not found',
      };
    }

    try {
      const response = await fetch(`${this.config.endpoint}/status/${jobId}`);

      if (!response.ok) {
        if (response.status === 404) {
          return {
            jobId,
            status: 'failed',
            error: 'Job not found',
          };
        }
        throw new Error(`Status check failed: ${response.status}`);
      }

      const result = await response.json();

      if (result.status === 'complete') {
        this.activeJobs.delete(jobId);
      }

      return {
        jobId,
        status: result.status,
        videoUrl: result.video_url,
        duration: result.duration,
      };
    } catch (error) {
      logger.error('Failed to get job status', error as Error, { jobId });
      return {
        jobId,
        status: 'failed',
        error: (error as Error).message,
      };
    }
  }

  async cancelRender(jobId: string): Promise<boolean> {
    try {
      const response = await fetch(`${this.config.endpoint}/cancel/${jobId}`, {
        method: 'POST',
      });

      if (response.ok) {
        this.activeJobs.delete(jobId);
        logger.info('Render job cancelled', { jobId });
        return true;
      }

      return false;
    } catch (error) {
      logger.error('Failed to cancel render job', error as Error, { jobId });
      return false;
    }
  }

  validateCode(code: string): { valid: boolean; error?: string } {
    if (!code || code.trim().length === 0) {
      return { valid: false, error: 'Code is empty' };
    }

    if (!code.includes('from manim import')) {
      return { valid: false, error: 'Code must import manim' };
    }

    if (!code.includes('class') || !code.includes('Scene')) {
      return { valid: false, error: 'Code must define a Scene class' };
    }

    const lines = code.split('\n');
    const indentedCode = lines.filter(line => line.startsWith(' ')).length;

    if (indentedCode < 2) {
      return { valid: false, error: 'Code must have proper indentation' };
    }

    return { valid: true };
  }
}

export const createRenderer = (config?: Partial<RendererConfig>): ManimRendererService => {
  return new ManimRendererService(config);
};
