# Extension Guide

## 🔌 Extending Enhanced Workers

This guide covers how to extend the enhanced workers with custom functionality, new AI providers, caching strategies, and monitoring capabilities.

## 🏗️ Architecture Overview

The workers are designed with a modular architecture that allows easy extension:

```
Enhanced Workers
├── Core Components (extendable)
│   ├── Analytics Engine
│   ├── Cache Manager
│   ├── Rate Limiter
│   └── Provider Manager
├── Extension Points
│   ├── Custom Providers
│   ├── Cache Strategies
│   ├── Monitoring Plugins
│   └── Middleware
└── Plugin System
    ├── Registration
    ├── Configuration
    └── Lifecycle Management
```

## 🤖 Adding Custom AI Providers

### 1. Implement the AIProvider Interface

Create a new provider class that implements the required interface:

```javascript
// custom-provider.js
import { AIProvider } from './ai/provider.js';

export class CustomAIProvider extends AIProvider {
  constructor(config) {
    super();
    this.apiKey = config.apiKey;
    this.endpoint = config.endpoint;
    this.model = config.model;
  }

  async generateCompletion(ctx, req) {
    try {
      const response = await fetch(this.endpoint, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${this.apiKey}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          prompt: req.prompt,
          max_tokens: req.maxTokens,
          temperature: req.temperature,
          model: req.model || this.model
        })
      });

      if (!response.ok) {
        throw new Error(`Provider error: ${response.status}`);
      }

      const data = await response.json();

      return {
        content: data.choices[0].text,
        modelInfo: {
          name: this.model,
          latency: Date.now() - ctx.startTime,
          accuracy: 0.85, // Provider-specific accuracy
          maxTokens: data.usage.total_tokens,
          rateLimit: 1000,
          concurrency: 10,
          streaming: false,
          enabled: true,
          blocked: false,
          inputCostPerToken: 0.001,
          outputCostPerToken: 0.002,
          maxInputTokens: 4096,
          maxOutputTokens: 4096,
          latencyMsThreshold: 5000,
          accuracyPctThreshold: 80,
          supportsVision: false
        },
        inputTokens: data.usage.prompt_tokens,
        outputTokens: data.usage.completion_tokens,
        finishReason: data.choices[0].finish_reason
      };
    } catch (error) {
      throw new Error(`Custom provider failed: ${error.message}`);
    }
  }

  async streamCompletion(ctx, req) {
    // Implement streaming if supported
    throw new Error('Streaming not implemented for custom provider');
  }

  async checkHealth(ctx) {
    try {
      const response = await fetch(`${this.endpoint}/health`, {
        timeout: 5000
      });
      return response.ok;
    } catch (error) {
      return false;
    }
  }

  getModelInfo() {
    return {
      name: this.model,
      latency: 1000, // Estimated latency
      accuracy: 0.85,
      maxTokens: 4096,
      rateLimit: 1000,
      concurrency: 10,
      streaming: false,
      enabled: true,
      blocked: false,
      inputCostPerToken: 0.001,
      outputCostPerToken: 0.002,
      maxInputTokens: 4096,
      maxOutputTokens: 4096,
      latencyMsThreshold: 5000,
      accuracyPctThreshold: 80,
      supportsVision: false
    };
  }
}
```

### 2. Register the Provider

Add the provider to the provider manager:

```javascript
// In your worker initialization
import { CustomAIProvider } from './custom-provider.js';

export default {
  async fetch(request, env, ctx) {
    // Initialize custom provider
    const customProvider = new CustomAIProvider({
      apiKey: env.CUSTOM_API_KEY,
      endpoint: 'https://api.custom-provider.com/v1',
      model: 'custom-model-v1'
    });

    // Register with provider manager
    const providerManager = new ProviderManager(providers, optimizer, analytics);
    providerManager.addProvider('custom', customProvider);

    // Use in request processing
    // ... rest of handler logic
  }
}
```

### 3. Configure the Provider

