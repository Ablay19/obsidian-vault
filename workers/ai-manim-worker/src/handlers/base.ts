import { Context, Hono } from 'hono';
import { createLogger } from '../utils/logger';
import { SessionService } from '../services/session';
import { AIFallbackService } from '../services/fallback';
import { ManimRendererService } from '../services/manim';
import type { Env, Platform, UserSession, ProcessingJob } from '../types';

const logger = createLogger({ level: 'info', component: 'base-message-handler' });

export interface MessageContext {
  platform: Platform;
  userId: string;
  content: string;
  messageId?: string;
  timestamp?: string;
}

export interface MessageResponse {
  success: boolean;
  message?: string;
  jobId?: string;
  error?: string;
}

export abstract class BaseMessageHandler {
  protected app: Hono;
  protected sessionService: SessionService;
  protected aiService: AIFallbackService;
  protected manimService: ManimRendererService;
  protected platform: Platform;

  constructor(
    sessionService: SessionService,
    aiService: AIFallbackService,
    manimService: ManimRendererService,
    platform: Platform
  ) {
    this.app = new Hono();
    this.sessionService = sessionService;
    this.aiService = aiService;
    this.manimService = manimService;
    this.platform = platform;
    this.setupRoutes();
  }

  protected abstract setupRoutes(): void;

  protected async handleMessage(context: MessageContext): Promise<MessageResponse> {
    try {
      logger.info(`Processing ${this.platform} message`, {
        user_id: context.userId,
        content_length: context.content.length,
        message_id: context.messageId
      });

      // Get or create session
      const session = await this.sessionService.getOrCreateSession(this.platform, context.userId);

      // Detect message type
      const isManimCode = this.detectManimCode(context.content);

      if (isManimCode) {
        return this.handleCodeSubmission(session, context);
      } else {
        return this.handleProblemSubmission(session, context);
      }

    } catch (error) {
      logger.error(`${this.platform} message processing failed`, error as Error);
      return {
        success: false,
        error: 'An error occurred processing your message. Please try again.'
      };
    }
  }

  private async handleProblemSubmission(session: UserSession, context: MessageContext): Promise<MessageResponse> {
    // Create job for AI processing
    const job = await this.sessionService.createJob(session.session_id, context.content, this.platform, "problem");

    // Send confirmation
    const confirmationMessage = await this.sendConfirmation(session.session_id, job.job_id, context.content);

    // Process asynchronously
    this.processProblemAsync(session.session_id, job.job_id, context.content);

    return {
      success: true,
      message: confirmationMessage,
      jobId: job.job_id
    };
  }

  private async handleCodeSubmission(session: UserSession, context: MessageContext): Promise<MessageResponse> {
    // Validate code
    const validation = this.validateManimCode(context.content);
    if (!validation.valid) {
      return {
        success: false,
        error: `Code validation failed: ${validation.errors.join(', ')}`
      };
    }

    // Create job for direct code processing
    const job = await this.sessionService.createJob(session.session_id, context.content, this.platform, "direct_code");

    // Send confirmation
    const confirmationMessage = await this.sendCodeConfirmation(session.session_id, job.job_id);

    // Process asynchronously
    this.processCodeAsync(session.session_id, job.job_id, context.content);

    return {
      success: true,
      message: confirmationMessage,
      jobId: job.job_id
    };
  }

  protected abstract sendConfirmation(sessionId: string, jobId: string, content: string): Promise<string>;
  protected abstract sendCodeConfirmation(sessionId: string, jobId: string): Promise<string>;
  protected abstract sendMessage(userId: string, message: string): Promise<void>;
  protected abstract sendVideoLink(sessionId: string, jobId: string, videoUrl: string): Promise<void>;
  protected abstract sendError(sessionId: string, error: string): Promise<void>;

  private detectManimCode(content: string): boolean {
    return content.includes('from manim import') &&
           (content.includes('class ') && content.includes('Scene'));
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

  private async processProblemAsync(sessionId: string, jobId: string, problem: string): Promise<void> {
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
        await this.sendError(sessionId, `❌ Failed to render video: ${renderResult.error}`);
        return;
      }

      // Update status and wait for completion
      await this.sessionService.updateJobStatus(jobId, 'rendering');

      // Poll for completion
      await this.pollJobCompletion(sessionId, jobId);

    } catch (error) {
      logger.error('Async problem processing failed', error as Error);
      await this.sendError(sessionId, '❌ An error occurred during processing. Please try again.');
    }
  }

  private async processCodeAsync(sessionId: string, jobId: string, code: string): Promise<void> {
    try {
      // Submit for rendering
      const renderRequest = {
        jobId,
        code,
        problem: 'Direct Manim code submission',
        outputFormat: 'mp4' as const,
        quality: 'medium' as const
      };

      const renderResult = await this.manimService.submitRender(renderRequest);

      if (renderResult.status === 'failed') {
        await this.sessionService.updateJobStatus(jobId, 'failed', { last_error: renderResult.error });
        await this.sendError(sessionId, `❌ Failed to render video: ${renderResult.error}`);
        return;
      }

      // Update status and wait for completion
      await this.sessionService.updateJobStatus(jobId, 'rendering');

      // Poll for completion
      await this.pollJobCompletion(sessionId, jobId);

    } catch (error) {
      logger.error('Async code processing failed', error as Error);
      await this.sendError(sessionId, '❌ An error occurred during processing. Please try again.');
    }
  }

  private async pollJobCompletion(sessionId: string, jobId: string): Promise<void> {
    const maxPolls = 40; // 20 minutes max
    let pollCount = 0;

    while (pollCount < maxPolls) {
      await new Promise(resolve => setTimeout(resolve, 30000)); // 30 seconds

      const job = await this.sessionService.getJob(jobId);
      const status = job?.status || 'unknown';
      pollCount++;

      if (status === 'ready') {
        // Get video URL and send to user
        if (job?.video_url) {
          await this.sendVideoLink(sessionId, jobId, job.video_url);
          await this.sessionService.updateJobStatus(jobId, 'delivered');
        }
        return;
      } else if (status === 'failed') {
        await this.sendError(sessionId, '❌ Video generation failed. Please check your input and try again.');
        return;
      }
    }

    await this.sendError(sessionId, '⏰ Video generation is taking longer than expected. You can check the status later on the web dashboard.');
  }

  public getApp(): Hono {
    return this.app;
  }
}