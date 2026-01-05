// Cache Management with Cloudflare KV
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
      await this.kv.delete(key);
      console.log(`Deleted cache key: ${key}`);
    } catch (error) {
      console.error('Cache delete error:', error);
    }
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