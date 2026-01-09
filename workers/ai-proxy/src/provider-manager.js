export class ProviderManager {
  constructor(providers, optimizer, analytics) {
    this.providers = providers;
    this.optimizer = optimizer;
    this.analytics = analytics;
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
}