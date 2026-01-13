# Developer Training

## 🎓 Developer Training Overview

This training program provides comprehensive onboarding for developers working with enhanced Cloudflare Workers. It covers architecture understanding, development best practices, testing strategies, and extension development.

## 📚 Training Modules

### Module 1: Architecture Fundamentals

#### Learning Objectives
- Understand the enhanced workers architecture
- Identify key components and their responsibilities
- Understand data flow through the system
- Learn about performance optimization strategies

#### Session 1: System Architecture

```yaml
duration: 2 hours
format: Lecture + Q&A

topics:
  - Overview of Enhanced Workers
  - Core Components:
    - Request Handler
    - Analytics Engine
    - Cache Manager
    - Rate Limiter
    - Provider Manager
    - Performance Profiler
  - Data Flow Diagram
  - Design Patterns Used

activities:
  - Architecture review exercise
  - Component mapping exercise
  - Data flow tracing exercise
```

#### Session 2: Performance Optimization

```yaml
duration: 2 hours
format: Workshop

topics:
  - Caching Strategies
    - LRU Cache Implementation
    - TTL Management
    - Cache Analytics
  - Rate Limiting
    - Token Bucket Algorithm
    - Sliding Window Algorithm
    - Adaptive Rate Limiting
  - Provider Management
    - Load Balancing
    - Failover Logic
    - Health Checking

activities:
  - Performance profiling exercise
  - Cache optimization exercise
  - Rate limiting implementation
```

### Module 2: Development Environment

#### Learning Objectives
- Set up local development environment
- Configure Wrangler CLI
- Understand testing frameworks
- Master debugging techniques

#### Session 1: Environment Setup

```yaml
duration: 1 hour
format: Hands-on Lab

prerequisites:
  - Node.js 16+ installed
  - Git installed
  - Code editor (VS Code recommended)

topics:
  - Installing Wrangler CLI
  - Authenticating with Cloudflare
  - Setting up project structure
  - Configuring environment variables
  - Running local development server

hands-on:
  1. Install Wrangler CLI
     ```bash
     npm install -g wrangler
     wrangler login
     ```

  2. Clone repository
     ```bash
     git clone <repository-url>
     cd workers
     npm install
     ```

  3. Configure environment
     ```bash
     cp .env.example .env
     # Edit .env with your configuration
     ```

  4. Run development server
     ```bash
     wrangler dev --env development
     ```

exercises:
  - Complete environment setup checklist
  - Verify local development server is running
  - Test API endpoints with curl or Postman
```

#### Session 2: Testing Framework

```yaml
duration: 2 hours
format: Workshop

topics:
  - Unit Testing
    - Testing framework setup
    - Writing test cases
    - Mocking dependencies
  - Integration Testing
    - API endpoint testing
    - Database testing
    - External service testing
  - End-to-End Testing
    - User flow testing
    - Performance testing
    - Load testing

hands-on:
  - Test Structure:
    ```javascript
    // ai-proxy/src/test/cache.test.js
    import { describe, it, expect, beforeEach } from 'vitest';
    import { CacheManager } from './cache.js';

    describe('CacheManager', () => {
      let cache;

      beforeEach(() => {
        cache = new CacheManager(10);
      });

      it('should store and retrieve values', () => {
        cache.set('key1', 'value1');
        expect(cache.get('key1')).toBe('value1');
      });

      it('should enforce TTL', async () => {
        cache.set('key1', 'value1', 100);
        expect(cache.get('key1')).toBe('value1');
        await new Promise(resolve => setTimeout(resolve, 150));
        expect(cache.get('key1')).toBeUndefined();
      });
    });
    ```

exercises:
  - Write unit tests for CacheManager
  - Write integration tests for API endpoints
  - Create performance benchmark tests
```

### Module 3: Core Features Implementation

#### Learning Objectives
- Understand request handling pipeline
- Implement caching strategies
- Integrate rate limiting
- Manage AI providers

#### Session 1: Request Handling

```yaml
duration: 2 hours
format: Code Walkthrough

topics:
  - Request Lifecycle
  - Middleware Pattern
  - Error Handling
  - Response Formatting

code_example:
  import { RequestHandler } from './request-handler.js';

  export default {
    async fetch(request, env, ctx) {
      const handler = new RequestHandler(env, ctx);

      try {
        // Parse request
        const parsedRequest = await handler.parseRequest(request);

        // Apply rate limiting
        await handler.applyRateLimiting(parsedRequest);

        // Check cache
        const cachedResponse = await handler.checkCache(parsedRequest);
        if (cachedResponse) {
          return cachedResponse;
        }

        // Process request
        const response = await handler.processRequest(parsedRequest);

        // Cache response
        await handler.cacheResponse(parsedRequest, response);

        return response;
      } catch (error) {
        return handler.handleError(error, request);
      }
    }
  };

hands-on:
  - Implement request validation middleware
  - Add custom error handling
  - Create response formatting utility
```

