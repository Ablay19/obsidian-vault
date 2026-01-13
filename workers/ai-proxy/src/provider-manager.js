export class ProviderManager {
  constructor(providers, optimizer, analytics) {
    this.providers = providers;
    this.optimizer = optimizer;
    this.analytics = analytics;

    // Enhanced load balancing and health tracking
    this.healthStatus = new Map();
    this.loadDistribution = new Map();
    this.performanceHistory = new Map();

    // Initialize health status for all providers
    Object.keys(providers.providers || {}).forEach(providerName => {
      this.healthStatus.set(providerName, {
        healthy: true,
        lastChecked: Date.now(),
        consecutiveFailures: 0,
        responseTime: 0,
        successRate: 1.0
      });

      this.loadDistribution.set(providerName, 0);
      this.performanceHistory.set(providerName, []);
    });
  }

  async selectProvider(requestData) {
    const { provider: requestedProvider, prompt } = requestData;
    return this.optimizer.selectProvider(prompt, {
      preferredProvider: requestedProvider,
      maxLatency: 2000,
      maxCostPerToken: 0.001,
      clientRegion: null // no cf in request
    });
  }

  async executeRequest(provider, prompt, requestData) {
    const { clientIP, userAgent } = requestData;
    const startTime = Date.now();

    try {
      const result = await this.providers.callProvider(provider.name, {
        prompt,
        maxTokens: 1000,
        temperature: 0.7
      });

      const responseTime = Date.now() - startTime;
      await this.analytics.trackRequest(provider.name, prompt.length, result, responseTime, clientIP, userAgent);

      return { result, provider: provider.name, responseTime };
    } catch (error) {
      await this.analytics.trackError(provider.name, error, prompt.length, clientIP);
      throw error;
    }
  }

  async handleFallback(failedProvider, requestData) {
    const { prompt, clientIP } = requestData;
    console.log(`Attempting fallback for failed provider: ${failedProvider.name}`);

    const fallback = await this.optimizer.getFallbackProvider(failedProvider, prompt.length);
    if (fallback) {
      return this.executeRequest(fallback, prompt, requestData);
    }
    return null;
  }

  // Enhanced load balancing methods
  async selectProviderWithLoadBalancing(requestData) {
    const { provider: requestedProvider, prompt } = requestData;

    // Get all healthy providers
    const healthyProviders = this.getHealthyProviders();

    if (healthyProviders.length === 0) {
      throw new Error('No healthy providers available');
    }

    // If specific provider requested and healthy, use it
    if (requestedProvider && healthyProviders.includes(requestedProvider)) {
      const provider = await this.optimizer.selectProvider(prompt, {
        preferredProvider: requestedProvider,
        maxLatency: 2000,
        maxCostPerToken: 0.001,
        clientRegion: null
      });
      if (provider) {
        this.loadDistribution.set(requestedProvider, this.loadDistribution.get(requestedProvider) + 1);
        return provider;
      }
    }

    // Use load balancing to select best provider
    const selectedProvider = this.selectBalancedProvider(healthyProviders, prompt);
    if (selectedProvider) {
      this.loadDistribution.set(selectedProvider.name, this.loadDistribution.get(selectedProvider.name) + 1);
    }
    return selectedProvider;
  }

  getHealthyProviders() {
    const healthy = [];
    for (const [providerName, status] of this.healthStatus) {
      if (status.healthy && status.successRate > 0.8) { // 80% success rate threshold
        healthy.push(providerName);
      }
    }
    return healthy;
  }

  selectBalancedProvider(healthyProviders, prompt) {
    let bestProvider = null;
    let bestScore = -1;

    for (const providerName of healthyProviders) {
      const score = this.calculateLoadBalanceScore(providerName);
      if (score > bestScore) {
        const provider = this.providers.providers[providerName];
        if (provider) {
          bestScore = score;
          bestProvider = {
            name: providerName,
            ...provider,
            score: score
          };
        }
      }
    }

    return bestProvider;
  }

  calculateLoadBalanceScore(providerName) {
    const currentLoad = this.loadDistribution.get(providerName) || 0;
    const healthStatus = this.healthStatus.get(providerName);
    const performance = this.performanceHistory.get(providerName) || [];

    // Base score from health and performance
    let score = healthStatus.successRate * 100;

    // Penalize high load
    score -= currentLoad * 2;

    // Bonus for good recent performance
    if (performance.length > 0) {
      const recentAvg = performance.slice(-5).reduce((a, b) => a + b, 0) / Math.min(5, performance.length);
      score += Math.max(0, 1000 - recentAvg); // Lower latency = higher score
    }

    return Math.max(0, score);
  }

  // Health checking methods
  async checkProviderHealth(providerName) {
    const status = this.healthStatus.get(providerName);
    if (!status) return false;

    try {
      // Simple health check - attempt a small request
      const testResult = await this.providers.callProvider(providerName, {
        prompt: "test",
        maxTokens: 10,
        temperature: 0.1
      });

      const responseTime = Date.now() - status.lastChecked;

      // Update health status
      status.healthy = true;
      status.lastChecked = Date.now();
      status.consecutiveFailures = 0;
      status.responseTime = responseTime;

      // Update performance history
      const perfHistory = this.performanceHistory.get(providerName) || [];
      perfHistory.push(responseTime);
      if (perfHistory.length > 100) {
        perfHistory.shift(); // Keep last 100 measurements
      }
      this.performanceHistory.set(providerName, perfHistory);

      // Update success rate
      status.successRate = Math.min(1.0, status.successRate + 0.1); // Gradual improvement

      return true;
    } catch (error) {
      // Update health status on failure
      status.healthy = false;
      status.consecutiveFailures++;
      status.successRate = Math.max(0.0, status.successRate - 0.2); // Gradual degradation

      return false;
    }
  }

  async performHealthChecks() {
    const promises = [];
    for (const providerName of this.healthStatus.keys()) {
      promises.push(this.checkProviderHealth(providerName));
    }
    await Promise.allSettled(promises);
  }

  // Performance tracking methods
  updateProviderPerformance(providerName, responseTime, success) {
    const status = this.healthStatus.get(providerName);
    if (status) {
      if (success) {
        status.successRate = Math.min(1.0, status.successRate + 0.01);
        status.consecutiveFailures = 0;
      } else {
        status.successRate = Math.max(0.0, status.successRate - 0.05);
        status.consecutiveFailures++;
      }
    }

    // Update performance history
    const perfHistory = this.performanceHistory.get(providerName) || [];
    perfHistory.push(responseTime);
    if (perfHistory.length > 100) {
      perfHistory.shift();
    }
    this.performanceHistory.set(providerName, perfHistory);
  }

  // Analytics and reporting
  getProviderStats() {
    const stats = {};
    for (const [providerName, status] of this.healthStatus) {
      const load = this.loadDistribution.get(providerName) || 0;
      const performance = this.performanceHistory.get(providerName) || [];
      const avgResponseTime = performance.length > 0
        ? performance.reduce((a, b) => a + b, 0) / performance.length
        : 0;

      stats[providerName] = {
        healthy: status.healthy,
        load: load,
        successRate: status.successRate,
        averageResponseTime: avgResponseTime,
        consecutiveFailures: status.consecutiveFailures,
        lastChecked: status.lastChecked
      };
    }
    return stats;
  }

  getLoadDistribution() {
    const total = Array.from(this.loadDistribution.values()).reduce((a, b) => a + b, 0);
    const distribution = {};

    for (const [provider, load] of this.loadDistribution) {
      distribution[provider] = {
        requests: load,
        percentage: total > 0 ? (load / total * 100).toFixed(1) + '%' : '0%'
      };
    }

    return distribution;
  }
}