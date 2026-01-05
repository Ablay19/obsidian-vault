// Intelligent Cost Optimization for AI Provider Selection
export class CostOptimizer {
  constructor(env) {
    this.env = env;
    this.usage = new Map(); // Track usage for rate limiting
    this.performance = new Map(); // Track actual performance
    this.costs = new Map(); // Track cost metrics
  }
  
  async selectProvider(prompt, constraints) {
    const promptLength = prompt.length;
    const estimatedTokens = Math.ceil(promptLength / 4); // Rough estimate
    
    const availableProviders = Object.entries(this.providers || {})
      .map(([name, config]) => ({
        name,
        ...config,
        score: this.calculateScore(name, config, constraints, estimatedTokens),
      }))
      .filter(provider => {
        // Filter by constraints
        if (estimatedTokens > provider.maxTokens) return false;
        if (provider.latency > constraints.maxLatency) return false;
        if (provider.costPerToken > constraints.maxCostPerToken) return false;
        if (!provider.enabled) return false;
        
        // Check rate limits
        const currentUsage = this.usage.get(provider.name) || 0;
        if (currentUsage >= this.getRateLimit(provider.name)) return false;
        
        return true;
      })
      .sort((a, b) => b.score - a.score);
    
    return availableProviders[0] || null;
  }
  
  async selectFallbackProvider(prompt, excludedProvider) {
    const constraints = {
      maxLatency: 3000, // More lenient for fallbacks
      maxCostPerToken: 0.002, // Higher cost tolerance
      clientRegion: 'global'
    };
    
    const provider = await this.selectProvider(prompt, constraints);
    return provider && provider.name !== excludedProvider ? provider : null;
  }
  
  calculateScore(providerName, config, constraints, estimatedTokens) {
    let score = 100;
    
    // Cost weighting (40%)
    const costScore = (1 - config.costPerToken / 0.001) * 40;
    score += costScore;
    
    // Latency weighting (30%)
    const latencyScore = (1 - config.latency / 2000) * 30;
    score += latencyScore;
    
    // Token limit weighting (20%)
    const tokenScore = Math.min(config.maxTokens / 8192, 1) * 20;
    score += tokenScore;
    
    // Reliability weighting (10%)
    const reliability = this.performance.get(providerName)?.successRate || 0.95;
    const reliabilityScore = reliability * 10;
    score += reliabilityScore;
    
    // Region proximity bonus (if available)
    if (constraints.clientRegion && this.isNearRegion(config.region, constraints.clientRegion)) {
      score += 10;
    }
    
    return Math.round(score);
  }
  
  isNearRegion(providerRegion, clientRegion) {
    const regionMap = {
      'us-east': ['iad', 'atl', 'mia', 'jfk', 'lga', 'ord'],
      'us-west': ['sfo', 'lax', 'sea', 'den', 'slc'],
      'europe': ['lhr', 'fra', 'ams', 'cdg', 'mad'],
      'asia': ['sin', 'hkg', 'nrt', 'bkk', 'mnl'],
      'australia': ['syd', 'mel', 'bne']
    };
    
    const providerRegions = regionMap[providerRegion] || [];
    return providerRegions.includes(clientRegion);
  }
  
  getRateLimit(providerName) {
    const limits = {
      gemini: 60,    // requests per minute
      groq: 30,
      claude: 20,
      gpt4: 100,
    };
    return limits[providerName] || 10;
  }
  
  updateUsage(providerName) {
    const current = this.usage.get(providerName) || 0;
    this.usage.set(providerName, current + 1);
    
    // Reset usage counter every minute
    setTimeout(() => {
      this.usage.set(providerName, 0);
    }, 60000);
  }
  
  recordPerformance(providerName, latency, success) {
    const current = this.performance.get(providerName) || { 
      totalRequests: 0, 
      totalLatency: 0, 
      failures: 0 
    };
    
    current.totalRequests++;
    current.totalLatency += latency;
    if (!success) current.failures++;
    
    this.performance.set(providerName, {
      ...current,
      avgLatency: current.totalLatency / current.totalRequests,
      successRate: (current.totalRequests - current.failures) / current.totalRequests,
    });
  }
  
  recordCost(providerName, tokensUsed, costInCents) {
    const current = this.costs.get(providerName) || { 
      totalTokens: 0, 
      totalCost: 0 
    };
    
    current.totalTokens += tokensUsed;
    current.totalCost += costInCents;
    
    this.costs.set(providerName, current);
  }
  
  getCostOptimizationReport() {
    return Array.from(this.costs.entries()).map(([provider, metrics]) => ({
      provider,
      totalTokens: metrics.totalTokens,
      totalCost: metrics.totalCost,
      avgCostPerToken: metrics.totalCost / metrics.totalTokens,
      performance: this.performance.get(providerName)
    }));
  }
}

export default CostOptimizer;