#### Session 2: Caching Implementation

```yaml
duration: 2 hours
format: Workshop

topics:
  - Cache Key Generation
  - TTL Management
  - Eviction Policies
  - Cache Analytics

code_example:
  export class CacheManager {
    constructor(maxSize = 100, defaultTTL = 3600000) {
      this.cache = new Map();
      this.maxSize = maxSize;
      this.defaultTTL = defaultTTL;
      this.metrics = { hits: 0, misses: 0 };
    }

    generateKey(prefix, params) {
      const sorted = Object.keys(params).sort()
        .map(k => `${k}:${params[k]}`).join('|');
      return `${prefix}:${sorted}`;
    }

    set(key, value, ttl = this.defaultTTL) {
      if (this.cache.size >= this.maxSize) {
        this.evictLRU();
      }

      this.cache.set(key, {
        value,
        createdAt: Date.now(),
        ttl,
        lastAccessed: Date.now()
      });
    }

    get(key) {
      const entry = this.cache.get(key);
      if (!entry) {
        this.metrics.misses++;
        return undefined;
      }

      if (Date.now() - entry.createdAt > entry.ttl) {
        this.cache.delete(key);
        this.metrics.misses++;
        return undefined;
      }

      entry.lastAccessed = Date.now();
      this.metrics.hits++;
      return entry.value;
    }

    evictLRU() {
      let oldest = null;
      let oldestKey = null;

      for (const [key, entry] of this.cache.entries()) {
        if (!oldest || entry.lastAccessed < oldest.lastAccessed) {
          oldest = entry;
          oldestKey = key;
        }
      }

      if (oldestKey) {
        this.cache.delete(oldestKey);
      }
    }

    getMetrics() {
      const total = this.metrics.hits + this.metrics.misses;
      return {
        ...this.metrics,
        hitRate: total > 0 ? this.metrics.hits / total : 0,
        size: this.cache.size
      };
    }
  }

exercises:
  - Implement cache key optimization
  - Add cache preloading functionality
  - Create cache analytics dashboard
```

#### Session 3: Rate Limiting

```yaml
duration: 2 hours
format: Workshop

topics:
  - Rate Limiting Algorithms
  - User Identification
  - Bucket Management
  - Rate Limit Analytics

code_example:
  export class RateLimiter {
    constructor(config = {}) {
      this.buckets = new Map();
      this.config = {
        maxRequests: config.maxRequests || 100,
        windowSize: config.windowSize || 60000, // 1 minute
        burstSize: config.burstSize || 20,
        ...config
      };
      this.metrics = { allowed: 0, denied: 0 };
    }

  async checkLimit(userId) {
    const now = Date.now();
    const bucket = this.buckets.get(userId);

    if (!bucket) {
      this.buckets.set(userId, {
        count: 1,
        windowStart: now,
        burstTokens: this.config.burstSize - 1
      });
      this.metrics.allowed++;
      return { allowed: true, remaining: this.config.maxRequests - 1 };
    }

    // Check if window has expired
    if (now - bucket.windowStart > this.config.windowSize) {
      bucket.count = 1;
      bucket.windowStart = now;
      bucket.burstTokens = this.config.burstSize;
      this.metrics.allowed++;
      return { allowed: true, remaining: this.config.maxRequests - 1 };
    }

    // Check burst tokens
    if (bucket.burstTokens > 0) {
      bucket.burstTokens--;
      this.metrics.allowed++;
      return { allowed: true, remaining: bucket.burstTokens };
    }

    // Check rate limit
    if (bucket.count < this.config.maxRequests) {
      bucket.count++;
      this.metrics.allowed++;
      return { allowed: true, remaining: this.config.maxRequests - bucket.count };
    }

    this.metrics.denied++;
    const resetTime = bucket.windowStart + this.config.windowSize;
    return {
      allowed: false,
      remaining: 0,
      resetTime: new Date(resetTime).toISOString()
    };
  }

  getMetrics() {
    const total = this.metrics.allowed + this.metrics.denied;
    return {
      ...this.metrics,
      denialRate: total > 0 ? this.metrics.denied / total : 0,
      activeBuckets: this.buckets.size
    };
  }
}

exercises:
  - Implement adaptive rate limiting
  - Add rate limit bypass for trusted users
  - Create rate limit monitoring dashboard
```

