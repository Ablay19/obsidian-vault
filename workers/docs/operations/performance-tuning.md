# Performance Tuning

## ⚡ Performance Tuning Overview

This guide covers comprehensive performance optimization techniques for enhanced Cloudflare Workers, including bottleneck identification, optimization strategies, and performance monitoring. Proper performance tuning ensures optimal response times, cost efficiency, and user experience.

## 🎯 Performance Goals

### Target Metrics

```yaml
performance_targets:
  response_time:
    p50: "<50ms"
    p95: "<200ms"
    p99: "<500ms"
  
  throughput:
    requests_per_second: ">1000"
    concurrent_connections: ">100"
  
  efficiency:
    cpu_usage: "<50%"
    memory_usage: "<10MB"
    cache_hit_rate: ">80%"
  
  cost:
    cost_per_request: "<$0.0001"
    monthly_cost: "<$100"
```

## 🔍 Performance Profiling

### 1. Request Timing Analysis

```javascript
// ai-proxy/src/profiler.js

export class RequestProfiler {
  constructor() {
    this.timings = {};
  }

  start(label) {
    this.timings[label] = {
      start: performance.now(),
      end: null,
      duration: null
    };
  }

  end(label) {
    if (this.timings[label]) {
      this.timings[label].end = performance.now();
      this.timings[label].duration = 
        this.timings[label].end - this.timings[label].start;
    }
  }

  getProfile() {
    const profile = {};
    for (const [label, timing] of Object.entries(this.timings)) {
      profile[label] = {
        duration: timing.duration,
        percent: this.calculatePercentage(timing.duration)
      };
    }
    return profile;
  }

  calculatePercentage(duration) {
    const total = Object.values(this.timings)
      .filter(t => t.duration)
      .reduce((sum, t) => sum + t.duration, 0);
    return total > 0 ? (duration / total * 100).toFixed(2) : 0;
  }

  export(format = 'json') {
    if (format === 'json') {
      return JSON.stringify(this.getProfile(), null, 2);
    } else if (format === 'text') {
      let output = '=== Request Profile ===\n';
      for (const [label, timing] of Object.entries(this.getProfile())) {
        output += `${label}: ${timing.duration.toFixed(2)}ms (${timing.percent}%)\n`;
      }
      return output;
    }
  }
}
```

### 2. Memory Usage Tracking

```javascript
// ai-proxy/src/memory-profiler.js

export class MemoryProfiler {
  constructor() {
    this.snapshots = [];
    this.maxSnapshots = 100;
  }

  takeSnapshot(label) {
    const snapshot = {
      label,
      timestamp: Date.now(),
      memory: {
        used: performance.memory?.usedJSHeapSize || 0,
        total: performance.memory?.totalJSHeapSize || 0,
        limit: performance.memory?.jsHeapSizeLimit || 0
      }
    };

    this.snapshots.push(snapshot);
    if (this.snapshots.length > this.maxSnapshots) {
      this.snapshots.shift();
    }

    return snapshot;
  }

  getMemoryTrend() {
    if (this.snapshots.length < 2) return null;

    const first = this.snapshots[0];
    const last = this.snapshots[this.snapshots.length - 1];
    const timeDiff = last.timestamp - first.timestamp;
    const memoryDiff = last.memory.used - first.memory.used;

    return {
      startMemory: first.memory.used,
      endMemory: last.memory.used,
      memoryDiff,
      timeDiff,
      memoryPerMs: memoryDiff / timeDiff,
      trend: memoryDiff > 0 ? 'increasing' : 'decreasing'
    };
  }

  checkMemoryLimit(limit) {
    const latest = this.snapshots[this.snapshots.length - 1];
    return {
      withinLimit: latest.memory.used < limit,
      usage: latest.memory.used,
      limit,
      percentUsed: (latest.memory.used / limit * 100).toFixed(2)
    };
  }
}
```

### 3. CPU Usage Monitoring

