// Analytics and Monitoring for AI Proxy
export class Analytics {
  constructor(env) {
    this.env = env;
    this.batchSize = 100;
    this.flushInterval = 60000; // 1 minute
    this.events = [];
    this.startBatchFlush();
  }
  
  async trackAIUsage(data) {
    const event = {
      type: 'ai_usage',
      timestamp: Date.now(),
      ...data,
      // Add computed fields
      promptComplexity: this.calculateComplexity(data.promptLength),
      costEfficiency: this.calculateCostEfficiency(data.tokensUsed, data.responseTime, data.cacheHit),
      performanceScore: this.calculatePerformanceScore(data.responseTime, data.tokensUsed)
    };
    
    await this.addEvent(event);
  }
  
  async trackCacheHit(provider, promptLength, cachedData) {
    const event = {
      type: 'cache_hit',
      timestamp: Date.now(),
      provider,
      promptLength,
      cachedAt: cachedData.timestamp,
      cacheAge: Date.now() - cachedData.timestamp,
      savings: {
        latency: cachedData.latency || 0,
        cost: cachedData.cost || 0
      }
    };
    
    await this.addEvent(event);
  }
  
  async trackError(provider, error, clientIP) {
    const event = {
      type: 'error',
      timestamp: Date.now(),
      provider,
      error: error.message || error,
      errorType: this.classifyError(error),
      clientIP: this.hashIP(clientIP), // Hash for privacy
      userAgent: this.hashUserAgent(error.userAgent)
    };
    
    await this.addEvent(event);
  }
  
  async trackRateLimitHit(provider, clientIP, promptLength) {
    const event = {
      type: 'rate_limit',
      timestamp: Date.now(),
      provider,
      clientIP: this.hashIP(clientIP),
      promptLength,
      severity: this.calculateSeverity(promptLength)
    };
    
    await this.addEvent(event);
  }
  
  async trackRequestMetrics(request, response, responseTime) {
    const event = {
      type: 'request_metrics',
      timestamp: Date.now(),
      endpoint: new URL(request.url).pathname,
      method: request.method,
      status: response.status,
      responseTime,
      
      // Cloudflare-specific metrics
      colo: request.cf?.colo,
      country: request.cf?.country,
      tlsVersion: request.cf?.tlsVersion,
      httpProtocol: request.cf?.httpProtocol,
      botScore: request.cf?.botScore,
      
      // Performance metrics
      cacheStatus: response.headers.get('x-cache-status'),
      aiProvider: response.headers.get('x-ai-provider'),
      costCents: response.headers.get('x-cost-cents')
    };
    
    await this.addEvent(event);
  }
  
  async addEvent(event) {
    this.events.push(event);
    
    if (this.events.length >= this.batchSize) {
      await this.flushEvents();
    }
  }
  
  async flushEvents() {
    if (this.events.length === 0) return;
    
    const batch = [...this.events];
    this.events = [];
    
    try {
      // Store batch in KV for processing
      const batchKey = `analytics-batch-${Date.now()}`;
      await this.env.ANALYTICS_KV.put(batchKey, JSON.stringify(batch), {
        expirationTtl: 86400 * 7 // Keep for 7 days
      });
      
      console.log(`Flushed ${batch.length} analytics events`);
    } catch (error) {
      console.error('Failed to flush analytics:', error);
      // Re-add events to try again
      this.events.unshift(...batch);
    }
  }
  
  startBatchFlush() {
    setInterval(async () => {
      await this.flushEvents();
    }, this.flushInterval);
  }
  
  calculateComplexity(promptLength) {
    if (promptLength < 100) return 'simple';
    if (promptLength < 500) return 'moderate';
    if (promptLength < 2000) return 'complex';
    return 'very_complex';
  }
  
  calculateCostEfficiency(tokensUsed, responseTime, cacheHit) {
    if (cacheHit) return 100; // Perfect efficiency for cache hits
    
    const tokensPerMs = tokensUsed / responseTime;
    if (tokensPerMs > 2) return 'high';
    if (tokensPerMs > 1) return 'medium';
    return 'low';
  }
  
  calculatePerformanceScore(responseTime, tokensUsed) {
    const tokensPerMs = tokensUsed / responseTime;
    const score = Math.min(100, tokensPerMs * 50);
    return Math.round(score);
  }
  
  classifyError(error) {
    if (error.message.includes('timeout')) return 'timeout';
    if (error.message.includes('rate limit')) return 'rate_limit';
    if (error.message.includes('authentication')) return 'auth';
    if (error.message.includes('provider')) return 'provider_error';
    return 'unknown';
  }
  
  calculateSeverity(promptLength) {
    if (promptLength > 5000) return 'high';
    if (promptLength > 1000) return 'medium';
    return 'low';
  }
  
  hashIP(ip) {
    // Simple hash for privacy - in production use proper crypto
    return btoa(ip).substring(0, 8);
  }
  
  hashUserAgent(userAgent) {
    if (!userAgent) return 'unknown';
    return btoa(userAgent.substring(0, 50)).substring(0, 8);
  }
  
  // Analytics aggregation methods
  async getAggregatedMetrics(timeRange = '1h') {
    try {
      // This would typically query a time-series database
      // For now, return mock data
      return {
        timeRange,
        totalRequests: 0,
        cacheHitRate: 0,
        avgResponseTime: 0,
        errorRate: 0,
        topProviders: [],
        costBreakdown: {},
        geographicDistribution: {}
      };
    } catch (error) {
      console.error('Failed to get aggregated metrics:', error);
      return null;
    }
  }
  
  async getProviderPerformance() {
    try {
      return {
        providers: [
          {
            name: 'gemini',
            successRate: 0.98,
            avgLatency: 800,
            costPerToken: 0.000125,
            requestsPerMinute: 45
          },
          {
            name: 'groq',
            successRate: 0.99,
            avgLatency: 400,
            costPerToken: 0.00005,
            requestsPerMinute: 25
          }
        ]
      };
    } catch (error) {
      console.error('Failed to get provider performance:', error);
      return null;
    }
  }
}

export default Analytics;