### Module 4: Extension Development

#### Learning Objectives
- Understand extension architecture
- Create custom providers
- Implement custom middleware
- Add custom analytics

#### Session 1: Custom Providers

```yaml
duration: 3 hours
format: Workshop

topics:
  - Provider Interface
  - Provider Registration
  - Health Checking
  - Cost Optimization

code_example:
  // ai-proxy/src/providers/custom-provider.js
  export class CustomProvider {
    constructor(config) {
      this.name = config.name;
      this.apiKey = config.apiKey;
      this.endpoint = config.endpoint;
      this.enabled = config.enabled !== false;
      this.priority = config.priority || 100;
      this.metrics = {
        requests: 0,
        errors: 0,
        totalLatency: 0
      };
    }

    async chat(messages, options = {}) {
      if (!this.enabled) {
        throw new Error('Provider is disabled');
      }

      const startTime = performance.now();
      this.metrics.requests++;

      try {
        const response = await fetch(`${this.endpoint}/chat`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${this.apiKey}`
          },
          body: JSON.stringify({
            messages,
            model: options.model || 'default',
            temperature: options.temperature || 0.7,
            max_tokens: options.maxTokens || 1000
          })
        });

        if (!response.ok) {
          throw new Error(`Provider returned ${response.status}`);
        }

        const data = await response.json();
        const latency = performance.now() - startTime;
        this.metrics.totalLatency += latency;

        return {
          content: data.content,
          model: data.model,
          usage: data.usage,
          latency,
          provider: this.name
        };
      } catch (error) {
        this.metrics.errors++;
        throw new Error(`Provider error: ${error.message}`);
      }
    }

    async healthCheck() {
      try {
        const startTime = performance.now();
        const response = await fetch(`${this.endpoint}/health`, {
          method: 'GET',
          headers: {
            'Authorization': `Bearer ${this.apiKey}`
          },
          signal: AbortSignal.timeout(5000)
        });

        const latency = performance.now() - startTime;

        return {
          healthy: response.ok,
          latency,
          lastChecked: new Date().toISOString()
        };
      } catch (error) {
        return {
          healthy: false,
          error: error.message,
          lastChecked: new Date().toISOString()
        };
      }
    }

    getMetrics() {
      const avgLatency = this.metrics.requests > 0
        ? this.metrics.totalLatency / this.metrics.requests
        : 0;

      return {
        ...this.metrics,
        errorRate: this.metrics.requests > 0
          ? this.metrics.errors / this.metrics.requests
          : 0,
        avgLatency
      };
    }
  }

  // Register custom provider
  export function registerCustomProvider(config) {
    return new CustomProvider(config);
  }

exercises:
  - Create a custom provider for a new AI service
  - Implement provider-specific health checks
  - Add cost tracking to your provider
```

#### Session 2: Custom Middleware

```yaml
duration: 2 hours
format: Workshop

topics:
  - Middleware Pattern
  - Request/Response Transformation
  - Authentication/Authorization
  - Logging and Monitoring