Add configuration in `wrangler.toml`:

```toml
[vars]
CUSTOM_API_KEY = "your-custom-api-key"

[ai_proxy.providers.custom]
enabled = true
priority = 3
cost_per_token = 0.0015
max_tokens = 4096
timeout = 10000
```

## 💾 Custom Cache Strategies

### 1. Implement Custom Cache Interface

Create a custom caching strategy:

```javascript
// custom-cache.js
export class PredictiveCache {
  constructor(options = {}) {
    this.cache = new Map();
    this.predictions = new Map();
    this.maxSize = options.maxSize || 1000;
    this.ttl = options.ttl || 3600000; // 1 hour
    this.analytics = options.analytics;
  }

  async get(key) {
    // Check main cache first
    let item = this.cache.get(key);

    if (item && Date.now() < item.expiry) {
      this.analytics?.recordCacheHit(key, item);
      return item.value;
    }

    // Check predictions for pre-warmed content
    const prediction = this.predictions.get(key);
    if (prediction && Date.now() < prediction.expiry) {
      this.analytics?.recordPredictiveHit(key);
      this.cache.set(key, prediction);
      return prediction.value;
    }

    this.analytics?.recordCacheMiss(key);
    return null;
  }

  async set(key, value, options = {}) {
    const ttl = options.ttl || this.ttl;
    const expiry = Date.now() + ttl;

    const cacheItem = {
      value,
      expiry,
      setAt: Date.now(),
      accessCount: 0,
      lastAccessed: Date.now()
    };

    // Evict if necessary
    if (this.cache.size >= this.maxSize) {
      await this.evictItems();
    }

    this.cache.set(key, cacheItem);
    this.analytics?.recordCacheSet(key, ttl);

    // Generate predictions based on access patterns
    await this.generatePredictions(key, value);
  }

  async evictItems() {
    // Implement custom eviction strategy
    const entries = Array.from(this.cache.entries());

    // Sort by access frequency and recency
    entries.sort(([,a], [,b]) => {
      const scoreA = a.accessCount * Math.exp((a.lastAccessed - Date.now()) / this.ttl);
      const scoreB = b.accessCount * Math.exp((b.lastAccessed - Date.now()) / this.ttl);
      return scoreA - scoreB;
    });

    // Remove least valuable items
    const toRemove = Math.ceil(this.maxSize * 0.1); // Remove 10%
    for (let i = 0; i < toRemove && entries.length > 0; i++) {
      const [key] = entries.shift();
      this.cache.delete(key);
      this.analytics?.recordCacheEviction(key);
    }
  }

  async generatePredictions(key, value) {
    // Analyze access patterns to predict future requests
    const patterns = await this.analyzeAccessPatterns(key);

    for (const pattern of patterns) {
      if (pattern.confidence > 0.7) { // High confidence prediction
        this.predictions.set(pattern.key, {
          value: await this.prefetchValue(pattern.key),
          expiry: Date.now() + (this.ttl / 2), // Shorter TTL for predictions
          predictedAt: Date.now(),
          confidence: pattern.confidence
        });
      }
    }
  }

  async analyzeAccessPatterns(key) {
    // Implement pattern analysis logic
    // This could use historical data, user behavior, etc.
    return [
      {
        key: `${key}_related`,
        confidence: 0.8,
        reason: 'Frequently accessed together'
      }
    ];
  }

  async prefetchValue(key) {
    // Implement prefetching logic
    // This could call external APIs, compute derived data, etc.
    return `prefetched-${key}`;
  }

  getStats() {
    return {
      size: this.cache.size,
      maxSize: this.maxSize,
      predictions: this.predictions.size,
      hitRate: this.calculateHitRate(),
      evictionRate: this.calculateEvictionRate()
    };
  }

  calculateHitRate() {
    // Implement hit rate calculation
    const total = this.analytics?.totalRequests || 1;
    const hits = this.analytics?.cacheHits || 0;
    return hits / total;
  }

  calculateEvictionRate() {
    // Implement eviction rate calculation
    const totalSets = this.analytics?.totalSets || 1;
    const evictions = this.analytics?.evictions || 0;
    return evictions / totalSets;
  }
}
```

