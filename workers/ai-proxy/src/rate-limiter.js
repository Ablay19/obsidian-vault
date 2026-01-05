// Rate Limiting with Cloudflare KV
export class RateLimiter {
  constructor(kvNamespace) {
    this.kv = kvNamespace;
  }
  
  async check(key, requests, windowSeconds) {
    try {
      const now = Date.now();
      const windowStart = now - (windowSeconds * 1000);
      const rateKey = `rate:${key}:${Math.floor(now / (windowSeconds * 1000))}`;
      
      const current = await this.kv.get(rateKey, { type: 'json' });
      const count = (current && current.count) || 0;
      
      if (count >= requests) {
        // Log rate limit hit
        await this.logRateLimitHit(key, requests, windowSeconds, count);
        return false;
      }
      
      // Increment counter
      await this.kv.put(rateKey, JSON.stringify({
        count: count + 1,
        timestamp: now,
        windowStart,
        windowEnd: now + (windowSeconds * 1000)
      }), {
        expirationTtl: windowSeconds + 60 // Extra minute for cleanup
      });
      
      return true;
    } catch (error) {
      console.error('Rate limiter error:', error);
      // Fail open - if rate limiting fails, allow the request
      return true;
    }
  }
  
  async logRateLimitHit(key, limit, window, actual) {
    const logKey = `rate-limit-log:${Date.now()}`;
    await this.kv.put(logKey, JSON.stringify({
      key,
      limit,
      windowSeconds: window,
      actualCount: actual,
      timestamp: Date.now()
    }), {
      expirationTtl: 86400 // Keep logs for 24 hours
    });
  }
  
  async getRateLimitStatus(key, windowSeconds = 60) {
    try {
      const now = Date.now();
      const windowStart = now - (windowSeconds * 1000);
      const rateKey = `rate:${key}:${Math.floor(now / (windowSeconds * 1000))}`;
      
      const current = await this.kv.get(rateKey, { type: 'json' });
      if (!current) {
        return {
          remaining: 'unlimited',
          resetTime: now + (windowSeconds * 1000)
        };
      }
      
      return {
        used: current.count,
        remaining: Math.max(0, requests - current.count),
        resetTime: current.windowEnd,
        windowStart: current.windowStart
      };
    } catch (error) {
      console.error('Rate limit status error:', error);
      return null;
    }
  }
  
  async reset(key) {
    try {
      const now = Date.now();
      const rateKey = `rate:${key}:${Math.floor(now / 60000)}`; // Current minute
      await this.kv.delete(rateKey);
      console.log(`Reset rate limit for key: ${key}`);
    } catch (error) {
      console.error('Rate limit reset error:', error);
    }
  }
  
  // Advanced rate limiting for different scenarios
  async checkBurst(key, maxBurst, refillRate, windowSeconds) {
    try {
      const now = Date.now();
      const tokenKey = `tokens:${key}`;
      
      const tokensData = await this.kv.get(tokenKey, { type: 'json' });
      let tokens = tokensData ? tokensData.tokens : maxBurst;
      let lastUpdate = tokensData ? tokensData.lastUpdate : now;
      
      // Refill tokens based on time passed
      const timePassed = (now - lastUpdate) / 1000; // seconds
      const tokensToAdd = Math.min(timePassed * refillRate, maxBurst - tokens);
      tokens += tokensToAdd;
      
      if (tokens < 1) {
        return {
          allowed: false,
          tokens,
          nextRefill: lastUpdate + Math.ceil((1 - tokens) / refillRate * 1000)
        };
      }
      
      // Consume one token
      tokens -= 1;
      
      await this.kv.put(tokenKey, JSON.stringify({
        tokens,
        lastUpdate: now,
        maxBurst,
        refillRate
      }), {
        expirationTtl: windowSeconds + 60
      });
      
      return {
        allowed: true,
        tokens,
        nextRefill: now + Math.ceil((maxBurst - tokens) / refillRate * 1000)
      };
    } catch (error) {
      console.error('Burst rate limiter error:', error);
      return { allowed: true };
    }
  }
}

export default RateLimiter;