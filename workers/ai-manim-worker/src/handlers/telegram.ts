import { Context, Hono } from 'hono';
import { SessionService } from '../services/session';
import { AIFallbackService } from '../services/fallback';
import { createLogger } from '../utils/logger';

const logger = createLogger({ level: 'info', component: 'telegram-handler' });

export interface TelegramUpdate {
  update_id: number;
  message?: {
    message_id: number;
    from?: {
      id: number;
      is_bot: boolean;
      first_name?: string;
      username?: string;
    };
    chat: {
      id: number;
      type: string;
    };
    text?: string;
    date: number;
  };
  callback_query?: {
    id: string;
    from: {
      id: number;
      is_bot: boolean;
      first_name?: string;
      username?: string;
    };
    message: {
      message_id: number;
      chat: {
        id: number;
      };
    };
    data: string;
  };
}

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

export class TelegramHandler {
  private sessionService: SessionService;
  private aiService: AIFallbackService;
  private app: Hono;

  constructor(sessionService: SessionService, aiService: AIFallbackService) {
    this.sessionService = sessionService;
    this.aiService = aiService;
    this.app = new Hono();
    this.setupRoutes();
  }

  private setupRoutes() {
    this.app.post('/webhook', async (c: Context) => {
      try {
        const update = await c.req.json<TelegramUpdate>();
        return this.handleUpdate(c, update);
      } catch (error) {
        logger.error('Failed to parse Telegram update', error as Error);
        return c.json({ ok: false, error: 'Invalid update format' }, 400);
      }
    });

    this.app.get('/health', async (c: Context) => {
      return c.json({ status: 'healthy', timestamp: Date.now() });
    });
  }

  private async handleUpdate(c: Context, update: TelegramUpdate): Promise<Response> {
    logger.debug('Received Telegram update', { updateId: update.update_id });

    if (update.message) {
      return this.handleMessage(c, update.message);
    }

    if (update.callback_query) {
      return this.handleCallbackQuery(c, update.callback_query);
    }

    return c.json({ ok: true });
  }

  private async handleMessage(c: Context, message: TelegramUpdate['message']): Promise<Response> {
    if (!message?.text) {
      return c.json({ ok: true });
    }

    const chatId = message.chat.id;
    const text = message.text;
    const userId = message.from?.id.toString() || 'anonymous';

    logger.info('Received message', { chatId, userId, textLength: text.length });

    if (text === '/start' || text === '/help') {
      return this.sendHelp(c, chatId);
    }

    if (text === '/status') {
      return this.sendStatus(c, chatId, userId);
    }

    if (text.startsWith('/')) {
      return this.sendUnknownCommand(c, chatId);
    }

    return this.handleProblemSubmission(c, chatId, userId, text);
  }

  private async handleCallbackQuery(c: Context, callback: NonNullable<TelegramUpdate['callback_query']>): Promise<Response> {
    const chatId = callback.message.chat.id;
    const data = callback.data;

    logger.info('Callback query', { chatId, data });

    return c.json({ ok: true });
  }

  private async sendHelp(c: Context, chatId: number): Promise<Response> {
    const helpText = `*AI Manim Video Generator*

Send me a mathematical problem or concept, and I'll generate an animated video explanation.

*Commands:*
• /start - Show this help message
• /status - Check your recent jobs

*Examples:*
• "Explain the Pythagorean theorem"
• "Show me how derivatives work"
• "Visualize a sine wave"

_Processing takes 1-5 minutes depending on complexity._`;

    return this.sendMessage(c, chatId, helpText, { parseMode: 'Markdown' });
  }

  private async sendStatus(c: Context, chatId: number, userId: string): Promise<Response> {
    const session = await this.sessionService.getOrCreateSession(chatId.toString());

    const statusText = `*Your Recent Jobs*

No recent jobs found.

Send me a problem to get started!`;

    return this.sendMessage(c, chatId, statusText, { parseMode: 'Markdown' });
  }

  private async sendUnknownCommand(c: Context, chatId: number): Promise<Response> {
    return this.sendMessage(c, chatId, 'Unknown command. Use /help for available commands.');
  }

  private async handleProblemSubmission(c: Context, chatId: number, userId: string, problem: string): Promise<Response> {
    const session = await this.sessionService.getOrCreateSession(chatId.toString());

    logger.info('Processing problem submission', {
      problemLength: problem.length,
    });

    const confirmationText = `🎬 *Processing Your Request*

Your problem has been queued for processing.

_Duration: 1-5 minutes_

I'll send you a link when the video is ready!`;

    await this.sendMessage(c, chatId, confirmationText, { parseMode: 'Markdown' });

    return c.json({ ok: true });
  }

  private async sendMessage(c: Context, chatId: number, text: string, options: Record<string, string> = {}): Promise<Response> {
    const payload: Record<string, unknown> = {
      chat_id: chatId,
      text,
      ...options,
    };

    const token = c.env.TELEGRAM_BOT_TOKEN;
    const response = await fetch(`https://api.telegram.org/bot${token}/sendMessage`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });

    if (!response.ok) {
      const errorText = await response.text();
      logger.error('Failed to send Telegram message', new Error(errorText), { chatId });
    }

    return c.json({ ok: response.ok });
  }

  getApp(): Hono {
    return this.app;
  }
}
