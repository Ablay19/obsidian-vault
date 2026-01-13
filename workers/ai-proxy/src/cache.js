// Enhanced Cache Management with Cloudflare KV and Intelligent Features
export class CacheManager {
  constructor(kvNamespace) {
    this.kv = kvNamespace;
  }
  
  async get(key) {
    try {
      if (!this.kv) return null;
      const value = await this.kv.get(key, { type: 'json' });
      if (!value) return null;
      
      // Check if expired
      if (value.expiresAt && Date.now() > value.expiresAt) {
        await this.kv.delete(key);
        return null;
      }
      
      return value;
    } catch (error) {
      console.error('Cache get error:', error);
      return null;
    }
  }
  
  async set(key, value, ttlSeconds = 3600) {
    try {
      if (!this.kv) return;
      const cacheEntry = {
        ...value,
        expiresAt: Date.now() + (ttlSeconds * 1000),
        cachedAt: Date.now()
      };
      
      await this.kv.put(key, JSON.stringify(cacheEntry), {
        expirationTtl: ttlSeconds,
        metadata: {
          provider: value.provider || 'unknown',
          cachedAt: cacheEntry.cachedAt.toString()
        }
      });
      
      console.log(`Cached key: ${key}, TTL: ${ttlSeconds}s`);
    } catch (error) {
      console.error('Cache set error:', error);
    }
  }
  
  async delete(key) {
    try {
      if (!this.kv) return;
      await this.kv.delete(key);
    } catch (error) {
      console.error('Cache delete error:', error);
    }
  }
}

// Intelligent Cache Manager with LRU Eviction and Analytics
export class SmartCache {
  constructor(maxSize = 100, analytics = null) {
    this.cache = new Map();
    this.accessTimes = new Map();
    this.maxSize = maxSize;
    this.analytics = analytics;
    this.stats = {
      hits: 0,
      misses: 0,
      evictions: 0,
      totalSets: 0
    };
  }

  set(key, value, ttl = 300000) { // 5 minutes default
    if (this.cache.size >= this.maxSize) {
      this.evictLeastRecentlyUsed();
    }

    this.cache.set(key, {
      value,
      expiry: Date.now() + ttl,
      createdAt: Date.now()
    });
    this.accessTimes.set(key, Date.now());
    this.stats.totalSets++;

    if (this.analytics) {
      this.analytics.trackCacheSet(key, ttl);
    }
  }

  get(key) {
    const item = this.cache.get(key);
    if (!item) {
      this.stats.misses++;
      if (this.analytics) {
        this.analytics.trackCacheMiss(key);
      }
      return null;
    }

    if (Date.now() > item.expiry) {
      this.delete(key);
      this.stats.misses++;
      if (this.analytics) {
        this.analytics.trackCacheMiss(key);
      }
      return null;
    }

    this.accessTimes.set(key, Date.now());
    this.stats.hits++;
    if (this.analytics) {
      this.analytics.trackCacheHit(key, item);
    }
    return item.value;
  }

  delete(key) {
    this.cache.delete(key);
    this.accessTimes.delete(key);
  }

  evictLeastRecentlyUsed() {
    let oldestKey = null;
    let oldestTime = Date.now();

    for (const [key, time] of this.accessTimes) {
      if (time < oldestTime) {
        oldestTime = time;
        oldestKey = key;
      }
    }

    if (oldestKey) {
      this.cache.delete(oldestKey);
      this.accessTimes.delete(oldestKey);
      this.stats.evictions++;
    }
  }

  getEfficiency() {
    const total = this.stats.hits + this.stats.misses;
    return total > 0 ? this.stats.hits / total : 0;
  }

  getReport() {
    return {
      hitRate: this.getEfficiency(),
      totalOperations: this.stats.hits + this.stats.misses,
      evictions: this.stats.evictions,
      cacheSize: this.cache.size,
      maxSize: this.maxSize,
      efficiency: `${(this.getEfficiency() * 100).toFixed(1)}%`
    };
  }

  clear() {
    this.cache.clear();
    this.accessTimes.clear();
    this.stats = {
      hits: 0,
      misses: 0,
      evictions: 0,
      totalSets: 0
    };
  }

  // Automatic cleanup of expired items
  cleanup() {
    const now = Date.now();
    for (const [key, item] of this.cache) {
      if (now > item.expiry) {
        this.delete(key);
      }
    }
  }
}

// Cache Analytics for monitoring cache performance
export class CacheAnalytics {
  constructor() {
    this.stats = {
      hits: 0,
      misses: 0,
      evictions: 0,
      totalSets: 0,
      hitLatencies: [],
      missLatencies: []
    };
    this.startTime = Date.now();
  }

  recordHit(key, item) {
    this.stats.hits++;
    const latency = Date.now() - item.createdAt;
    this.stats.hitLatencies.push(latency);
    if (this.stats.hitLatencies.length > 1000) {
      this.stats.hitLatencies.shift();
    }
  }

  recordMiss(key) {
    this.stats.misses++;
  }

  recordEviction() {
    this.stats.evictions++;
  }

  recordSet(key, ttl) {
    this.stats.totalSets++;
  }

  getEfficiency() {
    const total = this.stats.hits + this.stats.misses;
    return total > 0 ? this.stats.hits / total : 0;
  }

  getAverageHitLatency() {
    if (this.stats.hitLatencies.length === 0) return 0;
    const sum = this.stats.hitLatencies.reduce((a, b) => a + b, 0);
    return sum / this.stats.hitLatencies.length;
  }

  getReport() {
    return {
      hitRate: this.getEfficiency(),
      totalOperations: this.stats.hits + this.stats.misses,
      evictions: this.stats.evictions,
      totalSets: this.stats.totalSets,
      averageHitLatency: this.getAverageHitLatency(),
      uptime: Date.now() - this.startTime,
      hitRatePercent: `${(this.getEfficiency() * 100).toFixed(1)}%`
    };
  }

  reset() {
    this.stats = {
      hits: 0,
      misses: 0,
      evictions: 0,
      totalSets: 0,
      hitLatencies: []
    };
    this.startTime = Date.now();
  }

  async clearByPattern(pattern) {
    try {
      // Note: KV doesn't support pattern matching directly
      // This would need to be implemented with a separate index
      console.log(`Pattern clearing not implemented: ${pattern}`);
    } catch (error) {
      console.error('Cache clear error:', error);
    }
  }
  
  async getStats() {
    try {
      // Get cache statistics (would need to implement custom tracking)
      return {
        hitRate: 0, // Would be tracked separately
        totalKeys: 0,
        totalSize: 0
      };
    } catch (error) {
      console.error('Cache stats error:', error);
      return null;
    }
  }
}

export default CacheManager;