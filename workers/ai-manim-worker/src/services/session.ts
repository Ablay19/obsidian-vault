import type { Env, UserSession, ProcessingJob, VideoMetadata, ProcessingStatus, Platform } from "../types";
import { createLogger } from "../utils/logger";
import { JOB_TTL_SECONDS } from "../utils/constants";

const SESSION_PREFIX = "session:";
const JOB_PREFIX = "job:";
const SESSION_TTL_SECONDS = 7 * 24 * 60 * 60;

export class SessionService {
  private env: Env;
  private logger: ReturnType<typeof createLogger>;

  constructor(env: Env) {
    this.env = env;
    this.logger = createLogger({
      level: (env.LOG_LEVEL as "debug" | "info" | "warn" | "error") || "info",
      component: "session-service",
    });
  }

  async getOrCreateSession(platform: Platform, userId: string): Promise<UserSession> {
    const existingSession = await this.findSessionByPlatformUserId(platform, userId);
    if (existingSession) {
      await this.extendSession(existingSession.session_id);
      return existingSession;
    }

    const sessionId = crypto.randomUUID();
    const now = new Date().toISOString();

    const session: UserSession = {
      session_id: sessionId,
      created_at: now,
      last_activity: now,
      platform_primary: platform,
      video_history: [],
      total_submissions: 0,
      successful_generations: 0,
    };

    // Set platform-specific ID
    switch (platform) {
      case "telegram":
        session.telegram_chat_id = userId;
        break;
      case "whatsapp":
        session.whatsapp_phone_number = userId;
        break;
      case "web":
        session.web_session_token = userId;
        break;
    }

    await this.saveSession(session);
    this.logger.info("Created new session", { session_id: sessionId, platform, user_id: userId });

    return session;
  }

  // Backward compatibility method
  async getOrCreateSessionByTelegramChatId(chatId: string): Promise<UserSession> {
    return this.getOrCreateSession("telegram", chatId);
  }

  async findSessionByPlatformUserId(platform: Platform, userId: string): Promise<UserSession | null> {
    const key = `sessions:by_${platform}:${userId}`;
    const sessionId = await this.env.SESSIONS.get(key) as string | null;
    if (!sessionId) return null;

    return this.getSession(sessionId);
  }

  // Backward compatibility method
  async findSessionByChatId(chatId: string): Promise<UserSession | null> {
    return this.findSessionByPlatformUserId("telegram", chatId);
  }

  async getSession(sessionId: string): Promise<UserSession | null> {
    const session = await this.env.SESSIONS.get<UserSession>(`${SESSION_PREFIX}${sessionId}`, "json");
    return session;
  }

  async saveSession(session: UserSession): Promise<void> {
    const sessionKey = `${SESSION_PREFIX}${session.session_id}`;
    const indexOperations: Promise<void>[] = [
      this.env.SESSIONS.put(sessionKey, JSON.stringify(session), { expirationTtl: SESSION_TTL_SECONDS }),
    ];

    // Create indexes for all platform IDs that exist
    if (session.telegram_chat_id) {
      indexOperations.push(
        this.env.SESSIONS.put(`sessions:by_telegram:${session.telegram_chat_id}`, session.session_id, { expirationTtl: SESSION_TTL_SECONDS })
      );
    }
    if (session.whatsapp_phone_number) {
      indexOperations.push(
        this.env.SESSIONS.put(`sessions:by_whatsapp:${session.whatsapp_phone_number}`, session.session_id, { expirationTtl: SESSION_TTL_SECONDS })
      );
    }
    if (session.web_session_token) {
      indexOperations.push(
        this.env.SESSIONS.put(`sessions:by_web:${session.web_session_token}`, session.session_id, { expirationTtl: SESSION_TTL_SECONDS })
      );
    }

    await Promise.all(indexOperations);
    this.logger.debug("Session saved", { session_id: session.session_id });
  }

  async extendSession(sessionId: string): Promise<void> {
    const session = await this.getSession(sessionId);
    if (!session) {
      return;
    }

    session.last_activity = new Date().toISOString();
    await this.saveSession(session);
  }

  async incrementSubmission(sessionId: string): Promise<void> {
    const session = await this.getSession(sessionId);
    if (!session) {
      return;
    }

    session.total_submissions += 1;
    await this.saveSession(session);
  }

