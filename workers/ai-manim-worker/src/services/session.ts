import type { Env, UserSession } from "../types";
import { createLogger } from "../utils/logger";

const SESSION_PREFIX = "session:";
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

  async getOrCreateSession(chatId: string): Promise<UserSession> {
    const existingSession = await this.findSessionByChatId(chatId);
    if (existingSession) {
      await this.extendSession(existingSession.session_id);
      return existingSession;
    }

    const sessionId = crypto.randomUUID();
    const now = new Date().toISOString();

    const session: UserSession = {
      session_id: sessionId,
      telegram_chat_id: chatId,
      created_at: now,
      last_activity: now,
      video_history: [],
      total_submissions: 0,
      successful_generations: 0,
    };

    await this.saveSession(session);
    this.logger.info("Created new session", { session_id: sessionId, chat_id: chatId });

    return session;
  }

  async findSessionByChatId(chatId: string): Promise<UserSession | null> {
    const key = `sessions:by_chat:${chatId}`;
    const sessionId = await this.env.KV.get(key) as string | null;

    if (!sessionId) {
      return null;
    }

    return this.getSession(sessionId);
  }

  async getSession(sessionId: string): Promise<UserSession | null> {
    const session = await this.env.KV.get<UserSession>(`${SESSION_PREFIX}${sessionId}`, "json");
    return session;
  }

  async saveSession(session: UserSession): Promise<void> {
    const sessionKey = `${SESSION_PREFIX}${session.session_id}`;
    const chatIndexKey = `sessions:by_chat:${session.telegram_chat_id}`;

    await Promise.all([
      this.env.KV.put(sessionKey, JSON.stringify(session), { expirationTtl: SESSION_TTL_SECONDS }),
      this.env.KV.put(chatIndexKey, session.session_id, { expirationTtl: SESSION_TTL_SECONDS }),
    ]);

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
