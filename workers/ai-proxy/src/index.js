// Cloudflare Workers AI Proxy for Obsidian Bot
import { AIProviders } from './providers.js';
import { CacheManager } from './cache.js';
import { RateLimiter } from './rate-limiter.js';
import { CostOptimizer } from './cost-optimizer.js';
import { Analytics } from './analytics.js';

class AIRequestProcessor {
  constructor(env) {
    this.env = env;
    this.cache = new CacheManager(env.AI_CACHE);
    this.providers = new AIProviders(env);
    this.limiter = new RateLimiter(env.AI_CACHE);
    this.optimizer = new CostOptimizer(env);
    this.analytics = new Analytics(env);
  }

  async processRequest(request) {
    const url = new URL(request.url);
    const startTime = Date.now();
    
    // Extract provider and request details
    const pathParts = url.pathname.split('/');
    const provider = pathParts[pathParts.length - 1];
    const prompt = await request.text();
    const clientIP = request.headers.get('cf-connecting-ip');
    const userAgent = request.headers.get('user-agent');
    
    try {
      // 1. Check cache first for identical prompts
      const cacheKey = await this.generateCacheKey(prompt, provider);
      const cached = await this.cache.get(cacheKey);
      if (cached && !cached.expired) {
        console.log(`Cache hit for provider: ${cached.provider}`);
        await this.analytics.trackCacheHit(provider, prompt.length, cached);
        
        return new Response(cached.data, {
          status: 200,
          headers: {
            'x-ai-provider': cached.provider,
            'x-cache-status': 'hit',
            'x-response-time': `${Date.now() - startTime}ms`
          }
        });
      }
      
      // 2. Rate limiting per provider and IP
      const rateLimitKey = `ai-${provider}-${clientIP}`;
      const isAllowed = await this.limiter.check(rateLimitKey, 1, 60);
      if (!isAllowed) {
        await this.analytics.trackRateLimitHit(provider, clientIP, prompt.length);
        return new Response('Rate limit exceeded. Please try again later.', { 
          status: 429,
          headers: {
            'x-cache-status': 'rate-limited',
            'retry-after': '60'
          }
        });
      }
      
      // 3. Select optimal provider based on current conditions
      const optimalProvider = await this.optimizer.selectProvider(prompt, {
        preferredProvider: provider,
        maxLatency: 2000,
        maxCostPerToken: 0.001,
        clientRegion: request.cf.colo
      });
      
      if (!optimalProvider) {
        return new Response('No available AI providers at the moment.', { 
          status: 503,
          headers: { 'x-error': 'no-providers-available' }
        });
      }
      
      // 4. Execute request with retries and fallbacks
      const response = await this.executeRequestWithRetry(prompt, optimalProvider, request);
      const responseTime = Date.now() - startTime;
      
      // 5. Cache response with intelligent TTL
      const ttl = this.calculateCacheTTL(prompt, optimalProvider);
      await this.cache.set(cacheKey, {
        data: response,
        provider: optimalProvider.name,
        timestamp: Date.now(),
        expired: false
      }, ttl);
      
      // 6. Track analytics
      await this.analytics.trackAIUsage({
        provider: optimalProvider.name,
        promptLength: prompt.length,
        responseLength: response.length,
        tokensUsed: this.estimateTokens(prompt, response),
        responseTime,
        cacheHit: false,
        clientIP,
        userAgent,
        colo: request.cf.colo,
        country: request.cf.country
      });
      
      return new Response(response, {
        status: 200,
        headers: {
          'x-ai-provider': optimalProvider.name,
          'x-cache-status': 'miss',
          'x-response-time': `${responseTime}ms`,
          'x-cache-ttl': `${ttl}s`,
          'x-cost-cents': optimalProvider.costCents
        }
      });
      
    } catch (error) {
      console.error('AI request processing error:', error);
      await this.analytics.trackError(provider, error, clientIP);
      
      return new Response('Internal server error processing AI request.', { 
        status: 500,
        headers: { 'x-error': 'internal-error' }
      });
    }
  }
  
  async executeRequestWithRetry(prompt, provider, originalRequest) {
    const maxRetries = 3;
    const baseDelay = 1000;
    
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
    
    return new Response('Not Found', { status: 404 });
  },
};