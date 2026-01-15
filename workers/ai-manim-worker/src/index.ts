import type { ExecutionContext } from '@cloudflare/workers-types';
import { Hono } from 'hono';
import { cors } from 'hono/cors';
import type { Env } from './types';
import { createLogger } from './utils/logger';
import { SessionService } from './services/session';
import { AIFallbackService } from './services/fallback';
import { TelegramHandler } from './handlers/telegram';
import { VideoHandler } from './handlers/video';
import { DebugHandler } from './handlers/debug';

export interface ProcessingJob {
  id: string;
  userId: string;
  chatId: number;
  problem: string;
  status: 'queued' | 'ai_generating' | 'code_validating' | 'rendering' | 'uploading' | 'ready' | 'delivered' | 'failed';
  createdAt: number;
  updatedAt: number;
  error?: string;
  videoUrl?: string;
}

export default {
  async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    const url = new URL(request.url);
    const logger = createLogger({
      level: (env.LOG_LEVEL as 'debug' | 'info' | 'warn' | 'error') || 'info',
      component: 'ai-manim-worker'
    });

    logger.info('Incoming request', {
      method: request.method,
      url: url.pathname,
    });

    try {
      const app = createApp(env, logger);
      return app.fetch(request, env, ctx);
    } catch (error) {
      logger.error('Request failed', error as Error, {
        method: request.method,
        url: url.pathname,
      });

      return new Response(
        JSON.stringify({
          status: 'error',
          message: (error as Error).message,
        }),
        { status: 500, headers: { 'Content-Type': 'application/json' } }
      );
    }
  },
};

function createApp(env: Env, logger: ReturnType<typeof createLogger>): Hono {
  const app = new Hono();

  app.use('/*', cors({
    origin: (origin) => origin,
    allowMethods: ['GET', 'POST', 'OPTIONS'],
    allowHeaders: ['Content-Type', 'Authorization', 'X-Telegram-Bot-Api-Secret-Token'],
  }));

  const sessionService = new SessionService(env);
  const aiService = new AIFallbackService(env);

  const telegramHandler = new TelegramHandler(sessionService, aiService, env.TELEGRAM_SECRET);
  const videoHandler = new VideoHandler();
  const debugHandler = new DebugHandler();

  app.route('/telegram', telegramHandler.getApp());
  app.route('/video', videoHandler.getApp());
  app.route('/debug', debugHandler.getApp());

  app.get('/health', async (c) => {
    return c.json({
      status: 'healthy',
      version: '1.0.0',
      timestamp: new Date().toISOString(),
      providers: {
        cloudflare: 'configured',
        groq: 'configured',
        huggingface: 'configured',
      },
    });
  });

  app.get('/ready', async (c) => {
    return c.json({
      ready: true,
      timestamp: new Date().toISOString(),
    });
  });

  return app;
}

export { createApp };
