import { Context, Hono } from 'hono';
import { createLogger } from '../utils/logger';
import { SessionService } from '../services/session';
import { AIFallbackService } from '../services/fallback';
import { ManimRendererService } from '../services/manim';
import type { Env, WhatsAppMessage, Platform } from '../types';

const logger = createLogger({ level: 'info', component: 'whatsapp-handler' });

export class WhatsAppHandler {
  private app: Hono;
  private sessionService: SessionService;
  private aiService: AIFallbackService;
  private manimService: ManimRendererService;
  private webhookSecret: string;

  constructor(
    sessionService: SessionService,
    aiService: AIFallbackService,
    manimService: ManimRendererService,
    webhookSecret: string
  ) {
    this.app = new Hono();
    this.sessionService = sessionService;
    this.aiService = aiService;
    this.manimService = manimService;
    this.webhookSecret = webhookSecret;
    this.setupRoutes();
  }

  private setupRoutes() {
    // WhatsApp webhook endpoint
    this.app.post('/webhook/whatsapp', async (c: Context) => {
      return this.handleWhatsAppWebhook(c);
    });

    // WhatsApp webhook verification (GET request for verification)
    this.app.get('/webhook/whatsapp', async (c: Context) => {
      return this.handleWhatsAppVerification(c);
    });
  }

  private async handleWhatsAppVerification(c: Context): Promise<Response> {
    const mode = c.req.query('hub.mode');
    const token = c.req.query('hub.verify_token');
    const challenge = c.req.query('hub.challenge');

    // Verify the webhook token
    if (mode === 'subscribe' && token === this.webhookSecret) {
      logger.info('WhatsApp webhook verified successfully');
      return c.text(challenge || '', 200);
    }

    logger.warn('WhatsApp webhook verification failed', { mode, token_provided: !!token });
    return c.text('Verification failed', 403);
  }

  private async handleWhatsAppWebhook(c: Context): Promise<Response> {
    try {
      // Validate webhook signature if provided
      const signature = c.req.header('X-Hub-Signature-256');
      if (signature && !this.validateWhatsAppSignature(await c.req.text(), signature)) {
        logger.warn('Invalid WhatsApp webhook signature');
        return c.json({ error: 'Invalid signature' }, 401);
      }

      // Parse the webhook payload
      const body = await c.req.json();
      const messages = this.extractMessagesFromWebhook(body);

      // Process each message
      for (const message of messages) {
        await this.processWhatsAppMessage(message);
      }

      return c.json({ status: 'ok' }, 200);

    } catch (error) {
      logger.error('WhatsApp webhook processing failed', error as Error);
      return c.json({ error: 'Processing failed' }, 500);
    }
  }

  private validateWhatsAppSignature(payload: string, signature: string): boolean {
    // WhatsApp uses SHA256 HMAC with the webhook secret
    // For now, we'll do basic validation - in production, implement proper HMAC verification
    return signature.startsWith('sha256=');
  }

  private extractMessagesFromWebhook(webhookData: any): WhatsAppMessage[] {
    const messages: WhatsAppMessage[] = [];

    try {
      if (webhookData.entry) {
        for (const entry of webhookData.entry) {
          if (entry.changes) {
            for (const change of entry.changes) {
              if (change.value && change.value.messages) {
                for (const msg of change.value.messages) {
                  const whatsappMessage: WhatsAppMessage = {
                    id: msg.id,
                    from: msg.from,
                    type: msg.type,
                    timestamp: msg.timestamp,
                  };

                  if (msg.text) {
                    whatsappMessage.text = {
                      body: msg.text.body
                    };
                  }

                  if (msg.image) {
                    whatsappMessage.image = {
                      id: msg.image.id,
                      mime_type: msg.image.mime_type,
                      sha256: msg.image.sha256
                    };
                  }

                  messages.push(whatsappMessage);
                }
              }
            }
          }
        }
      }
    } catch (error) {
      logger.error('Failed to extract messages from webhook', error as Error);
    }

    return messages;
  }

  private async processWhatsAppMessage(message: WhatsAppMessage): Promise<void> {
    try {
      logger.info('Processing WhatsApp message', {
        message_id: message.id,
        from: message.from,
        type: message.type
      });

      // Only process text messages for now
      if (message.type !== 'text' || !message.text) {
        logger.info('Ignoring non-text message', { type: message.type });
        return;
      }

      const userId = message.from;
      const content = message.text.body.trim();

      // Detect if this is Manim code or a problem description
      const isManimCode = this.detectManimCode(content);

      if (isManimCode) {
        // Handle direct code submission
        await this.handleWhatsAppCodeSubmission(userId, content);
      } else {
        // Handle problem submission
        await this.handleWhatsAppProblemSubmission(userId, content);
      }

    } catch (error) {
      logger.error('Failed to process WhatsApp message', error as Error, {
        message_id: message.id,
        from: message.from
      });
    }
  }

  private detectManimCode(content: string): boolean {
    // Check for Manim code patterns
    return content.includes('from manim import') &&
           (content.includes('class ') && content.includes('Scene'));
  }

  private async handleWhatsAppProblemSubmission(userId: string, problem: string): Promise<void> {
    try {
      logger.info('Processing WhatsApp problem submission', { user_id: userId });

      // Get or create session for WhatsApp user
      const session = await this.sessionService.getOrCreateSession("whatsapp", userId);

      // Create job for AI processing
      const job = await this.sessionService.createJob(session.session_id, problem, "whatsapp", "problem");

      // Send confirmation
      await this.sendWhatsAppMessage(userId, `🎬 Processing your problem: "${problem.substring(0, 50)}..."\n\nJob ID: ${job.job_id}\n\nThis may take 2-5 minutes.`);

      // Process asynchronously (would use execution context in Cloudflare Workers)
      setTimeout(() => this.processWhatsAppProblemAsync(userId, job.job_id, problem), 0);

    } catch (error) {
      logger.error('WhatsApp problem submission failed', error as Error);
      await this.sendWhatsAppMessage(userId, '❌ Sorry, there was an error processing your request. Please try again.');
    }
  }