### 2. Integrate Custom Cache

Use the custom cache in your worker:

```javascript
// In your worker
import { PredictiveCache } from './custom-cache.js';
import { CacheAnalytics } from './cache.js';

export default {
  async fetch(request, env, ctx) {
    // Initialize custom cache with analytics
    const cacheAnalytics = new CacheAnalytics();
    const cache = new PredictiveCache({
      maxSize: 1000,
      ttl: 3600000,
      analytics: cacheAnalytics
    });

    // Use in request processing
    const cacheKey = generateCacheKey(request);

    // Try cache first
    let response = await cache.get(cacheKey);
    if (response) {
      return new Response(response);
    }

    // Generate response
    const result = await processRequest(request);

    // Cache result
    await cache.set(cacheKey, result.content, { ttl: 1800000 });

    return new Response(result.content);
  }
}
```

## 📊 Custom Monitoring Plugins

### 1. Create Monitoring Plugin

Implement a custom monitoring plugin:

```javascript
// custom-monitor.js
export class CustomMonitor {
  constructor(options = {}) {
    this.name = options.name || 'custom-monitor';
    this.interval = options.interval || 30000; // 30 seconds
    this.metrics = new Map();
    this.alerts = [];
    this.subscriptions = new Set();
  }

  // Initialize the monitor
  async initialize(env, ctx) {
    // Set up periodic monitoring
    ctx.waitUntil(this.startMonitoring());

    // Register cleanup
    ctx.waitUntil(this.setupCleanup());
  }

  async startMonitoring() {
    // Start periodic monitoring loop
    const monitor = async () => {
      try {
        await this.collectMetrics();
        await this.checkThresholds();
        await this.notifySubscribers();
      } catch (error) {
        console.error(`${this.name} monitoring error:`, error);
      }
    };

    // Initial run
    await monitor();

    // Set up periodic runs
    setInterval(monitor, this.interval);
  }

  async collectMetrics() {
    // Collect custom metrics
    const metrics = {
      timestamp: Date.now(),
      customMetric1: await this.measureCustomMetric1(),
      customMetric2: await this.measureCustomMetric2(),
      systemLoad: await this.measureSystemLoad(),
      errorRate: await this.measureErrorRate()
    };

    this.metrics.set(Date.now(), metrics);

    // Keep only last hour of metrics
    const cutoff = Date.now() - (60 * 60 * 1000);
    for (const [timestamp] of this.metrics) {
      if (timestamp < cutoff) {
        this.metrics.delete(timestamp);
      }
    }
  }

  async measureCustomMetric1() {
    // Implement custom metric measurement
    // This could be API response times, database queries, etc.
    return Math.random() * 100; // Placeholder
  }

  async measureCustomMetric2() {
    // Another custom metric
    return Math.random() * 50; // Placeholder
  }

  async measureSystemLoad() {
    // Measure system load
    try {
      // Use available system APIs
      return 0.75; // Placeholder
    } catch (error) {
      return 0;
    }
  }

  async measureErrorRate() {
    // Calculate error rate from recent metrics
    const recentMetrics = Array.from(this.metrics.values()).slice(-10);
    if (recentMetrics.length === 0) return 0;

    const errors = recentMetrics.filter(m => m.errorRate > 0.05).length;
    return errors / recentMetrics.length;
  }

  async checkThresholds() {
    const latest = Array.from(this.metrics.values()).pop();
    if (!latest) return;

    // Define thresholds
    const thresholds = {
      customMetric1: { warning: 80, critical: 95 },
      customMetric2: { warning: 40, critical: 45 },
      systemLoad: { warning: 0.8, critical: 0.95 },
      errorRate: { warning: 0.05, critical: 0.1 }
    };

    // Check each metric
    for (const [metric, value] of Object.entries(latest)) {
      if (metric === 'timestamp') continue;

      const threshold = thresholds[metric];
      if (!threshold) continue;

      if (value >= threshold.critical) {
        await this.triggerAlert('critical', metric, value, threshold.critical);
      } else if (value >= threshold.warning) {
        await this.triggerAlert('warning', metric, value, threshold.warning);
      }
    }
  }

  async triggerAlert(level, metric, value, threshold) {
    const alert = {
      id: `alert-${Date.now()}`,
      level,
      metric,
      value,
      threshold,
      timestamp: Date.now(),
      message: `${level.toUpperCase()}: ${metric} is ${value} (threshold: ${threshold})`
    };

    this.alerts.push(alert);

    // Keep only recent alerts
    if (this.alerts.length > 100) {
      this.alerts.shift();
    }

    // Notify subscribers
    await this.notifySubscribers(alert);
  }

  subscribe(callback) {
    this.subscriptions.add(callback);
  }

  unsubscribe(callback) {
    this.subscriptions.delete(callback);
  }

  async notifySubscribers(data) {
    const promises = [];
    for (const callback of this.subscriptions) {
      promises.push(
        Promise.resolve().then(() => callback(data)).catch(error => {
          console.error(`${this.name} subscriber error:`, error);
        })
      );
    }
    await Promise.allSettled(promises);
  }

  getMetrics(options = {}) {
    const { since, limit = 100 } = options;

    let metrics = Array.from(this.metrics.values());

    // Filter by time if specified
    if (since) {
      metrics = metrics.filter(m => m.timestamp >= since);
    }

    // Limit results
    if (metrics.length > limit) {
      metrics = metrics.slice(-limit);
    }

    return {
      name: this.name,
      metrics,
      alerts: this.alerts.slice(-10), // Last 10 alerts
      stats: this.getStats()
    };
  }

  getStats() {
    const metrics = Array.from(this.metrics.values());
    if (metrics.length === 0) {
      return {
        totalMetrics: 0,
        avgCustomMetric1: 0,
        avgCustomMetric2: 0,
        alertsTriggered: 0
      };
    }

    const avgCustomMetric1 = metrics.reduce((sum, m) => sum + m.customMetric1, 0) / metrics.length;
    const avgCustomMetric2 = metrics.reduce((sum, m) => sum + m.customMetric2, 0) / metrics.length;

    return {
      totalMetrics: metrics.length,
      avgCustomMetric1,
      avgCustomMetric2,
      alertsTriggered: this.alerts.length,
      uptime: Date.now() - metrics[0].timestamp
    };
  }

  async setupCleanup() {
    // Set up cleanup on worker termination
    return new Promise((resolve) => {
      // Cleanup logic if needed
      resolve();
    });
  }
}
```

