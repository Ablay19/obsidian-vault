// Cloudflare Workers AI Proxy for Obsidian Bot
import { AIProviders } from './providers.js';
import { CacheManager } from './cache.js';
import { RateLimiter } from './rate-limiter.js';
import { CostOptimizer } from './cost-optimizer.js';
import { Analytics } from './analytics.js';
import { RequestHandler } from './request-handler.js';
import { ProviderManager } from './provider-manager.js';

class AIRequestProcessor {
  constructor(env) {
    this.env = env;
    this.cache = new CacheManager(env.AI_CACHE || null);
    this.providers = new AIProviders(env);
    this.limiter = new RateLimiter(env.AI_CACHE || null);
    this.optimizer = new CostOptimizer(env, this.providers);
    this.analytics = new Analytics(env);
    this.requestHandler = new RequestHandler();
    this.providerManager = new ProviderManager(this.providers, this.optimizer, this.analytics);
  }

  async processRequest(request) {
    const startTime = Date.now();
    const requestData = await this.requestHandler.parseRequest(request);
    
    try {
      // 1. Check cache first for identical prompts
      const cacheKey = await this.generateCacheKey(requestData.prompt, requestData.provider);
      const cached = await this.cache.get(cacheKey);
      if (cached && !cached.expired) {
        console.log(`Cache hit for provider: ${cached.provider}`);
        await this.analytics.trackCacheHit(requestData.provider, requestData.prompt.length, cached);

        return this.requestHandler.buildResponse(cached.data, {
          provider: cached.provider,
          cacheStatus: 'hit',
          responseTime: Date.now() - startTime
        });
      }
      
      // 2. Rate limiting per provider and IP
      const rateLimitKey = `ai-${requestData.provider}-${requestData.clientIP}`;
      const isAllowed = await this.limiter.check(rateLimitKey, 1, 60);
      if (!isAllowed) {
        await this.analytics.trackRateLimitHit(requestData.provider, requestData.clientIP, requestData.prompt.length);
        return this.requestHandler.buildErrorResponse(
          { message: "Rate limit exceeded. Please try again later.", status: 429 },
          Date.now() - startTime
        );
      }
      
      // 3. Select optimal provider based on current conditions
      const optimalProvider = await this.providerManager.selectProvider(requestData);
      
      if (!optimalProvider) {
        return this.requestHandler.buildErrorResponse(
          { message: "No available AI providers at the moment.", status: 503 },
          Date.now() - startTime
        );
      }
      
      // 4. Execute request with retries and fallbacks
      const execResult = await this.providerManager.executeRequest(optimalProvider, requestData.prompt, requestData);
      const response = execResult.result;
      const responseTime = execResult.responseTime;
      
      // 5. Cache response with intelligent TTL
      const ttl = this.calculateCacheTTL(requestData.prompt, optimalProvider);
      await this.cache.set(cacheKey, {
        data: response,
        provider: execResult.provider,
        timestamp: Date.now(),
        expired: false
      }, ttl);
      
      // 6. Track analytics
      await this.analytics.trackAIUsage({
        provider: execResult.provider,
        promptLength: requestData.prompt.length,
        responseLength: response.length,
        tokensUsed: this.estimateTokens(requestData.prompt, response),
        responseTime,
        cacheHit: false,
        clientIP: requestData.clientIP,
        userAgent: requestData.userAgent,
        colo: request.cf?.colo,
        country: request.cf?.country
      });
      
      return this.requestHandler.buildResponse(response, {
        provider: execResult.provider,
        cacheStatus: 'miss',
        responseTime,
        extraHeaders: {
          'x-cache-ttl': `${ttl}s`,
          'x-cost-cents': optimalProvider.costCents
        }
      });
      
    } catch (error) {
      console.error('AI request processing error:', error);
      await this.analytics.trackError(requestData.provider, error, requestData.clientIP);

      return this.requestHandler.buildErrorResponse(
        { message: "Internal server error processing AI request.", status: 500 },
        Date.now() - startTime
      );
    }
    }
  }
  
  async executeRequestWithRetry(prompt, provider, originalRequest) {
    const maxRetries = 3;
    const baseDelay = 1000;
    const startTime = Date.now();
    
    for (let attempt = 1; attempt <= maxRetries; attempt++) {
      try {
        const response = await this.executeRequest(prompt, provider, originalRequest);
        
        // Record successful performance
        await this.optimizer.recordPerformance(provider.name, Date.now() - startTime, true);
        return response;
        
      } catch (error) {
        console.error(`Attempt ${attempt} failed for provider ${provider.name}:`, error);
        
        // Record failure
        await this.optimizer.recordPerformance(provider.name, Date.now() - startTime, false);
        
        if (attempt === maxRetries) {
          // Try fallback provider
          const fallbackProvider = await this.optimizer.selectFallbackProvider(prompt, provider.name);
          if (fallbackProvider) {
            console.log(`Falling back to ${fallbackProvider.name}`);
            return await this.executeRequest(prompt, fallbackProvider, originalRequest);
          }
          throw error;
        }
        
        // Exponential backoff
        const delay = baseDelay * Math.pow(2, attempt - 1);
        await new Promise(resolve => setTimeout(resolve, delay));
      }
    }
  }
  
  async executeRequest(prompt, provider, originalRequest) {
    if (provider.name === 'cloudflare' || provider.name === 'llama3') {
      // Use Cloudflare AI directly
      try {
        const response = await this.env.AI.run(provider.model, {
          prompt: prompt
        });
        return response.response || '';
      } catch (error) {
        console.error('Cloudflare AI error:', error);
        throw new Error(`Cloudflare AI error: ${error.message}`);
      }
    }
    
    const headers = {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${provider.apiKey}`,
      'User-Agent': 'ObsidianBot/1.0 (CloudflareWorkers)'
    };
    
    const requestBody = this.formatRequestForProvider(prompt, provider);
    
    const response = await fetch(provider.endpoint, {
      method: 'POST',
      headers,
      body: JSON.stringify(requestBody)
    });
    
    if (!response.ok) {
      throw new Error(`Provider error: ${response.status} ${response.statusText}`);
    }
    
    const data = await response.json();
    return this.extractResponseFromProvider(data, provider);
  }
  
  formatRequestForProvider(prompt, provider) {
    switch (provider.name) {
      case 'gemini':
        return {
          contents: [{ parts: [{ text: prompt }] }]
        };
      case 'groq':
        return {
          model: provider.model,
          messages: [{ role: 'user', content: prompt }]
        };
      case 'claude':
        return {
          model: provider.model,
          max_tokens: provider.maxTokens,
          messages: [{ role: 'user', content: prompt }]
        };
      case 'gpt4':
        return {
          model: provider.model,
          messages: [{ role: 'user', content: prompt }],
          max_tokens: provider.maxTokens
        };
      case 'cloudflare':
      case 'llama3':
        return {
          prompt: prompt
        };
      default:
        throw new Error(`Unsupported provider: ${provider.name}`);
    }
  }
  
  extractResponseFromProvider(data, provider) {
    switch (provider.name) {
      case 'gemini':
        return data.candidates?.[0]?.content?.parts?.[0]?.text || '';
      case 'groq':
      case 'gpt4':
        return data.choices?.[0]?.message?.content || '';
      case 'claude':
        return data.content?.[0]?.text || '';
      case 'cloudflare':
      case 'llama3':
        return data.response || '';
      default:
        throw new Error(`Cannot extract response from provider: ${provider.name}`);
    }
  }
  
  async generateCacheKey(prompt, provider) {
    const encoder = new TextEncoder();
    const data = encoder.encode(prompt + provider);
    const hashBuffer = await crypto.subtle.digest('SHA-256', data);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
    return `ai:${provider}:${hashHex}`;
  }
  
  calculateCacheTTL(prompt, provider) {
    // Different caching strategies based on prompt type
    if (this.isCodeGeneration(prompt)) {
      return 86400; // 24 hours for code
    } else if (this.isGeneralQuestion(prompt)) {
      return 3600;  // 1 hour for general queries
    } else if (this.isCreativeWriting(prompt)) {
      return 1800;  // 30 minutes for creative content
    }
    return 7200; // Default 2 hours
  }
  
  isCodeGeneration(prompt) {
    const codeKeywords = ['function', 'class', 'def ', 'import', 'export', 'const', 'let', 'var'];
    return codeKeywords.some(keyword => prompt.toLowerCase().includes(keyword));
  }
  
  isGeneralQuestion(prompt) {
    const questionIndicators = ['?', 'what', 'how', 'why', 'when', 'where', 'who'];
    return questionIndicators.some(indicator => 
      prompt.toLowerCase().includes(indicator)
    );
  }
  
  isCreativeWriting(prompt) {
    const creativeKeywords = ['story', 'poem', 'creative', 'imagine', 'write a', 'generate a'];
    return creativeKeywords.some(keyword => prompt.toLowerCase().includes(keyword));
  }
  
  estimateTokens(prompt, response) {
    // Rough estimation: 1 token ≈ 4 characters
    return Math.ceil((prompt.length + response.length) / 4);
  }
}

export default {
  async fetch(request, env, ctx) {
    const processor = new AIRequestProcessor(env);
    
    // Handle CORS preflight
    if (request.method === 'OPTIONS') {
      return new Response(null, {
        headers: {
          'Access-Control-Allow-Origin': '*',
          'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
          'Access-Control-Allow-Headers': 'Content-Type, Authorization',
          'Access-Control-Max-Age': '86400'
        }
      });
    }
    
    // Route AI proxy requests
    if (request.method === 'POST' && request.url.includes('/ai/proxy/')) {
      return processor.processRequest(request);
    }
    
    // Health check endpoint
    if (request.url.includes('/health')) {
      return new Response('OK', { status: 200 });
    }
    
    // API status endpoint
    if (request.url.includes('/status')) {
      return new Response(JSON.stringify({
        status: 'healthy',
        timestamp: new Date().toISOString(),
        version: '1.0.0'
      }), {
        headers: { 'Content-Type': 'application/json' }
      });
    }
    
    // AI binding test endpoint
    if (request.url.includes('/ai-test')) {
      try {
        const aiBinding = !!env.AI;
        const models = env.AI ? await env.AI.list() : null;
        return new Response(JSON.stringify({
          hasAIBinding: aiBinding,
          availableModels: models
        }), {
          headers: { 'Content-Type': 'application/json' }
        });
      } catch (error) {
        return new Response(JSON.stringify({
          error: error.message
        }), {
          status: 500,
          headers: { 'Content-Type': 'application/json' }
        });
      }
    }
    
    return new Response('Not Found', { status: 404 });
  },
};