  async incrementSuccess(sessionId: string, renderDuration: number): Promise<void> {
    const session = await this.getSession(sessionId);
    if (!session) {
      return;
    }

    session.successful_generations += 1;
    await this.saveSession(session);
  }

  async createJob(sessionId: string, problemText: string, platform: Platform, submissionType: "problem" | "direct_code" = "problem"): Promise<ProcessingJob> {
    const jobId = crypto.randomUUID();
    const now = new Date().toISOString();

    const job: ProcessingJob = {
      job_id: jobId,
      session_id: sessionId,
      submission_type: submissionType,
      platform: platform,
      problem_text: submissionType === "problem" ? problemText : undefined,
      manim_code: submissionType === "direct_code" ? problemText : undefined,
      status: 'queued',
      created_at: now,
      started_at: undefined,
      completed_at: undefined,
    };

    await this.saveJob(job);
    await this.extendSession(sessionId);
    await this.incrementSubmission(sessionId);

    this.logger.info('Job created', { job_id: jobId, session_id: sessionId });

    return job;
  }

  async updateJobStatus(jobId: string, status: ProcessingStatus, data?: Partial<ProcessingJob>): Promise<void> {
    const job = await this.getJob(jobId);
    if (!job) {
      this.logger.warn('Job not found for update', { job_id: jobId });
      return;
    }

    job.status = status;
    if (status === 'ai_generating' && !job.started_at) {
      job.started_at = new Date().toISOString();
    }
    if (status === 'ready' || status === 'delivered' || status === 'failed') {
      job.completed_at = new Date().toISOString();
    }

    if (data) {
      Object.assign(job, data);
    }

    const ttl = JOB_TTL_SECONDS[status] || 3600;
    await this.env.SESSIONS.put(`${JOB_PREFIX}${jobId}`, JSON.stringify(job), {
      expirationTtl: ttl,
    });

    this.logger.debug('Job status updated', { job_id: jobId, status, ttl });
  }

  async getJob(jobId: string): Promise<ProcessingJob | null> {
    const job = await this.env.SESSIONS.get<ProcessingJob>(`${JOB_PREFIX}${jobId}`, 'json');
    return job;
  }

  async saveJob(job: ProcessingJob): Promise<void> {
    const ttl = JOB_TTL_SECONDS[job.status] || 3600;
    await this.env.SESSIONS.put(`${JOB_PREFIX}${job.job_id}`, JSON.stringify(job), {
      expirationTtl: ttl,
    });
  }

  async updateJobWithVideo(jobId: string, videoUrl: string, videoKey: string, expiresAt: string): Promise<void> {
    await this.updateJobStatus(jobId, 'ready', {
      video_url: videoUrl,
      video_key: videoKey,
      video_expires_at: expiresAt,
    });
  }

  async trackVideoAccess(jobId: string): Promise<void> {
    const job = await this.getJob(jobId);
    if (!job) {
      return;
    }

    await this.updateJobStatus(jobId, 'delivered', {
      completed_at: new Date().toISOString(),
    });

    const session = await this.getSession(job.session_id);
    if (session) {
      await this.addVideoToHistory(session.session_id, {
        job_id: job.job_id,
        problem_preview: (job.problem_text || job.manim_code || "").substring(0, 50),
        status: 'delivered',
        render_duration_seconds: job.render_duration_seconds,
      });
    }
  }

  async updateSessionHistory(sessionId: string, metadata: VideoMetadata): Promise<void> {
    await this.addVideoToHistory(sessionId, metadata);
  }

  async addVideoToHistory(sessionId: string, videoMetadata: {
    job_id: string;
    problem_preview: string;
    status: string;
    render_duration_seconds?: number;
  }): Promise<void> {
    const session = await this.getSession(sessionId);
    if (!session) {
      return;
    }

    session.video_history.push({
      job_id: videoMetadata.job_id,
      problem_preview: videoMetadata.problem_preview,
      status: videoMetadata.status as any,
      created_at: new Date().toISOString(),
      delivered_at: videoMetadata.status === "delivered" ? new Date().toISOString() : undefined,
      render_duration_seconds: videoMetadata.render_duration_seconds,
    });

    if (session.video_history.length > 100) {
      session.video_history = session.video_history.slice(-100);
    }

    await this.saveSession(session);
  }
}

export function createSessionService(env: Env): SessionService {
  return new SessionService(env);
}