### 2. Integrate Monitoring Plugin

Add the custom monitor to your worker:

```javascript
// In your worker
import { CustomMonitor } from './custom-monitor.js';

export default {
  async fetch(request, env, ctx) {
    // Initialize custom monitor
    const monitor = new CustomMonitor({
      name: 'enhanced-monitor',
      interval: 15000 // 15 seconds
    });

    // Initialize the monitor
    await monitor.initialize(env, ctx);

    // Subscribe to alerts
    monitor.subscribe((alert) => {
      console.log('Alert received:', alert);

      // Send to external monitoring system
      // await sendToMonitoringSystem(alert);
    });

    // Add monitoring endpoint
    if (request.url.pathname === '/custom-metrics') {
      const metrics = monitor.getMetrics({
        since: Date.now() - (60 * 60 * 1000), // Last hour
        limit: 50
      });

      return new Response(JSON.stringify(metrics), {
        headers: { 'Content-Type': 'application/json' }
      });
    }

    // Continue with normal request processing
    // ... rest of handler logic
  }
}
```

## 🔧 Custom Middleware

### 1. Create Middleware Function

Implement custom middleware:

```javascript
// custom-middleware.js
export class CustomMiddleware {
  constructor(options = {}) {
    this.name = options.name || 'custom-middleware';
    this.priority = options.priority || 10;
    this.enabled = options.enabled !== false;
  }

  async execute(request, env, ctx, next) {
    if (!this.enabled) {
      return next(request, env, ctx);
    }

    // Pre-processing
    const startTime = Date.now();
    const requestId = this.generateRequestId();

    // Add custom headers
    const modifiedRequest = new Request(request);
    modifiedRequest.headers.set('X-Request-ID', requestId);
    modifiedRequest.headers.set('X-Middleware', this.name);

    // Log request
    console.log(`[${this.name}] Processing request ${requestId}`);

    try {
      // Execute next middleware/handler
      const response = await next(modifiedRequest, env, ctx);

      // Post-processing
      const processingTime = Date.now() - startTime;

      // Add response headers
      const modifiedResponse = new Response(response.body, response);
      modifiedResponse.headers.set('X-Processing-Time', `${processingTime}ms`);
      modifiedResponse.headers.set('X-Middleware-Processed', this.name);

      // Log response
      console.log(`[${this.name}] Completed request ${requestId} in ${processingTime}ms`);

      // Custom logic based on response
      if (response.status >= 400) {
        await this.handleErrorResponse(modifiedRequest, modifiedResponse);
      }

      return modifiedResponse;

    } catch (error) {
      // Error handling
      console.error(`[${this.name}] Error processing request ${requestId}:`, error);

      // Custom error handling
      await this.handleError(error, modifiedRequest);

      // Return error response or re-throw
      throw error;
    }
  }

  generateRequestId() {
    return `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  async handleErrorResponse(request, response) {
    // Custom logic for error responses
    // Could log to external system, send alerts, etc.
    console.warn(`[${this.name}] Error response: ${response.status} for ${request.url}`);
  }

  async handleError(error, request) {
    // Custom error handling logic
    // Could send to error tracking service, etc.
    console.error(`[${this.name}] Error:`, error);
  }

  getStats() {
    // Return middleware statistics
    return {
      name: this.name,
      priority: this.priority,
      enabled: this.enabled,
      processedRequests: 0, // Would track in real implementation
      averageProcessingTime: 0,
      errorsHandled: 0
    };
  }
}