```javascript
// ai-proxy/src/cpu-profiler.js

export class CPUProfiler {
  constructor() {
    this.cpuSamples = [];
    this.maxSamples = 1000;
  }

  startSampling(interval = 100) {
    this.intervalId = setInterval(() => {
      const sample = {
        timestamp: Date.now(),
        stackTrace: this.getStackTrace()
      };
      this.cpuSamples.push(sample);
      if (this.cpuSamples.length > this.maxSamples) {
        this.cpuSamples.shift();
      }
    }, interval);
  }

  stopSampling() {
    if (this.intervalId) {
      clearInterval(this.intervalId);
      this.intervalId = null;
    }
  }

  getStackTrace() {
    // This would use performance profiling APIs
    return new Error().stack;
  }

  analyzeHotspots() {
    const hotspots = new Map();

    for (const sample of this.cpuSamples) {
      const functions = this.parseStackTrace(sample.stackTrace);
      for (const func of functions) {
        const count = hotspots.get(func) || 0;
        hotspots.set(func, count + 1);
      }
    }

    // Convert to array and sort by frequency
    return Array.from(hotspots.entries())
      .map(([func, count]) => ({ function: func, count, percent: (count / this.cpuSamples.length * 100).toFixed(2) }))
      .sort((a, b) => b.count - a.count);
  }

  parseStackTrace(stackTrace) {
    // Parse stack trace to extract function names
    return stackTrace.split('\n')
      .filter(line => line.trim())
      .map(line => this.extractFunctionName(line));
  }

  extractFunctionName(line) {
    // Extract function name from stack trace line
    const match = line.match(/at (\w+)/);
    return match ? match[1] : 'anonymous';
  }
}
```

## 🚀 Optimization Strategies

### 1. Caching Optimization

#### Cache Key Design

```javascript
// ai-proxy/src/cache-optimizer.js

export class CacheOptimizer {
  constructor() {
    this.cache = new Map();
    this.cacheMetrics = {
      hits: 0,
      misses: 0,
      evictions: 0,
      size: 0
    };
  }

  generateCacheKey(prefix, params) {
    // Create a deterministic cache key
    const sortedParams = Object.keys(params)
      .sort()
      .map(key => `${key}:${params[key]}`)
      .join('|');

    return `${prefix}:${sortedParams}`;
  }

  optimizeCacheKey(key) {
    // Shorten cache keys for better performance
    const hash = this.simpleHash(key);
    return key.length > 50 ? `${key.substring(0, 25)}...${hash}` : key;
  }

  simpleHash(str) {
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
      const char = str.charCodeAt(i);
      hash = ((hash << 5) - hash) + char;
      hash = hash & hash;
    }
    return Math.abs(hash).toString(16);
  }

  get(key) {
    const optimizedKey = this.optimizeCacheKey(key);
    const value = this.cache.get(optimizedKey);

    if (value !== undefined) {
      this.cacheMetrics.hits++;
      return value;
    }

    this.cacheMetrics.misses++;
    return undefined;
  }

  set(key, value, ttl = 3600000) {
    const optimizedKey = this.optimizeCacheKey(key);
    
    // Evict oldest entries if cache is full
    if (this.cache.size >= this.maxCacheSize) {
      this.evictOldest();
    }

    this.cache.set(optimizedKey, {
      value,
      createdAt: Date.now(),
      ttl,
      accessCount: 1,
      lastAccessed: Date.now()
    });

    this.cacheMetrics.size = this.cache.size;
  }

  evictOldest() {
    let oldestKey = null;
    let oldestTime = Infinity;

    for (const [key, data] of this.cache.entries()) {
      if (data.lastAccessed < oldestTime) {
        oldestTime = data.lastAccessed;
        oldestKey = key;
      }
    }

    if (oldestKey) {
      this.cache.delete(oldestKey);
      this.cacheMetrics.evictions++;
    }
  }
}
```

#### Cache Preloading

```javascript
// Preload frequently accessed data
export async function preloadCache(context) {
  const preloadData = [
    { key: 'config', ttl: 3600000 },
    { key: 'popular_queries', ttl: 1800000 },
    { key: 'user_stats', ttl: 300000 }
  ];

  for (const item of preloadData) {
    try {
      const data = await fetchFromSource(item.key);
      context.cache.set(item.key, data, item.ttl);
    } catch (error) {
      console.error(`Failed to preload cache for ${item.key}:`, error);
    }
  }
}
```

### 2. Request Optimization

#### Request Batching

