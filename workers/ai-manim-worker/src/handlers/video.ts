import { Context, Hono } from 'hono';
import { createLogger } from '../utils/logger';

const logger = createLogger({ level: 'info', component: 'video-handler' });

export interface VideoRenderRequest {
  jobId: string;
  manimCode: string;
  problem: string;
  userId: string;
  chatId: number;
}

export interface VideoRenderResponse {
  jobId: string;
  status: string;
  videoUrl?: string;
  error?: string;
}

export class VideoHandler {
  private app: Hono;

  constructor() {
    this.app = new Hono();
    this.setupRoutes();
  }

  private setupRoutes() {
    this.app.post('/submit', async (c: Context) => {
      try {
        const body = await c.req.json<VideoRenderRequest>();
        return this.handleSubmit(c, body);
      } catch (error) {
        logger.error('Failed to parse video request', error as Error);
        return c.json({ error: 'Invalid request format' }, 400);
      }
    });

    this.app.get('/status/:jobId', async (c: Context) => {
      const jobId = c.req.param('jobId');
      return this.handleStatus(c, jobId);
    });

    this.app.post('/complete', async (c: Context) => {
      try {
        const body = await c.req.json();
        return this.handleComplete(c, body);
      } catch (error) {
        logger.error('Failed to parse completion callback', error as Error);
        return c.json({ error: 'Invalid request format' }, 400);
      }
    });

    this.app.get('/health', async (c: Context) => {
      return c.json({ status: 'healthy', timestamp: Date.now() });
    });
  }

  private async handleSubmit(c: Context, request: VideoRenderRequest): Promise<Response> {
    logger.info('Submitting video render', {
      jobId: request.jobId,
      codeLength: request.manimCode.length,
    });

    return c.json({
      jobId: request.jobId,
      status: 'queued',
      message: 'Video render job queued',
      estimatedTime: '2-5 minutes',
    });
  }

  private async handleStatus(c: Context, jobId: string): Promise<Response> {
    logger.debug('Getting job status', { jobId });

    return c.json({
      jobId,
      status: 'processing',
      progress: 0,
      message: 'Video is being rendered',
    });
  }

  private async handleComplete(c: Context, body: Record<string, unknown>): Promise<Response> {
    const { jobId, videoUrl, error } = body;

    if (error) {
      logger.error(`Video render failed for job ${jobId}: ${String(error)}`);
      return c.json({ jobId, status: 'failed', error: String(error) });
    }

    logger.info('Video render complete', { jobId, videoUrl });

    return c.json({
      jobId,
      status: 'ready',
      videoUrl,
    });
  }

  getApp(): Hono {
    return this.app;
  }
}