// Factory function for creating middleware
export function createCustomMiddleware(options) {
  return new CustomMiddleware(options);
}

// Middleware registry
export class MiddlewareRegistry {
  constructor() {
    this.middlewares = new Map();
  }

  register(name, middleware) {
    this.middlewares.set(name, middleware);
  }

  get(name) {
    return this.middlewares.get(name);
  }

  getAll() {
    return Array.from(this.middlewares.values());
  }

  getSorted() {
    // Sort by priority (lower number = higher priority)
    return Array.from(this.middlewares.values())
      .sort((a, b) => a.priority - b.priority);
  }
}
```

### 2. Use Custom Middleware

Integrate middleware into your worker:

```javascript
// In your worker
import { createCustomMiddleware, MiddlewareRegistry } from './custom-middleware.js';

export default {
  async fetch(request, env, ctx) {
    // Initialize middleware registry
    const registry = new MiddlewareRegistry();

    // Register custom middleware
    registry.register('logging', createCustomMiddleware({
      name: 'request-logger',
      priority: 1
    }));

    registry.register('auth', createCustomMiddleware({
      name: 'auth-checker',
      priority: 2
    }));

    registry.register('custom', createCustomMiddleware({
      name: 'business-logic',
      priority: 10
    }));

    // Create middleware chain
    const middlewares = registry.getSorted();

    // Execute middleware chain
    let currentRequest = request;

    for (const middleware of middlewares) {
      currentRequest = await middleware.execute(currentRequest, env, ctx, async (req, env, ctx) => {
        // This would be the final handler
        return await processFinalRequest(req, env, ctx);
      });
    }

    return currentRequest;
  }
}
```

## 🔌 Plugin System

### 1. Create Plugin Interface

Define a plugin interface:

```javascript
// plugin-interface.js
export class PluginInterface {
  constructor() {
    this.name = 'base-plugin';
    this.version = '1.0.0';
    this.description = 'Base plugin interface';
  }

