import { Hono } from 'hono';
import type { ProcessingStatus } from '../types';
import { createLogger } from '../utils/logger';

const logger = createLogger({ level: 'info', component: 'renderer-callback' });

export function createRendererCallbackHandler(app: Hono) {
  app.post('/webhook/renderer', async (c) => {
    try {
      const payload = await c.req.json();

      logger.info('Received renderer callback', {
        job_id: payload.job_id,
        status: payload.status,
      });

      const jobStatus: ProcessingStatus = mapRendererStatusToJobStatus(payload.status);

      c.header('X-Job-ID', payload.job_id);
      c.header('X-Job-Status', jobStatus);

      logger.info('Renderer callback processed', { job_id: payload.job_id });

      return c.json({ success: true, job_id: payload.job_id });
    } catch (error) {
      logger.error('Error processing renderer callback', error as Error);
      return c.json({ success: false, error: (error as Error).message }, 500);
    }
  });
}

function mapRendererStatusToJobStatus(
  rendererStatus: string
): ProcessingStatus {
  switch (rendererStatus) {
    case 'success':
      return 'ready';
    case 'error':
    case 'timeout':
      return 'failed';
    default:
      return 'ai_generating';
  }
}
