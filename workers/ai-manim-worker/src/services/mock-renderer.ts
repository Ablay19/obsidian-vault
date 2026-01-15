import { createLogger } from '../utils/logger';

const logger = createLogger({ level: 'info', component: 'mock-renderer' });

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

export class MockRendererService {
  private config: RendererConfig;
  private jobs: Map<string, ManimRenderRequest>;

  constructor(config: Partial<RendererConfig> = {}) {
    this.config = {
      endpoint: config.endpoint || 'http://localhost:8080',
      timeout: config.timeout || 30000,
      maxRetries: config.maxRetries || 3,
    };
    this.jobs = new Map();
  }

  async submitRender(request: ManimRenderRequest): Promise<ManimRenderResponse> {
    logger.info('Mock: Submitting render job', {
      jobId: request.jobId,
      codeLength: request.code.length,
    });

    this.jobs.set(request.jobId, request);

    return {
      jobId: request.jobId,
      status: 'queued',
    };
  }

  async getStatus(jobId: string): Promise<ManimRenderResponse> {
    const request = this.jobs.get(jobId);
    
    if (!request) {
      return {
        jobId,
        status: 'failed',
        error: 'Job not found',
      };
    }

    return {
      jobId,
      status: 'complete',
      videoUrl: `https://example.com/videos/${jobId}.mp4`,
      duration: 15,
    };
  }

  async cancelRender(jobId: string): Promise<boolean> {
    return this.jobs.delete(jobId);
  }

  validateCode(code: string): { valid: boolean; error?: string } {
    if (!code || code.trim().length === 0) {
      return { valid: false, error: 'Code is empty' };
    }

    if (!code.includes('from manim import') && !code.includes('import manim')) {
      return { valid: false, error: 'Code must import manim' };
    }

    if (!code.includes('class') || !code.includes('Scene')) {
      return { valid: false, error: 'Code must define a Scene class' };
    }

    return { valid: true };
  }
}

export const createMockRenderer = (config?: Partial<RendererConfig>): MockRendererService => {
  return new MockRendererService(config);
};