  // Lifecycle methods
  async initialize(env, ctx) {
    // Called when plugin is loaded
  }

  async destroy() {
    // Called when plugin is unloaded
  }

  // Hook methods (extend as needed)
  async onRequest(request, env, ctx) {
    // Called for each request
    return request;
  }

  async onResponse(response, request, env, ctx) {
    // Called for each response
    return response;
  }

  async onError(error, request, env, ctx) {
    // Called when errors occur
    return error;
  }

  // Configuration
  getConfigSchema() {
    // Return JSON schema for configuration
    return {};
  }

  // Metadata
  getMetadata() {
    return {
      name: this.name,
      version: this.version,
      description: this.description,
      hooks: ['onRequest', 'onResponse', 'onError']
    };
  }
}

// Plugin manager
export class PluginManager {
  constructor() {
    this.plugins = new Map();
    this.enabledPlugins = new Set();
  }

  register(name, pluginClass, config = {}) {
    this.plugins.set(name, { pluginClass, config });
  }

  async enable(name, env, ctx) {
    const pluginDef = this.plugins.get(name);
    if (!pluginDef) {
      throw new Error(`Plugin ${name} not found`);
    }

    const plugin = new pluginDef.pluginClass();
    await plugin.initialize(env, ctx);

    this.enabledPlugins.add(name);
    return plugin;
  }

  async disable(name) {
    const plugin = this.enabledPlugins.get(name);
    if (plugin) {
      await plugin.destroy();
      this.enabledPlugins.delete(name);
    }
  }

  async executeHook(hookName, ...args) {
    const results = [];

    for (const plugin of this.enabledPlugins.values()) {
      if (typeof plugin[hookName] === 'function') {
        try {
          const result = await plugin[hookName](...args);
          results.push(result);
        } catch (error) {
          console.error(`Plugin error in ${hookName}:`, error);
        }
      }
    }

    return results;
  }

  getStats() {
    return {
      totalPlugins: this.plugins.size,
      enabledPlugins: this.enabledPlugins.size,
      disabledPlugins: this.plugins.size - this.enabledPlugins.size
    };
  }
}
```

### 2. Create Custom Plugin

Implement a custom plugin:

```javascript
// custom-plugin.js
import { PluginInterface } from './plugin-interface.js';

export class CustomPlugin extends PluginInterface {
  constructor() {
    super();
    this.name = 'custom-analytics';
    this.version = '1.0.0';
    this.description = 'Custom analytics and monitoring plugin';
    this.requestCount = 0;
    this.errorCount = 0;
  }

  async initialize(env, ctx) {
    console.log(`[${this.name}] Initializing custom plugin`);

    // Set up periodic reporting
    ctx.waitUntil(this.startReporting());

    // Load configuration
    this.config = env.CUSTOM_PLUGIN_CONFIG || {};
  }

  async destroy() {
    console.log(`[${this.name}] Destroying custom plugin`);
    // Cleanup resources
  }

  async onRequest(request, env, ctx) {
    this.requestCount++;

    // Add custom headers
    const modifiedRequest = new Request(request);
    modifiedRequest.headers.set('X-Custom-Plugin', this.name);
    modifiedRequest.headers.set('X-Request-Number', this.requestCount.toString());

    // Custom request processing
    console.log(`[${this.name}] Processing request ${this.requestCount}: ${request.url}`);

    return modifiedRequest;
  }

  async onResponse(response, request, env, ctx) {
    // Add custom response headers
    const modifiedResponse = new Response(response.body, response);
    modifiedResponse.headers.set('X-Custom-Processed', 'true');
    modifiedResponse.headers.set('X-Request-Total', this.requestCount.toString());

    // Custom response processing
    if (response.status >= 400) {
      console.warn(`[${this.name}] Error response: ${response.status}`);
    }

    return modifiedResponse;
  }