```javascript
// ai-proxy/src/request-batcher.js

export class RequestBatcher {
  constructor(maxBatchSize = 10, maxWaitTime = 100) {
    this.maxBatchSize = maxBatchSize;
    this.maxWaitTime = maxWaitTime;
    this.batches = new Map();
  }

  async batchRequest(type, params) {
    const batchId = this.getBatchId(type, params);

    return new Promise((resolve, reject) => {
      if (!this.batches.has(batchId)) {
        this.batches.set(batchId, {
          requests: [],
          timer: null
        });

        // Start timer for this batch
        this.batches.get(batchId).timer = setTimeout(() => {
          this.flushBatch(batchId);
        }, this.maxWaitTime);
      }

      const batch = this.batches.get(batchId);
      batch.requests.push({ params, resolve, reject });

      // Flush immediately if batch is full
      if (batch.requests.length >= this.maxBatchSize) {
        clearTimeout(batch.timer);
        this.flushBatch(batchId);
      }
    });
  }

  getBatchId(type, params) {
    // Group similar requests together
    return `${type}:${Object.keys(params).sort().join(',')}`;
  }

  async flushBatch(batchId) {
    const batch = this.batches.get(batchId);
    if (!batch) return;

    this.batches.delete(batchId);

    try {
      const response = await this.executeBatch(batch.requests);
      batch.requests.forEach((request, index) => {
        request.resolve(response[index]);
      });
    } catch (error) {
      batch.requests.forEach(request => {
        request.reject(error);
      });
    }
  }

  async executeBatch(requests) {
    // Implement batch execution logic
    return Promise.all(
      requests.map(request => this.executeRequest(request.params))
    );
  }
}
```

#### Request Deduplication

```javascript
// ai-proxy/src/request-deduplicator.js

export class RequestDeduplicator {
  constructor() {
    this.pendingRequests = new Map();
  }

  async deduplicateRequest(key, requestFn) {
    // Check if identical request is in progress
    if (this.pendingRequests.has(key)) {
      return this.pendingRequests.get(key);
    }

    // Create new request
    const promise = requestFn()
      .finally(() => {
        this.pendingRequests.delete(key);
      });

    this.pendingRequests.set(key, promise);
    return promise;
  }
}
```

### 3. Data Compression

```javascript
// ai-proxy/src/compression.js

export class DataCompressor {
  constructor() {
    this.compressionStats = {
      originalSize: 0,
      compressedSize: 0,
      ratio: 0
    };
  }

  async compress(data, algorithm = 'gzip') {
    if (algorithm === 'gzip') {
      return this.gzipCompress(data);
    } else if (algorithm === 'brotli') {
      return this.brotliCompress(data);
    }
    return data;
  }

  async gzipCompress(data) {
    const str = typeof data === 'string' ? data : JSON.stringify(data);
    const compressed = new Response(str).compressGzip();
    const compressedText = await compressed.text();

    this.updateStats(str.length, compressedText.length);
    return compressedText;
  }

  async brotliCompress(data) {
    const str = typeof data === 'string' ? data : JSON.stringify(data);
    // Use Brotli compression if available
    const compressed = new Response(str).compressBrotli();
    const compressedText = await compressed.text();

    this.updateStats(str.length, compressedText.length);
    return compressedText;
  }

  updateStats(originalSize, compressedSize) {
    this.compressionStats.originalSize += originalSize;
    this.compressionStats.compressedSize += compressedSize;
    this.compressionStats.ratio = 
      (1 - this.compressionStats.compressedSize / this.compressionStats.originalSize) * 100;
  }

  getCompressionRatio() {
    return {
      originalSize: this.compressionStats.originalSize,
      compressedSize: this.compressionStats.compressedSize,
      ratio: this.compressionStats.ratio.toFixed(2) + '%'
    };
  }
}
```

### 4. Connection Pooling

