import { Context, Hono } from 'hono';
import { createLogger } from '../utils/logger';
import { SessionService } from '../services/session';
import { ManimRendererService, ManimRenderRequest } from '../services/manim';
import type { Env, DirectCodeSubmission, Platform } from '../types';

const logger = createLogger({ level: 'info', component: 'code-handler' });

export class CodeHandler {
  private app: Hono;
  private sessionService: SessionService;
  private manimService: ManimRendererService;

  constructor(env: Env) {
    this.app = new Hono();
    this.sessionService = new SessionService(env);
    this.manimService = new ManimRendererService({});
    this.setupRoutes();
  }

  private setupRoutes() {
    // POST /api/v1/code - Submit direct Manim code
    this.app.post('/code', async (c: Context) => {
      try {
        const body = await c.req.json<DirectCodeSubmission>();

        // For now, assume web platform and anonymous user
        // In production, this would come from authentication
        const platform: Platform = 'web';
        const userId = 'anonymous-web-user';

        return this.handleDirectCodeSubmission(c, body, platform, userId);
      } catch (error) {
        logger.error('Failed to parse code submission', error as Error);
        return c.json({ error: 'Invalid request format' }, 400);
      }
    });

    // GET /api/v1/code - Validate Manim code syntax
    this.app.get('/code', async (c: Context) => {
      const code = c.req.query('code');
      if (!code) {
        return c.json({ error: 'Code parameter required' }, 400);
      }

      return this.handleCodeValidation(c, code);
    });
  }

  private async handleDirectCodeSubmission(
    c: Context,
    submission: DirectCodeSubmission,
    platform: Platform,
    userId: string
  ) {
    try {
      logger.info('Processing direct code submission', {
        platform,
        user_id: userId,
        code_length: submission.code.length,
        quality: submission.options?.quality || 'medium',
        format: submission.options?.format || 'mp4'
      });

      // Get or create session
      const session = await this.sessionService.getOrCreateSession(platform, userId);

      // Create job with direct code submission type
      const job = await this.sessionService.createJob(
        session.session_id,
        submission.code,
        platform,
        'direct_code'
      );

      // Validate code syntax (basic check)
      const validation = await this.validateManimCode(submission.code);
      if (!validation.valid) {
        logger.warn('Code validation failed', { job_id: job.job_id, errors: validation.errors });
        await this.sessionService.updateJobStatus(job.job_id, 'failed', { last_error: validation.errors.join('; ') });
        return c.json({
          error: 'Code validation failed',
          details: validation.errors
        }, 400);
      }

      // Submit to renderer
      const renderRequest: ManimRenderRequest = {
        jobId: job.job_id,
        code: submission.code,
        problem: 'Direct Manim code submission',
        outputFormat: submission.options?.format === 'webm' ? 'webm' : 'mp4',
        quality: submission.options?.quality || 'medium'
      };

      const renderResult = await this.manimService.submitRender(renderRequest);

      if (renderResult.status === 'failed') {
        logger.error('Render submission failed', undefined, { job_id: job.job_id, render_error: renderResult.error });
        await this.sessionService.updateJobStatus(job.job_id, 'failed', { last_error: renderResult.error });
        return c.json({ error: renderResult.error }, 500);
      }

      // Update job status to rendering
      await this.sessionService.updateJobStatus(job.job_id, 'rendering');

      logger.info('Direct code submission successful', { job_id: job.job_id });

      return c.json({
        job_id: job.job_id,
        status: 'queued',
        message: 'Code submitted for rendering',
        estimated_time: '2-5 minutes'
      }, 202);

    } catch (error) {
      logger.error('Direct code submission failed', error as Error);
      return c.json({ error: 'Internal server error' }, 500);
    }
  }

  private async handleCodeValidation(c: Context, code: string) {
    try {
      const validation = await this.validateManimCode(code);

      return c.json({
        valid: validation.valid,
        errors: validation.errors,
        warnings: validation.warnings
      });

    } catch (error) {
      logger.error('Code validation failed', error as Error);
      return c.json({ error: 'Validation service unavailable' }, 500);
    }
  }

  private async validateManimCode(code: string): Promise<{
    valid: boolean;
    errors: string[];
    warnings: string[];
  }> {
    const errors: string[] = [];
    const warnings: string[] = [];

    // Basic syntax checks
    if (!code.includes('from manim import')) {
      errors.push('Code must import from manim');
    }

    if (!code.includes('class ') || !code.includes('Scene')) {
      errors.push('Code must define a Scene class');
    }

    // Check for construct method
    if (!code.includes('def construct')) {
      errors.push('Scene class must have a construct method');
    }

    // Check code length
    if (code.length < 50) {
      warnings.push('Code seems very short for a meaningful animation');
    }

    if (code.length > 50000) {
      errors.push('Code exceeds maximum allowed length (50KB)');
    }

    // Check for potentially dangerous operations
    const dangerousPatterns = [
      'import os',
      'import sys',
      'import subprocess',
      '__import__',
      'eval(',
      'exec('
    ];

    for (const pattern of dangerousPatterns) {
      if (code.includes(pattern)) {
        errors.push(`Potentially unsafe operation detected: ${pattern}`);
      }
    }

    return {
      valid: errors.length === 0,
      errors,
      warnings
    };
  }

  public getRouter(): Hono {
    return this.app;
  }
}