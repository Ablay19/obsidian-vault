// Enhanced Rate Limiting with Multiple Algorithms
export class RateLimiter {
  constructor(kvNamespace) {
    this.kv = kvNamespace;

    // Token bucket configuration
    this.buckets = new Map(); // In-memory buckets for fast access

    // Sliding window configuration
    this.windows = new Map(); // Request timestamps for sliding window

    // Statistics
    this.stats = {
      totalRequests: 0,
      blockedRequests: 0,
      slidingWindowChecks: 0,
      tokenBucketRefills: 0
    };
  }

  // Enhanced rate limiting with multiple algorithms
  allow(key, options = {}) {
    const {
      algorithm = 'token-bucket',
      capacity = 100,        // Token bucket capacity
      refillRate = 10,       // Tokens per second
      windowSize = 60,       // Sliding window in seconds
      burstAllowance = 20    // Burst capacity
    } = options;

    this.stats.totalRequests++;

    switch (algorithm) {
      case 'token-bucket':
        return this.checkTokenBucket(key, capacity, refillRate, burstAllowance);
      case 'sliding-window':
        return this.checkSlidingWindow(key, capacity, windowSize);
      case 'fixed-window':
        return this.checkFixedWindow(key, capacity, windowSize);
      default:
        return this.checkTokenBucket(key, capacity, refillRate, burstAllowance);
    }
  }

  // Token Bucket Algorithm
  checkTokenBucket(key, capacity, refillRate, burstAllowance) {
    const now = Date.now();
    const bucketKey = `tb-${key}`;

    let bucket = this.buckets.get(bucketKey);
    if (!bucket) {
      bucket = {
        tokens: capacity,
        lastRefill: now,
        maxCapacity: capacity + burstAllowance
      };
      this.buckets.set(bucketKey, bucket);
    }

    // Refill tokens based on time passed
    const timePassed = now - bucket.lastRefill;
    const tokensToAdd = Math.floor((timePassed / 1000) * refillRate);

    if (tokensToAdd > 0) {
      bucket.tokens = Math.min(bucket.maxCapacity, bucket.tokens + tokensToAdd);
      bucket.lastRefill = now;
      this.stats.tokenBucketRefills++;
    }

    // Check if we have enough tokens
    if (bucket.tokens >= 1) {
      bucket.tokens--;
      return true;
    }

    this.stats.blockedRequests++;
    return false;
  }

  // Sliding Window Algorithm
  checkSlidingWindow(key, maxRequests, windowSize) {
    const now = Date.now();
    const windowKey = `sw-${key}`;

    let requests = this.windows.get(windowKey) || [];
    this.stats.slidingWindowChecks++;

    // Remove requests outside the window
    const cutoff = now - (windowSize * 1000);
    requests = requests.filter(timestamp => timestamp > cutoff);

    // Check if under limit
    if (requests.length < maxRequests) {
      requests.push(now);
      this.windows.set(windowKey, requests);
      return true;
    }

    this.stats.blockedRequests++;
    return false;
  }

  // Fixed Window Algorithm (existing implementation enhanced)
  async checkFixedWindow(key, maxRequests, windowSize) {
    return this.check(key, maxRequests, windowSize);
  }

  // Legacy method for backward compatibility
  async check(key, requests, windowSeconds) {
    try {
      if (!this.kv) return true; // Allow all if no KV binding
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

  // Enhanced statistics and reporting
  getStats() {
    const totalRequests = this.stats.totalRequests;
    const blockedRequests = this.stats.blockedRequests;
    const blockRate = totalRequests > 0 ? (blockedRequests / totalRequests) * 100 : 0;

    return {
      totalRequests,
      blockedRequests,
      blockRate: `${blockRate.toFixed(2)}%`,
      slidingWindowChecks: this.stats.slidingWindowChecks,
      tokenBucketRefills: this.stats.tokenBucketRefills,
      activeBuckets: this.buckets.size,
      activeWindows: this.windows.size
    };
  }

  // Graceful degradation - allow some requests even when rate limited
  allowWithDegradation(key, options = {}) {
    const baseAllow = this.allow(key, options);

    // If rate limited, allow 10% of requests for critical operations
    if (!baseAllow) {
      const degradationKey = `degrade-${key}`;
      const random = Math.random();
      if (random < 0.1) { // 10% chance
        // Log degradation allowance
        this.logDegradationAllowance(key, degradationKey);
        return true;
      }
    }

    return baseAllow;
  }

  // Burst handling for temporary load spikes
  allowBurst(key, burstCapacity = 50, recoveryRate = 5) {
    const burstKey = `burst-${key}`;
    const now = Date.now();

    let burstData = this.buckets.get(burstKey);
    if (!burstData) {
      burstData = {
        tokens: burstCapacity,
        lastRefill: now,
        recoveryRate
      };
      this.buckets.set(burstKey, burstData);
    }

    // Recover tokens over time
    const timePassed = now - burstData.lastRefill;
    const tokensToRecover = Math.floor((timePassed / 1000) * burstData.recoveryRate);

    if (tokensToRecover > 0) {
      burstData.tokens = Math.min(burstCapacity, burstData.tokens + tokensToRecover);
      burstData.lastRefill = now;
    }

    if (burstData.tokens > 0) {
      burstData.tokens--;
      return true;
    }

    return false;
  }

  // Reset rate limits for a key (admin function)
  reset(key) {
    // Clear all variants of the key
    this.buckets.delete(`tb-${key}`);
    this.buckets.delete(`burst-${key}`);
    this.windows.delete(`sw-${key}`);

    // Clear from KV if available
    if (this.kv) {
      // Would need to clear KV keys with pattern matching
      // This is simplified for the example
    }
  }

  // Get reset time for rate limits
  getResetTime(key, algorithm = 'token-bucket', windowSize = 60) {
    switch (algorithm) {
      case 'sliding-window':
        return Date.now() + (windowSize * 1000);
      case 'fixed-window':
        // For fixed window, next window starts at next interval
        const now = Date.now();
        const windowStart = Math.floor(now / (windowSize * 1000)) * (windowSize * 1000);
        return windowStart + (windowSize * 1000);
      case 'token-bucket':
      default:
        // For token bucket, tokens refill continuously
        return Date.now() + 1000; // Next second for token refill
    }
  }

  // Logging methods
  async logRateLimitHit(key, limit, window, actual) {
    const logKey = `rate-limit-log:${Date.now()}`;
    await this.kv.put(logKey, JSON.stringify({
      key,
      limit,
      windowSeconds: window,
      actualCount: actual,
      timestamp: Date.now(),
      algorithm: 'enhanced'
    }), {
      expirationTtl: 86400 // Keep logs for 24 hours
    });
  }

  logDegradationAllowance(key, degradationKey) {
    // Log when graceful degradation allows a request
    console.log(`Rate limit degradation allowed for key: ${key} (${degradationKey})`);
  }

  // Cleanup expired data
  cleanup() {
    const now = Date.now();

    // Clean up old sliding windows (older than 10 minutes)
    for (const [key, requests] of this.windows) {
      const cutoff = now - (10 * 60 * 1000); // 10 minutes
      const filtered = requests.filter(timestamp => timestamp > cutoff);
      if (filtered.length === 0) {
        this.windows.delete(key);
      } else {
        this.windows.set(key, filtered);
      }
    }

    // Clean up old token buckets (older than 1 hour)
    for (const [key, bucket] of this.buckets) {
      if (now - bucket.lastRefill > (60 * 60 * 1000)) { // 1 hour
        this.buckets.delete(key);
      }
    }
  }
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