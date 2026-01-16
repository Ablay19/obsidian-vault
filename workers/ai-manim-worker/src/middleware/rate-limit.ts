import { createLogger } from '../utils/logger';
import { RATE_LIMIT_PER_MINUTE, RATE_LIMIT_PER_HOUR } from '../utils/constants';

const logger = createLogger({ level: 'info', component: 'rate-limit' });

interface RateLimitEntry {
  count: number;
  windowStart: number;
  hourlyCount: number;
  hourlyWindowStart: number;
}

const rateLimits = new Map<string, RateLimitEntry>();
const MINUTE_MS = 60000;
const HOUR_MS = 3600000;

export class RateLimiter {
  private maxPerMinute: number;
  private maxPerHour: number;

  constructor(config?: { perMinute?: number; perHour?: number }) {
    this.maxPerMinute = config?.perMinute || RATE_LIMIT_PER_MINUTE;
    this.maxPerHour = config?.perHour || RATE_LIMIT_PER_HOUR;
  }

  checkRateLimit(identifier: string): { allowed: boolean; retryAfter?: number } {
    const now = Date.now();
    const entry = rateLimits.get(identifier);

    if (!entry) {
      rateLimits.set(identifier, {
        count: 1,
        windowStart: now,
        hourlyCount: 1,
        hourlyWindowStart: now,
      });

      return { allowed: true };
    }

    const elapsedMinute = now - entry.windowStart;
    const elapsedHour = now - entry.hourlyWindowStart;

    if (elapsedMinute > MINUTE_MS) {
      entry.count = 1;
      entry.windowStart = now;
    } else {
      entry.count++;
    }

    if (elapsedHour > HOUR_MS) {
      entry.hourlyCount = 1;
      entry.hourlyWindowStart = now;
    } else {
      entry.hourlyCount++;
    }

    if (entry.count > this.maxPerMinute) {
      const retryAfter = entry.windowStart + MINUTE_MS - now;
      logger.warn('Rate limit exceeded (minute)', { identifier, count: entry.count });

      return { allowed: false, retryAfter: Math.ceil(retryAfter / 1000) };
    }

    if (entry.hourlyCount > this.maxPerHour) {
      const retryAfter = entry.hourlyWindowStart + HOUR_MS - now;
      logger.warn('Rate limit exceeded (hour)', { identifier, count: entry.hourlyCount });

      return { allowed: false, retryAfter: Math.ceil(retryAfter / 1000) };
    }

    rateLimits.set(identifier, entry);

    return { allowed: true };
  }

  reset(identifier: string): void {
    rateLimits.delete(identifier);
    logger.info('Rate limit reset', { identifier });
  }

  cleanup(): void {
    const now = Date.now();
    for (const [key, entry] of rateLimits.entries()) {
      if (now - entry.windowStart > MINUTE_MS * 60) {
        rateLimits.delete(key);
      }
    }
  }
}

export const createRateLimiter = (config?: { perMinute?: number; perHour?: number }): RateLimiter => {
  return new RateLimiter(config);
};