  private async handleWhatsAppCodeSubmission(userId: string, code: string): Promise<void> {
    try {
      logger.info('Processing WhatsApp code submission', { user_id: userId });

      // Get or create session for WhatsApp user
      const session = await this.sessionService.getOrCreateSession("whatsapp", userId);

      // Create job for direct code processing
      const job = await this.sessionService.createJob(session.session_id, code, "whatsapp", "direct_code");

      // Send confirmation
      await this.sendWhatsAppMessage(userId, `🎬 Processing your Manim code...\n\nJob ID: ${job.job_id}\n\nThis may take 2-5 minutes.`);

       // Process asynchronously
       setTimeout(() => this.processWhatsAppCodeAsync(userId, job.job_id, code), 0);

    } catch (error) {
      logger.error('WhatsApp code submission failed', error as Error);
      await this.sendWhatsAppMessage(userId, '❌ Sorry, there was an error processing your code. Please check the syntax and try again.');
    }
  }

  private async processWhatsAppProblemAsync(userId: string, jobId: string, problem: string): Promise<void> {
    try {
      // Generate Manim code using AI
      const manimCode = await this.aiService.generateManimCode(problem);

      // Submit for rendering
      const renderRequest = {
        jobId,
        code: manimCode,
        problem,
        outputFormat: 'mp4' as const,
        quality: 'medium' as const
      };

      const renderResult = await this.manimService.submitRender(renderRequest);

      if (renderResult.status === 'failed') {
        await this.sessionService.updateJobStatus(jobId, 'failed', { last_error: renderResult.error });
        await this.sendWhatsAppMessage(userId, `❌ Failed to render video: ${renderResult.error}`);
        return;
      }

      // Update status and wait for completion
      await this.sessionService.updateJobStatus(jobId, 'rendering');

      // Poll for completion (simplified - in production use webhooks)
      await this.pollWhatsAppJobStatus(userId, jobId);

    } catch (error) {
      logger.error('Async WhatsApp problem processing failed', error as Error);
      await this.sendWhatsAppMessage(userId, '❌ An error occurred during processing. Please try again.');
    }
  }

  private async processWhatsAppCodeAsync(userId: string, jobId: string, code: string): Promise<void> {
    try {
      // Validate code syntax
      const validation = this.validateManimCode(code);
      if (!validation.valid) {
        await this.sessionService.updateJobStatus(jobId, 'failed', { last_error: validation.errors.join('; ') });
        await this.sendWhatsAppMessage(userId, `❌ Code validation failed:\n${validation.errors.join('\n')}`);
        return;
      }

      // Submit for rendering
      const renderRequest = {
        jobId,
        code,
        problem: 'Direct Manim code submission via WhatsApp',
        outputFormat: 'mp4' as const,
        quality: 'medium' as const
      };

      const renderResult = await this.manimService.submitRender(renderRequest);

      if (renderResult.status === 'failed') {
        await this.sessionService.updateJobStatus(jobId, 'failed', { last_error: renderResult.error });
        await this.sendWhatsAppMessage(userId, `❌ Failed to render video: ${renderResult.error}`);
        return;
      }

      // Update status and wait for completion
      await this.sessionService.updateJobStatus(jobId, 'rendering');

      // Poll for completion
      await this.pollWhatsAppJobStatus(userId, jobId);

    } catch (error) {
      logger.error('Async WhatsApp code processing failed', error as Error);
      await this.sendWhatsAppMessage(userId, '❌ An error occurred during processing. Please try again.');
    }
  }

  private async pollWhatsAppJobStatus(userId: string, jobId: string): Promise<void> {
    const maxPolls = 40; // 20 minutes max
    let pollCount = 0;

    while (pollCount < maxPolls) {
      await new Promise(resolve => setTimeout(resolve, 30000)); // 30 seconds

      const job = await this.sessionService.getJob(jobId);
      const status = job?.status || 'unknown';
      pollCount++;

      if (status === 'ready') {
        // Get video URL
        const job = await this.sessionService.getJob(jobId);
        if (job?.video_url) {
          await this.sendWhatsAppMessage(userId, `✅ Your video is ready!\n\n🎬 ${job.video_url}`);
          await this.sessionService.updateJobStatus(jobId, 'delivered');
        }
        return;
      } else if (status === 'failed') {
        await this.sendWhatsAppMessage(userId, '❌ Video generation failed. Please check your input and try again.');
        return;
      }
    }

    await this.sendWhatsAppMessage(userId, '⏰ Video generation is taking longer than expected. You can check the status later on the web dashboard.');
  }

  private async sendWhatsAppMessage(to: string, message: string): Promise<void> {
    // TODO: Implement WhatsApp Business API call
    // For now, just log the message
    logger.info('WhatsApp message would be sent', { to, message_length: message.length });
    console.log(`WhatsApp → ${to}: ${message}`);
  }

  private validateManimCode(code: string): { valid: boolean; errors: string[] } {
    const errors: string[] = [];

    if (!code.includes('from manim import')) {
      errors.push('Code must import from manim');
    }

    if (!code.includes('class ') || !code.includes('Scene')) {
      errors.push('Code must define a Scene class');
    }

    if (!code.includes('def construct')) {
      errors.push('Scene class must have a construct method');
    }

    return {
      valid: errors.length === 0,
      errors
    };
  }

  public getApp(): Hono {
    return this.app;
  }
}