```javascript
// ai-proxy/src/connection-pool.js

export class ConnectionPool {
  constructor(maxConnections = 10) {
    this.maxConnections = maxConnections;
    this.activeConnections = new Map();
    this.idleConnections = new Map();
  }

  async getConnection(url) {
    // Check for idle connection
    if (this.idleConnections.has(url)) {
      const connection = this.idleConnections.get(url);
      this.idleConnections.delete(url);
      this.activeConnections.set(url, connection);
      return connection;
    }

    // Create new connection if under limit
    if (this.activeConnections.size < this.maxConnections) {
      const connection = await this.createConnection(url);
      this.activeConnections.set(url, connection);
      return connection;
    }

    // Wait for connection to become available
    return await this.waitForConnection(url);
  }

  releaseConnection(url) {
    if (this.activeConnections.has(url)) {
      const connection = this.activeConnections.get(url);
      this.activeConnections.delete(url);
      this.idleConnections.set(url, connection);
    }
  }

  async createConnection(url) {
    // Create connection to URL
    return new Promise((resolve) => {
      // Simulate connection creation
      setTimeout(() => resolve({ url, connected: true }), 100);
    });
  }

  async waitForConnection(url) {
    return new Promise((resolve) => {
      const checkInterval = setInterval(() => {
        if (this.activeConnections.size < this.maxConnections) {
          clearInterval(checkInterval);
          this.getConnection(url).then(resolve);
        }
      }, 100);
    });
  }
}
```

## 📊 Performance Benchmarks

### Benchmark Suite

```javascript
// ai-proxy/src/benchmark.js

export class BenchmarkRunner {
  constructor() {
    this.results = [];
  }

  async runBenchmark(name, fn, iterations = 100) {
    const times = [];

    // Warm-up run
    await fn();

    // Benchmark iterations
    for (let i = 0; i < iterations; i++) {
      const start = performance.now();
      await fn();
      const end = performance.now();
      times.push(end - start);
    }

    const stats = this.calculateStats(times);
    this.results.push({
      name,
      iterations,
      ...stats
    });

    return stats;
  }

  calculateStats(times) {
    const sorted = times.slice().sort((a, b) => a - b);
    const sum = times.reduce((a, b) => a + b, 0);
    const mean = sum / times.length;

    return {
      min: Math.min(...times),
      max: Math.max(...times),
      mean,
      median: sorted[Math.floor(sorted.length / 2)],
      p90: sorted[Math.floor(sorted.length * 0.9)],
      p95: sorted[Math.floor(sorted.length * 0.95)],
      p99: sorted[Math.floor(sorted.length * 0.99)],
      standardDeviation: this.calculateStandardDeviation(times, mean)
    };
  }

  calculateStandardDeviation(times, mean) {
    const squaredDiffs = times.map(time => Math.pow(time - mean, 2));
    const avgSquaredDiff = squaredDiffs.reduce((a, b) => a + b, 0) / times.length;
    return Math.sqrt(avgSquaredDiff);
  }

  printResults() {
    console.table(this.results);
  }

  exportResults(format = 'json') {
    if (format === 'json') {
      return JSON.stringify(this.results, null, 2);
    } else if (format === 'csv') {
      const headers = ['name', 'iterations', 'min', 'max', 'mean', 'median', 'p95', 'p99'];
      const rows = this.results.map(r => 
        headers.map(h => r[h]).join(',')
      );
      return [headers.join(','), ...rows].join('\n');
    }
  }
}
```

### Load Testing

```javascript
// ai-proxy/src/load-test.js

export class LoadTester {
  constructor(url, maxConcurrent = 100) {
    this.url = url;
    this.maxConcurrent = maxConcurrent;
    this.results = [];
  }

  async runLoadTest(duration = 60000) {
    const startTime = Date.now();
    const promises = [];

    while (Date.now() - startTime < duration) {
      if (promises.length < this.maxConcurrent) {
        promises.push(this.makeRequest());
      } else {
        await Promise.race(promises);
        promises.splice(promises.indexOf(Promise.race(promises)), 1);
      }
    }

    await Promise.all(promises);
    return this.analyzeResults();
  }

  async makeRequest() {
    const start = performance.now();
    try {
      const response = await fetch(this.url);
      const end = performance.now();
      this.results.push({
        success: response.ok,
        status: response.status,
        duration: end - start,
        timestamp: start
      });
    } catch (error) {
      const end = performance.now();
      this.results.push({
        success: false,
        error: error.message,
        duration: end - start,
        timestamp: start
      });
    }
  }

  analyzeResults() {
    const successful = this.results.filter(r => r.success);
    const failed = this.results.filter(r => !r.success);
    const durations = successful.map(r => r.duration);

    return {
      totalRequests: this.results.length,
      successfulRequests: successful.length,
      failedRequests: failed.length,
      successRate: (successful.length / this.results.length * 100).toFixed(2) + '%',
      minDuration: Math.min(...durations),
      maxDuration: Math.max(...durations),
      meanDuration: durations.reduce((a, b) => a + b, 0) / durations.length,
      requestsPerSecond: this.results.length / ((this.results[this.results.length - 1].timestamp - this.results[0].timestamp) / 1000)
    };
  }
}
```

