import { Context, Hono } from 'hono';
import { createLogger } from '../utils/logger';

const logger = createLogger({ level: 'info', component: 'debug-handler' });

export class DebugHandler {
  private app: Hono;

  constructor() {
    this.app = new Hono();
    this.setupRoutes();
  }

  private setupRoutes() {
    this.app.get('/env', async (c: Context) => {
      const env = c.env as Record<string, unknown>;
      const telegramSecret = env.TELEGRAM_SECRET as string || '';
      
      return c.json({
        telegramSecretLength: telegramSecret.length,
        telegramSecretPrefix: telegramSecret.substring(0, 5) || 'empty',
        telegramSecret: telegramSecret ? '***' : 'empty',
        allEnvKeys: Object.keys(env).filter(k => !k.startsWith('__') && !k.includes('_JS') && !k.includes('$$')),
      });
    });

    this.app.get('/health', async (c: Context) => {
      return c.json({ status: 'healthy', timestamp: Date.now() });
    });
  }

  getApp(): Hono {
    return this.app;
  }
}