  async onError(error, request, env, ctx) {
    this.errorCount++;

    // Custom error handling
    console.error(`[${this.name}] Error ${this.errorCount}:`, error);

    // Could send to external error tracking
    // await this.reportError(error, request);

    return error; // Return modified error or original
  }

  async startReporting() {
    // Periodic reporting
    setInterval(() => {
      console.log(`[${this.name}] Stats: ${this.requestCount} requests, ${this.errorCount} errors`);
    }, 60000); // Every minute
  }

  getConfigSchema() {
    return {
      type: 'object',
      properties: {
        enableReporting: {
          type: 'boolean',
          default: true,
          description: 'Enable periodic reporting'
        },
        reportInterval: {
          type: 'number',
          default: 60000,
          description: 'Report interval in milliseconds'
        }
      }
    };
  }
}
```

### 3. Use Plugin System

Integrate plugins into your worker:

```javascript
// In your worker
import { PluginManager } from './plugin-interface.js';
import { CustomPlugin } from './custom-plugin.js';

export default {
  async fetch(request, env, ctx) {
    // Initialize plugin manager
    const pluginManager = new PluginManager();

    // Register plugins
    pluginManager.register('custom-analytics', CustomPlugin, {
      enableReporting: true,
      reportInterval: 30000
    });

    // Enable plugins
    await pluginManager.enable('custom-analytics', env, ctx);

    // Execute request hooks
    let modifiedRequest = request;
    const requestResults = await pluginManager.executeHook('onRequest', request, env, ctx);
    if (requestResults.length > 0) {
      modifiedRequest = requestResults[requestResults.length - 1]; // Use last result
    }

    // Process request
    let response;
    try {
      response = await processRequest(modifiedRequest, env, ctx);
    } catch (error) {
      // Execute error hooks
      const errorResults = await pluginManager.executeHook('onError', error, request, env, ctx);
      throw errorResults.length > 0 ? errorResults[errorResults.length - 1] : error;
    }

    // Execute response hooks
    const responseResults = await pluginManager.executeHook('onResponse', response, modifiedRequest, env, ctx);
    if (responseResults.length > 0) {
      response = responseResults[responseResults.length - 1]; // Use last result
    }

    return response;
  }
}
```

## 🎯 Best Practices for Extensions

### Performance Considerations
- Minimize synchronous operations in hot paths
- Use efficient data structures for caching
- Implement proper cleanup for long-lived objects
- Monitor memory usage in extensions

### Error Handling
- Always implement proper error boundaries
- Provide meaningful error messages
- Log errors with sufficient context
- Implement graceful degradation

### Configuration Management
- Use environment variables for sensitive data
- Provide sensible defaults
- Validate configuration on startup
- Document all configuration options

### Testing Extensions
- Write unit tests for extension logic
- Test integration with core systems
- Verify performance impact
- Test error scenarios thoroughly

### Documentation
- Document all extension points
- Provide examples and usage patterns
- Include configuration schemas
- Update API documentation

### Security Considerations
- Validate all inputs thoroughly
- Implement proper access controls
- Avoid exposing sensitive information
- Follow security best practices

## 📚 Examples

### Complete Custom Provider Example

See the `examples/` directory for complete working examples of:
- Custom AI provider implementations
- Advanced caching strategies
- Monitoring plugins
- Middleware extensions
- Plugin system usage

### Quick Start Templates

Use these templates to quickly create new extensions:

```bash
# Create new provider
npx wrangler generate provider-template --type=ai-provider

# Create new middleware
npx wrangler generate middleware-template --type=request-middleware

# Create new plugin
npx wrangler generate plugin-template --type=monitoring-plugin
```

This extension guide provides the foundation for building powerful, customized worker functionality while maintaining compatibility with the core enhanced worker system.