## 🔧 Optimization Checklist

### Cache Optimization
- [ ] Cache hit rate >80%
- [ ] Cache key optimization implemented
- [ ] TTL values tuned appropriately
- [ ] Cache preloading configured
- [ ] Cache eviction policy optimized

### Request Optimization
- [ ] Request batching implemented
- [ ] Request deduplication active
- [ ] Connection pooling configured
- [ ] Keep-alive connections enabled
- [ ] Request compression enabled

### Data Optimization
- [ ] Response compression enabled
- [ ] JSON response size minimized
- [ ] Unnecessary fields removed
- [ ] Data pagination implemented
- [ ] Streaming responses for large data

### Code Optimization
- [ ] Hotspots identified and optimized
- [ ] Memory leaks addressed
- [ ] Unnecessary computations removed
- [ ] Algorithms optimized
- [ ] Code minified for production

### Monitoring
- [ ] Performance metrics tracked
- [ ] Profiling enabled
- [ ] Bottlenecks identified
- [ ] Alert thresholds set
- [ ] Regular performance reviews scheduled

## 📈 Performance Optimization Workflow

```javascript
// ai-proxy/src/performance-optimizer.js

export class PerformanceOptimizer {
  constructor(context) {
    this.context = context;
    this.profiler = new RequestProfiler();
    this.memoryProfiler = new MemoryProfiler();
    this.cpuProfiler = new CPUProfiler();
  }

  async analyzePerformance() {
    const analysis = {
      requestTiming: await this.analyzeRequestTiming(),
      memoryUsage: await this.analyzeMemoryUsage(),
      cpuUsage: await this.analyzeCpuUsage(),
      bottlenecks: this.identifyBottlenecks(),
      recommendations: this.generateRecommendations()
    };

    return analysis;
  }

  async analyzeRequestTiming() {
    // Analyze request timing from profiler data
    const profile = this.profiler.getProfile();
    
    return {
      slowestOperation: Object.entries(profile)
        .sort((a, b) => b[1].duration - a[1].duration)[0],
      totalDuration: Object.values(profile).reduce((sum, t) => sum + t.duration, 0),
      operationBreakdown: profile
    };
  }

  async analyzeMemoryUsage() {
    // Analyze memory usage trends
    const trend = this.memoryProfiler.getMemoryTrend();
    
    return {
      currentUsage: this.memoryProcessor.getCurrentUsage(),
      trend: trend?.trend,
      growthRate: trend?.memoryPerMs,
      memoryLeaks: this.detectMemoryLeaks()
    };
  }

  async analyzeCpuUsage() {
    // Analyze CPU hotspots
    const hotspots = this.cpuProfiler.analyzeHotspots();
    
    return {
      topHotspots: hotspots.slice(0, 10),
      cpuIntensiveFunctions: hotspots.filter(h => h.percent > 10)
    };
  }

  identifyBottlenecks() {
    // Identify performance bottlenecks
    return {
      cache: this.identifyCacheBottlenecks(),
      network: this.identifyNetworkBottlenecks(),
      computation: this.identifyComputationBottlenecks()
    };
  }

  generateRecommendations() {
    // Generate optimization recommendations
    return [
      {
        category: 'caching',
        priority: 'high',
        recommendation: 'Increase cache TTL for frequently accessed data'
      },
      {
        category: 'network',
        priority: 'medium',
        recommendation: 'Enable response compression'
      },
      {
        category: 'computation',
        priority: 'low',
        recommendation: 'Optimize JSON parsing in hot path'
      }
    ];
  }
}
```

## 🔗 Related Documentation

- [Monitoring Setup](./monitoring.md)
- [Deployment Guide](./deployment.md)
- [Performance Optimization Guide](../../docs/guides/performance-optimization.md)
- [Troubleshooting Guide](../user-guides/troubleshooting.md)
- [API Reference](../developer-docs/api-reference.md)