code_example:
  // ai-proxy/src/middleware/custom-middleware.js
  export function createMiddleware(config) {
    return async function middleware(request, env, ctx, next) {
      // Before request processing
      const startTime = performance.now();
      const requestId = generateRequestId();

      // Add request metadata
      request.metadata = {
        requestId,
        startTime,
        userId: extractUserId(request),
        path: new URL(request.url).pathname
      };

      try {
        // Log incoming request
        logRequest(request);

        // Validate request
        const validationResult = await validateRequest(request, config);
        if (!validationResult.valid) {
          return createErrorResponse(validationResult.errors, 400);
        }

        // Process request
        const response = await next(request, env, ctx);

        // Add response headers
        response.headers.set('X-Request-ID', requestId);
        response.headers.set('X-Response-Time', 
          `${(performance.now() - startTime).toFixed(2)}ms`
        );

        // Log response
        logResponse(request, response, startTime);

        return response;
      } catch (error) {
        // Handle error
        const errorResponse = createErrorResponse([error.message], 500);
        
        // Log error
        logError(request, error, startTime);

        return errorResponse;
      }
    };
  }

  function generateRequestId() {
    return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }

  function extractUserId(request) {
    return request.headers.get('X-User-ID') || 'anonymous';
  }

  async validateRequest(request, config) {
    const errors = [];

    // Check required headers
    if (config.requiredHeaders) {
      for (const header of config.requiredHeaders) {
        if (!request.headers.get(header)) {
          errors.push(`Missing required header: ${header}`);
        }
      }
    }

    // Check request size
    if (config.maxRequestSize) {
      const size = parseInt(request.headers.get('Content-Length') || '0');
      if (size > config.maxRequestSize) {
        errors.push(`Request size exceeds limit: ${size}`);
      }
    }

    return { valid: errors.length === 0, errors };
  }

  function createErrorResponse(errors, status) {
    return new Response(JSON.stringify({
      error: errors,
      timestamp: new Date().toISOString()
    }), {
      status,
      headers: { 'Content-Type': 'application/json' }
    });
  }

  function logRequest(request) {
    console.log({
      type: 'request',
      requestId: request.metadata.requestId,
      method: request.method,
      url: request.url,
      userId: request.metadata.userId,
      path: request.metadata.path,
      timestamp: new Date().toISOString()
    });
  }

  function logResponse(request, response, startTime) {
    console.log({
      type: 'response',
      requestId: request.metadata.requestId,
      status: response.status,
      duration: performance.now() - startTime,
      timestamp: new Date().toISOString()
    });
  }

  function logError(request, error, startTime) {
    console.error({
      type: 'error',
      requestId: request.metadata.requestId,
      error: error.message,
      stack: error.stack,
      duration: performance.now() - startTime,
      timestamp: new Date().toISOString()
    });
  }

exercises:
  - Create authentication middleware
  - Implement request validation middleware
  - Add custom logging middleware
```

### Module 5: Advanced Topics

#### Learning Objectives
- Master performance optimization
- Understand security best practices
- Learn about deployment strategies
- Master troubleshooting techniques

#### Session 1: Performance Optimization

```yaml
duration: 2 hours
format: Workshop

topics:
  - Profiling Techniques
  - Bottleneck Identification
  - Optimization Strategies
  - Benchmarking

hands-on:
  - Profile an API endpoint
  - Identify performance bottlenecks
  - Implement optimizations
  - Measure improvement

exercises:
  - Optimize a slow API endpoint
  - Implement caching for expensive operations
  - Add connection pooling
  - Reduce memory usage
```

#### Session 2: Security Best Practices

```yaml
duration: 2 hours
format: Workshop

topics:
  - Input Validation
  - Output Sanitization
  - Authentication/Authorization
  - Rate Limiting
  - Error Handling

hands-on:
  - Conduct security audit
  - Fix identified vulnerabilities
  - Implement security headers
  - Add input validation

exercises:
  - Add input validation to all endpoints
  - Implement authentication mechanism
  - Add rate limiting to prevent abuse
  - Secure sensitive data handling
```

## 📊 Assessment

### Knowledge Check

```yaml
format: Multiple choice + coding exercises

topics:
  - Architecture understanding
  - Component interactions
  - Performance optimization
  - Error handling
  - Testing strategies

passing_score: 80%
```

### Practical Assessment

```yaml
format: Hands-on project

duration: 4 hours

tasks:
  1. Implement a new feature
  2. Write comprehensive tests
  3. Optimize performance
  4. Document implementation

evaluation_criteria:
  - Code quality
  - Test coverage
  - Performance
  - Documentation
  - Best practices adherence
```

## 📖 Additional Resources

### Documentation
- [Architecture Guide](../developer-docs/architecture.md)
- [API Reference](../developer-docs/api-reference.md)
- [Extension Guide](../developer-docs/extension-guide.md)
- [Testing Guide](../developer-docs/testing.md)

### Tools
- [Wrangler CLI](https://developers.cloudflare.com/workers/wrangler/)
- [Vitest](https://vitest.dev/)
- [ESLint](https://eslint.org/)
- [Prettier](https://prettier.io/)

### Best Practices
- [Cloudflare Workers Best Practices](https://developers.cloudflare.com/workers/best-practices/)
- [JavaScript Performance](https://web.dev/fast/)
- [API Design](https://restfulapi.net/)
- [Testing Best Practices](https://martinfowler.com/articles/practical-test-pyramid.html)

## 🔗 Related Documentation

- [Operations Training](./operations-training.md)
- [Performance Workshop](./performance-workshop.md)
- [Extension Guide](../developer-docs/extension-guide.md)
- [Testing Guide](../developer-docs/testing.md)
